package chat

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Client struct {
	Conn         *websocket.Conn
	UserID       int
	Name         string
	Mu           sync.Mutex
	IsAuthorized bool
	Send         chan []byte
}

func NewClient(conn *websocket.Conn, auth bool, userID int, username string) *Client {
	return &Client{
		UserID:       userID,
		Conn:         conn,
		Name:         username,
		IsAuthorized: auth,
		Send:         make(chan []byte, 256),
	}
}

func (c *Client) WriteJSON(v any) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	return c.Conn.WriteJSON(v)
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.Conn.Close()
		ticker.Stop()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
