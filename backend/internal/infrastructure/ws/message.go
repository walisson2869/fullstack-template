package ws

import "encoding/json"

// Envelope is the typed wire format for all WebSocket messages.
type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
