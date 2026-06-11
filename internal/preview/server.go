// Package preview runs the live-preview HTTP server: serves the SPA, an
// iframe-hosted paged.js view of the document, pushes a reload signal over
// WebSocket on file changes, and exposes /print to trigger PDF generation.
package preview

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/hinkolas/mdoc/internal/assets"
	"github.com/hinkolas/mdoc/internal/document"
	"github.com/hinkolas/mdoc/internal/print"
	"github.com/hinkolas/mdoc/internal/render"
	"github.com/hinkolas/mdoc/internal/theme"
)

// Server is a single-client live-preview server bound to one source document.
type Server struct {
	docPath string
	version string

	mu           sync.RWMutex
	conn         *websocket.Conn
	send         chan []byte
	httpSrv      *http.Server
	port         int
	themeWarning string // last non-fatal theme diagnostic, surfaced via /status
}

// New returns a server for the given document. The document is re-read from
// disk on every render so edits in any external editor are picked up.
func New(docPath, version string) *Server {
	abs, _ := filepath.Abs(docPath)
	return &Server{docPath: abs, version: version}
}

// Port returns the port the server is listening on. Only valid after Start.
func (s *Server) Port() int { return s.port }

// URL returns the preview origin (no path).
func (s *Server) URL() string { return fmt.Sprintf("http://127.0.0.1:%d/", s.port) }

// DocPath is the absolute path of the document this server is serving.
func (s *Server) DocPath() string { return s.docPath }

// Start binds the server (port 0 = auto-pick) and serves in the background.
func (s *Server) Start(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.port = ln.Addr().(*net.TCPAddr).Port

	r := mux.NewRouter()
	r.HandleFunc("/", s.handleIndex).Methods("GET")
	r.HandleFunc("/preview", s.handlePreview).Methods("GET")
	r.HandleFunc("/preview/body", s.handlePreviewBody).Methods("GET")
	r.HandleFunc("/ws", s.handleWebSocket)
	r.HandleFunc("/print", s.handlePrint).Methods("POST")
	r.HandleFunc("/status", s.handleStatus).Methods("GET")
	r.HandleFunc("/open-url", s.handleOpenURL).Methods("POST")
	r.PathPrefix("/_/ui/").Handler(http.StripPrefix("/_/ui/", http.FileServer(http.FS(assets.UI()))))
	r.PathPrefix("/_/vendor/").Handler(http.StripPrefix("/_/vendor/", http.FileServer(http.FS(assets.Vendor()))))
	r.PathPrefix("/assets/").HandlerFunc(s.serveDocAssets)

	s.httpSrv = &http.Server{Handler: r}
	go func() {
		if err := s.httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("preview server: %v", err)
		}
	}()
	return nil
}

// Shutdown stops the HTTP server.
func (s *Server) Shutdown() error {
	if s.httpSrv == nil {
		return nil
	}
	return s.httpSrv.Close()
}

// PushReload notifies the connected client that the underlying document or
// theme changed and the iframe should be reloaded.
func (s *Server) PushReload() error {
	return s.sendMessage(message{Event: "reload"})
}

// resolve re-reads the document from disk and resolves its theme, recording
// the (non-fatal) theme diagnostic so /status can report it. A non-nil error
// means the document itself couldn't be read or parsed — that's fatal for the
// request; a missing/broken theme is not, since Resolve falls back to the
// built-in default and reports it via the warning instead.
func (s *Server) resolve() (*document.Document, *theme.Theme, error) {
	doc, err := document.Open(s.docPath)
	if err != nil {
		return nil, nil, err
	}
	thm, warn := theme.Resolve(doc.Config.Theme, doc.Dir)
	s.setThemeWarning(warn)
	return doc, thm, nil
}

func (s *Server) setThemeWarning(err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	s.mu.Lock()
	s.themeWarning = msg
	s.mu.Unlock()
}

