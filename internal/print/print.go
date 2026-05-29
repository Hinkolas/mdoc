// Package print runs the one-shot pipeline: parse -> render -> headless
// Chromium -> wait for paged.js -> page.PDF -> write to disk.
package print

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/proto"

	"github.com/hinkolas/mdoc/internal/assets"
	"github.com/hinkolas/mdoc/internal/browser"
	"github.com/hinkolas/mdoc/internal/document"
	"github.com/hinkolas/mdoc/internal/render"
	"github.com/hinkolas/mdoc/internal/theme"
)

// Options configures Print.
type Options struct {
	// OutputPath is the PDF file to write. Empty means write next to the
	// source document as <basename>.pdf.
	OutputPath string
	// WriteHTML, if true, also writes the rendered HTML alongside the PDF
	// for debugging.
	WriteHTML bool
	// Version is propagated into the render System.Version field.
	Version string
}

// ResolveOutputPath returns the absolute path Print will write to: the
// explicit outputPath when given, otherwise <source-basename>.pdf next to
// the document. Exposed so callers can check for an existing file before
// committing to a render.
func ResolveOutputPath(doc *document.Document, outputPath string) (string, error) {
	if outputPath == "" {
		base := filepath.Base(doc.Path)
		ext := filepath.Ext(base)
		outputPath = filepath.Join(filepath.Dir(doc.Path), base[:len(base)-len(ext)]+".pdf")
	}
	return filepath.Abs(outputPath)
}

// Print renders a document to PDF and writes it to disk. Returns the
// absolute path of the resulting PDF.
func Print(doc *document.Document, thm *theme.Theme, opts Options) (string, error) {
	absOut, err := ResolveOutputPath(doc, opts.OutputPath)
	if err != nil {
		return "", fmt.Errorf("resolve output path: %w", err)
	}

	srv, err := startPrintServer(doc.Dir)
	if err != nil {
		return "", err
	}
	defer srv.shutdown()

	html, err := render.Render(doc, thm, render.Options{
		VendorBase: srv.url + "/_/vendor",
		BaseHref:   srv.url + "/",
		Version:    opts.Version,
	})
	if err != nil {
		return "", err
	}
	srv.setHTML(html)

	if opts.WriteHTML {
		debug := absOut[:len(absOut)-len(filepath.Ext(absOut))] + ".html"
		if err := os.WriteFile(debug, []byte(html), 0o644); err != nil {
			return "", fmt.Errorf("write debug html: %w", err)
		}
	}

	pdf, err := renderPDF(srv.url + "/")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(absOut, pdf, 0o644); err != nil {
		return "", fmt.Errorf("write pdf: %w", err)
	}
	return absOut, nil
}

// renderPDF launches a headless browser, navigates to the prepared URL,
// waits for paged.js to finish pagination, then captures the PDF.
func renderPDF(url string) ([]byte, error) {
	br, err := browser.Headless()
	if err != nil {
		return nil, err
	}
	defer br.Close()

	page := br.Page()

	if err := page.Navigate(url); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}
	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("wait load: %w", err)
	}

	// shell.html exposes window.__mdocPagedDone as a Promise that resolves
	// once Paged.Previewer.preview() finishes. Waiting on it here is what
	// guarantees the .pagedjs_page elements are in the DOM before PDF
	// capture. 60s is a generous ceiling for very large documents.
	timedPage := page.Timeout(60 * time.Second)
	if _, err := timedPage.Eval(`async () => { await window.__mdocPagedDone; return true; }`); err != nil {
		return nil, fmt.Errorf("wait for paged.js: %w", err)
	}

	// Paged.js has already laid out each printable page into a .pagedjs_page
	// element with its own @page-derived size and margins. Tell Chromium to
	// honor those CSS page sizes and not add any of its own margins on top.
	printBg := true
	preferCSS := true
	zero := 0.0
	stream, err := page.PDF(&proto.PagePrintToPDF{
		PrintBackground:   printBg,
		PreferCSSPageSize: preferCSS,
		MarginTop:         &zero,
		MarginBottom:      &zero,
		MarginLeft:        &zero,
		MarginRight:       &zero,
	})
	if err != nil {
		return nil, fmt.Errorf("page.PDF: %w", err)
	}
	defer stream.Close()
	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("read pdf stream: %w", err)
	}
	if len(data) == 0 {
		return nil, errors.New("empty PDF from page.PDF")
	}
	return data, nil
}

// printServer is a minimal HTTP server used only for the duration of a
// single print invocation. It serves three things:
//
//   - "/"            : the rendered document HTML
//   - "/_/vendor/*"  : embedded paged.js + KaTeX
//   - "/<rest>"      : files inside the source document's directory, so
//                      relative <img> / <a href> URLs resolve.
//
// Using HTTP instead of file:// lets us sidestep Chromium's same-origin
// restrictions on file:// resources, which silently break vendor loading
// when the rendered HTML and the vendor tree live in different directories.
type printServer struct {
	url     string
	docDir  string
	html    string
	httpSrv *http.Server
	ln      net.Listener
}

func startPrintServer(docDir string) (*printServer, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	ps := &printServer{
		url:    fmt.Sprintf("http://%s", ln.Addr().String()),
		docDir: docDir,
		ln:     ln,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", ps.handleAny)
	mux.Handle("/_/vendor/", http.StripPrefix("/_/vendor/", http.FileServer(http.FS(assets.Vendor()))))
	ps.httpSrv = &http.Server{Handler: mux}
	go ps.httpSrv.Serve(ln)
	return ps, nil
}

func (ps *printServer) setHTML(html string) { ps.html = html }

func (ps *printServer) shutdown() {
	if ps.httpSrv != nil {
		_ = ps.httpSrv.Close()
	}
}

func (ps *printServer) handleAny(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/index.html" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, ps.html)
		return
	}
	// Anything else is a relative reference from the document body; serve
	// it out of the document's directory with a path-traversal guard.
	rel := strings.TrimPrefix(r.URL.Path, "/")
	full := filepath.Join(ps.docDir, filepath.FromSlash(rel))
	abs, err := filepath.Abs(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !strings.HasPrefix(abs, ps.docDir+string(filepath.Separator)) && abs != ps.docDir {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, abs)
}
