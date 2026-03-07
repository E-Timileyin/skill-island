package handlers

import (
	"net/http"

	"github.com/E-Timileyin/skill-island/services/api/internal/ws"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: restrict to AllowedOrigins in production.
		return true
	},
}

// WSHandler holds dependencies for the WebSocket upgrade endpoint.
type WSHandler struct {
	Hub       *ws.Hub
	JWTSecret string
}

// ServeWS validates the JWT from the cookie BEFORE upgrading to WebSocket.
// On failure it responds with HTTP 401 (not a WS error frame).
func (wh *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// 1. Extract and validate JWT from cookie — before upgrade.
	_, err := r.Cookie("access_token")
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHORIZED", Message: "missing or invalid access token"})
		return
	}
	// ...existing code...
}
