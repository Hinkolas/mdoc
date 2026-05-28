package preview

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Local-only server; any origin is fine.
	CheckOrigin: func(*http.Request) bool { return true },
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}

	s.mu.Lock()
	if s.conn != nil {
		_ = s.conn.Close()
		close(s.send)
	}
	s.conn = conn
	s.send = make(chan []byte, 32)
	send := s.send
	s.mu.Unlock()

	// Push an initial render as soon as the client connects so the window
	// isn't blank while we wait for the first file change.
	go func() {
		if err := s.PushRender(); err != nil {
			log.Printf("initial render: %v", err)
		}
	}()

	go s.writePump(conn, send)
	s.readPump(conn)
}

func (s *Server) readPump(conn *websocket.Conn) {
	defer func() {
		s.mu.Lock()
		if s.conn == conn {
			s.conn = nil
			close(s.send)
			s.send = nil
		}
		s.mu.Unlock()
		_ = conn.Close()
	}()
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var m message
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		if m.Event == "reload" {
			if err := s.PushRender(); err != nil {
				log.Printf("reload render: %v", err)
			}
		}
	}
}

func (s *Server) writePump(conn *websocket.Conn, send <-chan []byte) {
	for data := range send {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}
