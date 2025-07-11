package ws

import (
	"encoding/json"
	"log"
)

type msgType struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (s *Server) dispatch(c *Client, msg []byte) {
	var mt msgType
	if err := json.Unmarshal(msg, &mt); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	s.RLock()
	handler, exists := s.handlers[mt.Type]
	s.RUnlock()

	if !exists {
		log.Printf("No handler found for message type: %s", mt.Type)
		return
	}

	handler(c, mt.Data)
}