// CurrentTheme re-resolves the current document's theme and reports what the
// watcher and live logger need: the resolved theme file path ("" for a
// built-in, so the file watch is dropped), the current non-fatal theme
// warning ("" when the theme is fine), and docErr — set when the document
// itself can't be read or parsed (which is fatal for rendering, unlike a
// theme problem).
func (s *Server) CurrentTheme() (path, warning, docErr string) {
	_, thm, err := s.resolve()
	if err != nil {
		return "", "", err.Error()
	}
	s.mu.RLock()
	warning = s.themeWarning
	s.mu.RUnlock()
	return thm.Path, warning, ""
}

// CurrentIncludes re-reads the document and reports the absolute paths of the
// files it pulls in via `:::include`, so the watcher can follow them. A document
// that can't be read (e.g. mid-edit) yields nil — the render path reports that
// error separately; here we just leave the include watch set unchanged-worthy.
func (s *Server) CurrentIncludes() []string {
	doc, err := document.Open(s.docPath)
	if err != nil {
		return nil
	}
	return doc.Includes
}

// handleStatus reports the latest non-fatal preview diagnostic — currently
// just the theme warning — so the SPA can show it without parsing the
// rendered HTML.
func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	warn := s.themeWarning
	s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]string{"warning": warn})
}

// handleIndex serves the preview SPA chrome (header + iframe).
func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	doc, err := document.Open(s.docPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmplBytes, err := assets.UIBytes("preview.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t, err := template.New("preview").Parse(string(tmplBytes))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = t.Execute(w, map[string]any{"Title": doc.Config.Title})
}

// handlePreview returns the document re-rendered through the same shell.html
// path the print pipeline uses, so the iframe's paged.js layout matches the
// PDF exactly.
func (s *Server) handlePreview(w http.ResponseWriter, _ *http.Request) {
	doc, thm, err := s.resolve()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html, err := render.Render(doc, thm, render.Options{
		VendorBase: "/_/vendor",
		BaseHref:   "/assets/",
		Version:    s.version,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Make sure browsers always re-fetch on iframe reload — otherwise edits
	// to the source can be masked by the disk cache.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, html)
}

// handlePreviewBody returns just the themed HTML (no shell wrap), used by
// the iframe to re-paginate in place without a full reload. Theme @page
// rules and other styles ride along inside <style> tags; the in-iframe
// paginate function extracts them and feeds them to paged.js's Polisher.
func (s *Server) handlePreviewBody(w http.ResponseWriter, _ *http.Request) {
	doc, thm, err := s.resolve()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html, _, err := render.RenderThemed(doc, thm, render.Options{
		VendorBase: "/_/vendor",
		BaseHref:   "/assets/",
		Version:    s.version,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, html)
}

func (s *Server) handlePrint(w http.ResponseWriter, _ *http.Request) {
	doc, thm, err := s.resolve()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmp, err := os.CreateTemp("", "mdoc-print-*.pdf")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	if _, err := print.Print(doc, thm, print.Options{OutputPath: tmp.Name(), Version: s.version}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f, err := os.Open(tmp.Name())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	name := strings.TrimSuffix(filepath.Base(doc.Path), filepath.Ext(doc.Path)) + ".pdf"
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	_, _ = io.Copy(w, f)
}

// handleOpenURL receives a URL clicked inside the preview iframe and
// hands it off to the operating system's default browser. Without this
// hop the iframe would navigate itself to the external page, leaving
// the user stuck inside a chromeless window with no back button.
func (s *Server) handleOpenURL(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := openInDefaultBrowser(body.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// serveDocAssets exposes files inside the document's directory at /assets/*
// so the preview can resolve relative <img>/<a> references. Path traversal
// is blocked: the resolved absolute path must stay inside docDir.
func (s *Server) serveDocAssets(w http.ResponseWriter, r *http.Request) {
	docDir := filepath.Dir(s.docPath)
	rel := strings.TrimPrefix(r.URL.Path, "/assets/")
	full := filepath.Join(docDir, filepath.FromSlash(rel))
	absFull, err := filepath.Abs(full)
	if err != nil || (!strings.HasPrefix(absFull, docDir+string(filepath.Separator)) && absFull != docDir) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, absFull)
}
