package preview

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

//go:embed preview.html
var previewHTML string

// Message represents a WebSocket message with an event name and optional data
type Message struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data,omitempty"`
}

// EventHandler handles incoming events from the client
type EventHandler func(data json.RawMessage) error

// PreviewServer manages the preview HTTP server and WebSocket connection
type PreviewServer struct {
	mu       sync.RWMutex
	conn     *websocket.Conn
	send     chan []byte
	handlers map[string]EventHandler

	// Document path for rendering
	DocumentPath string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local development
	},
}

// NewPreviewServer creates a new preview server instance for the given document
func NewPreviewServer(documentPath string) *PreviewServer {
	s := &PreviewServer{
		handlers:     make(map[string]EventHandler),
		DocumentPath: documentPath,
	}
	// Register built-in event handlers
	s.RegisterHandler("render_request", s.handleRenderRequest)
	return s
}

// RegisterHandler registers a handler for a specific event
func (s *PreviewServer) RegisterHandler(event string, handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[event] = handler
}

// Send sends a message to the connected client
func (s *PreviewServer) Send(msg Message) error {
	s.mu.RLock()
	send := s.send
	s.mu.RUnlock()

	if send == nil {
		return fmt.Errorf("no client connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	select {
	case send <- data:
		return nil
	default:
		return fmt.Errorf("send buffer full")
	}
}

// SendEvent sends a message with no data (event only)
func (s *PreviewServer) SendEvent(event string) error {
	return s.Send(Message{Event: event})
}

// SendEventWithData sends a message with event and data payload
func (s *PreviewServer) SendEventWithData(event string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	return s.Send(Message{Event: event, Data: jsonData})
}

// IsConnected returns true if a client is connected
func (s *PreviewServer) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.conn != nil
}

// Start starts the preview server on the specified port
func (s *PreviewServer) Start(port int) error {
	router := mux.NewRouter()

	// Serve the preview HTML page
	router.HandleFunc("/preview", s.handlePreviewPage).Methods("GET")

	// WebSocket endpoint
	router.HandleFunc("/ws", s.handleWebSocket)

	// Root redirect to preview
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/preview", http.StatusTemporaryRedirect)
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Preview server starting at http://localhost%s/preview", addr)

	return http.ListenAndServe(addr, router)
}

// handlePreviewPage serves the embedded preview HTML
func (s *PreviewServer) handlePreviewPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(previewHTML))
}
