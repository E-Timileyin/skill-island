package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// HeartbeatInterval is how often the client should ping.
	HeartbeatInterval = 5 * time.Second
	// StaleThreshold is how long to wait before considering a conn stale.
	StaleThreshold = 15 * time.Second
	// writeWait is the max time to wait for a write to complete.
	writeWait = 10 * time.Second
	// maxMessageSize is the max inbound message size (64 KB).
	maxMessageSize = 64 * 1024
)

// Client represents a single WebSocket connection from a player.
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	profileID string
	roomID    string // empty until joined
	send      chan []byte

	mu       sync.Mutex
	lastPing time.Time
}

// NewClient creates a Client bound to the given hub and connection.
func NewClient(hub *Hub, conn *websocket.Conn, profileID string) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		profileID: profileID,
		send:      make(chan []byte, 64),
		lastPing:  time.Now(),
	}
}

// ProfileID returns the client's profile identifier.
func (c *Client) ProfileID() string { return c.profileID }

// RoomID returns the client's current room (empty if unmatched).
func (c *Client) RoomID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.roomID
}

// SetRoomID updates the room assignment.
func (c *Client) SetRoomID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.roomID = id
}

// LastPing returns when the client last pinged.
func (c *Client) LastPing() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastPing
}

// touchPing refreshes the last-ping timestamp.
func (c *Client) touchPing() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastPing = time.Now()
}

// IsStale reports whether the client has not pinged within StaleThreshold.
func (c *Client) IsStale() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Since(c.lastPing) > StaleThreshold
}

// Send enqueues a message for writing; drops silently if the buffer is full.
func (c *Client) Send(msg []byte) {
	select {
	case c.send <- msg:
	default:
		log.Printf("ws: send buffer full for profile %s, dropping message", c.profileID)
	}
}

// SendJSON marshals v and enqueues the bytes.
func (c *Client) SendJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("ws: marshal error for profile %s: %v", c.profileID, err)
		return
	}
	c.Send(data)
}

// Close cleanly shuts down the client connection.
func (c *Client) Close() {
	close(c.send)
	c.conn.Close()
}

// ReadPump reads messages from the WebSocket and dispatches them to the hub.
// It must be called in its own goroutine. When it returns, the hub unregisters
// the client and the connection is closed.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(StaleThreshold))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(StaleThreshold))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("ws: read error profile=%s: %v", c.profileID, err)
			}
			return
		}

		// Peek at the type field.
		var env struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &env); err != nil {
			c.SendJSON(ErrorMsg{Type: "error", Code: "BAD_MESSAGE", Message: "invalid JSON"})
			continue
		}

		switch env.Type {
		case "heartbeat_ping":
			c.touchPing()
			c.conn.SetReadDeadline(time.Now().Add(StaleThreshold))
			c.SendJSON(HeartbeatPong{
				Type:       "heartbeat_pong",
				Timestamp:  time.Now().UnixMilli(),
				ServerTime: time.Now().UnixMilli(),
			})

		case "join_room":
			var msg JoinRoom
			if err := json.Unmarshal(raw, &msg); err != nil {
				c.SendJSON(ErrorMsg{Type: "error", Code: "BAD_MESSAGE", Message: "invalid join_room"})
				continue
			}
			c.hub.handleJoinRoom(c, msg)

		case "player_action":
			var msg PlayerAction
			if err := json.Unmarshal(raw, &msg); err != nil {
				c.SendJSON(ErrorMsg{Type: "error", Code: "BAD_MESSAGE", Message: "invalid player_action"})
				continue
			}
			c.hub.handlePlayerAction(c, msg)

		case "leave_room":
			c.hub.handleLeaveRoom(c)

		default:
			c.SendJSON(ErrorMsg{Type: "error", Code: "UNKNOWN_TYPE", Message: "unknown message type"})
		}
	}
}

// WritePump drains the send channel and writes messages to the WebSocket.
// It must be called in its own goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel closed.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
