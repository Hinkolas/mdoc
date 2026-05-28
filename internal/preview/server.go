// Package preview runs the live-preview HTTP server: serves the SPA, pushes
// freshly rendered HTML over WebSocket on file changes, and proxies an HTTP
// endpoint that triggers the print pipeline.
package preview

import (
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

	mu      sync.RWMutex
	conn    *websocket.Conn
	send    chan []byte
	httpSrv *http.Server
	port    int
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
	r.HandleFunc("/ws", s.handleWebSocket)
	r.HandleFunc("/print", s.handlePrint).Methods("POST")
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

// PushRender re-reads the document, runs the pipeline, and pushes the
// resulting HTML over the WebSocket to the connected client (if any).
func (s *Server) PushRender() error {
	doc, err := document.Open(s.docPath)
	if err != nil {
		return err
	}
	thm, err := theme.Resolve(doc.Config.Theme, doc.Dir)
	if err != nil {
		return err
	}
	html, _, err := render.RenderThemed(doc, thm, render.Options{
		VendorBase: "/_/vendor",
		BaseHref:   s.URL(),
		Version:    s.version,
	})
	if err != nil {
		return err
	}
	return s.sendMessage(message{Event: "render", Data: renderPayload{HTML: html}})
}

// CurrentThemePath returns the resolved theme path of the current document.
// Used by callers to wire up a file watcher on it.
func (s *Server) CurrentThemePath() (string, error) {
	doc, err := document.Open(s.docPath)
	if err != nil {
		return "", err
	}
	thm, err := theme.Resolve(doc.Config.Theme, doc.Dir)
	if err != nil {
		return "", err
	}
	return thm.Path, nil
}

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

func (s *Server) handlePrint(w http.ResponseWriter, _ *http.Request) {
	doc, err := document.Open(s.docPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	thm, err := theme.Resolve(doc.Config.Theme, doc.Dir)
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

// serveDocAssets exposes files inside the document's directory at /assets/*
// so the preview can resolve relative <img>/<a> references. Path traversal
// is blocked: the resolved absolute path must stay inside docDir.
func (s *Server) serveDocAssets(w http.ResponseWriter, r *http.Request) {
	docDir := filepath.Dir(s.docPath)
	rel := strings.TrimPrefix(r.URL.Path, "/assets/")
	full := filepath.Join(docDir, filepath.FromSlash(rel))
	absFull, err := filepath.Abs(full)
	if err != nil || !strings.HasPrefix(absFull, docDir+string(filepath.Separator)) && absFull != docDir {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, absFull)
}
