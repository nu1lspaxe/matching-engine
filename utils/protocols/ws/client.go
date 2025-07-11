package ws

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	conn      *websocket.Conn
	send      chan []byte
	onClose   func(*Client)
	onMessage func(*Client, []byte)
}

func (c *Client) Read() {
	defer func() {
		if c.onClose != nil {
			c.onClose(c)
		}
		c.conn.Close()
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			c.conn.WriteMessage(websocket.CloseMessage, []byte(
				NewError(ErrReadMessage, err.Error()).Error(),
			))
			return
		}

		if c.onMessage != nil {
			c.onMessage(c, msg)
		}
	}
}

func (c *Client) Write() {
	defer c.conn.Close()

	for {
		msg, ok := <-c.send
		if !ok {
			return
		}
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			c.conn.WriteMessage(websocket.CloseMessage, []byte(
				NewError(ErrWriteMessage, err.Error()).Error(),
			))
			return
		}
	}
}
