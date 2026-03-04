package api

import (
	"log"
	"net/http"

	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
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
	cookie, err := r.Cookie("access_token")
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHORIZED", Message: "missing access token"})
		return
	}

	claims, err := auth.ValidateAccessToken(cookie.Value, wh.JWTSecret)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHORIZED", Message: "invalid or expired token"})
		return
	}

	if claims.Role != "student" {
		writeJSON(w, http.StatusForbidden, APIError{Code: "FORBIDDEN", Message: "only students can join games"})
		return
	}

	if claims.ProfileID == "" {
		writeJSON(w, http.StatusForbidden, APIError{Code: "FORBIDDEN", Message: "student profile required"})
		return
	}

	// 2. Upgrade to WebSocket.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws: upgrade error: %v", err)
		return // Upgrade already wrote an HTTP error.
	}

	// 3. Create client and register with hub.
	client := ws.NewClient(wh.Hub, conn, claims.ProfileID)
	wh.Hub.Register(client)

	// 4. Start read/write pumps.
	go client.WritePump()
	go client.ReadPump()
}
