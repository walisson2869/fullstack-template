package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	ws "backend/internal/infrastructure/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// TODO: restrict to known origins in production.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WsHandler godoc
//
//	@Summary		Open a WebSocket connection
//	@Description	Upgrades HTTP to WebSocket. Pass a Firebase ID token as `?token=<token>`. Returns 401 when the token is missing or invalid.
//	@Tags			websocket
//	@Produce		json
//	@Param			token	query		string	true	"Firebase ID token"
//	@Success		101		{string}	string	"Switching Protocols"
//	@Failure		401		{object}	map[string]string
//	@Router			/ws [get]
func (h *Handler) WsHandler(c *gin.Context) {
	if h.verifier != nil {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		if _, err := h.verifier.VerifyIDToken(c.Request.Context(), token); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Upgrader already wrote the error response.
		return
	}

	client := ws.NewClient(h.hub, conn)
	go client.WritePump()
	go client.ReadPump()
}
