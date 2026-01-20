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

// Message represents a WebSocket message with a type and optional payload
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// CommandHandler handles incoming commands from the client
type CommandHandler func(payload json.RawMessage) error

// PreviewServer manages the preview HTTP server and WebSocket connection
type PreviewServer struct {
	mu       sync.RWMutex
	conn     *websocket.Conn
	send     chan []byte
	handlers map[string]CommandHandler

	// Callbacks for external integration
	OnSave func() error
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local development
	},
}

// NewPreviewServer creates a new preview server instance
func NewPreviewServer() *PreviewServer {
	s := &PreviewServer{
		handlers: make(map[string]CommandHandler),
	}
	s.RegisterHandler("save", s.handleSave)
	s.RegisterHandler("save_at", s.handleSaveAt)
	return s
}

// RegisterHandler registers a handler for a specific command
func (s *PreviewServer) RegisterHandler(command string, handler CommandHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[command] = handler
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

// SendPayload sends a message with a typed payload
func (s *PreviewServer) SendPayload(msgType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	return s.Send(Message{Type: msgType, Payload: data})
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
