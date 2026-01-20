package preview

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// handleWebSocket handles WebSocket connection upgrades
func (s *PreviewServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Close existing connection if any
	s.mu.Lock()
	if s.conn != nil {
		s.conn.Close()
		close(s.send)
		log.Printf("Closed existing connection")
	}
	s.conn = conn
	s.send = make(chan []byte, 256)
	send := s.send
	s.mu.Unlock()

	log.Printf("Client connected")

	// Start read and write pumps
	go s.writePump(conn, send)
	s.readPump(conn) // Blocks until connection closes
}

// readPump reads messages from the WebSocket connection
func (s *PreviewServer) readPump(conn *websocket.Conn) {
	defer func() {
		s.mu.Lock()
		if s.conn == conn {
			s.conn = nil
			close(s.send)
			s.send = nil
		}
		s.mu.Unlock()
		conn.Close()
		log.Printf("Client disconnected")
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Invalid JSON message: %v", err)
			continue
		}

		s.mu.RLock()
		handler, exists := s.handlers[msg.Event]
		s.mu.RUnlock()

		if exists {
			if err := handler(msg.Data); err != nil {
				log.Printf("Handler error for '%s': %v", msg.Event, err)
			}
		} else {
			log.Printf("Unknown event: %s", msg.Event)
		}
	}
}

// writePump writes messages to the WebSocket connection
func (s *PreviewServer) writePump(conn *websocket.Conn, send <-chan []byte) {
	defer conn.Close()

	for message := range send {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}
