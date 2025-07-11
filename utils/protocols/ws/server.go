package ws

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type HandlerFunc func(c *Client, msg []byte)

type Server struct {
	sync.RWMutex
	upgrader websocket.Upgrader
	clients  map[*Client]bool
	handlers map[string]HandlerFunc
}

func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:  make(map[*Client]bool),
		handlers: make(map[string]HandlerFunc),
	}
}

func (s *Server) On(event string, handler HandlerFunc) {
	s.Lock()
	defer s.Unlock()
	s.handlers[event] = handler
}

func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		conn:      conn,
		send:      make(chan []byte),
		onClose:   s.handleClientClose,
		onMessage: s.handleClientMessage,
	}

	s.Lock()
	s.clients[client] = true
	s.Unlock()

	go client.Read()
	go client.Write()
}

func (s *Server) Broadcast(msg []byte) {
	s.RLock()
	defer s.RUnlock()
	for c := range s.clients {
		select {
		case c.send <- msg:
		default:
			s.removeClient(c)
		}
	}
}

func (s *Server) removeClient(client *Client) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.clients[client]; ok {
		delete(s.clients, client)
		close(client.send)
	}
}

func (s *Server) handleClientClose(client *Client) {
	s.removeClient(client)
}

func (s *Server) handleClientMessage(client *Client, msg []byte) {
	s.dispatch(client, msg)
}
