package ws

import (
	"context"
	"encoding/json"
)

// Hub maintains the set of active WebSocket clients and broadcasts messages to them.
// All mutations are serialised through the Run goroutine — no locking needed on the
// clients map itself.
type Hub struct {
	clients    map[*Client]struct{}
	broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

// NewHub allocates a Hub with buffered channels.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run processes register, unregister, and broadcast events until ctx is cancelled.
// Call this in its own goroutine.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for c := range h.clients {
				close(c.Send)
			}
			return
		case c := <-h.Register:
			h.clients[c] = struct{}{}
		case c := <-h.Unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.Send)
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.Send <- msg:
				default:
					// Slow client: drop and disconnect.
					close(c.Send)
					delete(h.clients, c)
				}
			}
		}
	}
}

// Publish marshals msgType + payload into an Envelope and queues it for broadcast.
// Safe to call from any goroutine (e.g. an Asynq worker or Redis Streams consumer).
func (h *Hub) Publish(msgType string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	env, err := json.Marshal(Envelope{Type: msgType, Payload: raw})
	if err != nil {
		return err
	}
	h.broadcast <- env
	return nil
}
