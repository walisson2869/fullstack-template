package ws

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client couples a single WebSocket connection to the Hub.
// Send is exported so hub_test.go can inject a client without a real conn.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	Send chan []byte
}

// NewClient allocates a Client and registers it with the hub.
// Call WritePump and ReadPump in separate goroutines after this returns.
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	c := &Client{
		hub:  hub,
		conn: conn,
		Send: make(chan []byte, 256),
	}
	hub.Register <- c
	return c
}

// ReadPump reads from the WebSocket and handles connection liveness.
// It unregisters the client and closes the connection when done.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait)) //nolint:errcheck
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket read error", "error", err)
			}
			break
		}
		// Server-to-client only for now; incoming messages are discarded.
	}
}

// WritePump writes queued messages to the WebSocket and sends periodic pings.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{}) //nolint:errcheck
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg) //nolint:errcheck
			// Drain any queued messages into the same WebSocket frame.
			for i := len(c.Send); i > 0; i-- {
				w.Write([]byte{'\n'}) //nolint:errcheck
				w.Write(<-c.Send)     //nolint:errcheck
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
