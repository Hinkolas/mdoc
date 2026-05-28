package preview

import (
	"encoding/json"
	"fmt"
)

// message is the wire format for WebSocket frames in both directions.
type message struct {
	Event string `json:"event"`
	Data  any    `json:"data,omitempty"`
}

// renderPayload is the body of a server -> client `render` event.
type renderPayload struct {
	HTML string `json:"html"`
}

// sendMessage marshals and queues a frame for the connected client. Returns
// an error if no client is connected or the send buffer is full.
func (s *Server) sendMessage(m message) error {
	s.mu.RLock()
	send := s.send
	s.mu.RUnlock()
	if send == nil {
		return fmt.Errorf("no client connected")
	}
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	select {
	case send <- data:
		return nil
	default:
		return fmt.Errorf("send buffer full")
	}
}
