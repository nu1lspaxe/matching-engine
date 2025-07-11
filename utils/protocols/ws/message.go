package ws

import (
	"encoding/json"
	"matching-engine/utils/logger"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (s *Server) dispatch(c *Client, msg []byte) {
	var m Message
	if err := json.Unmarshal(msg, &m); err != nil {
		logger.Error("Failed to unmarshal message", err.Error())
		return
	}

	s.RLock()
	handler, exists := s.handlers[m.Type]
	s.RUnlock()

	if !exists {
		logger.Error("No handler found for message type", m.Type)
		return
	}

	handler(c, m.Data)
}
