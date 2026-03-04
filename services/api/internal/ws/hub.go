package ws

import (
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Hub manages all WebSocket clients and game rooms.
type Hub struct {
	mu sync.RWMutex

	// rooms is the set of active rooms keyed by room ID.
	rooms map[string]*Room

	// clients is the set of all connected clients keyed by profile ID.
	clients map[string]*Client

	// profileRoom maps a profile ID to its active room ID (one room per profile).
	profileRoom map[string]string

	// waitingRooms maps game_type → room ID for rooms in WAITING state.
	waitingRooms map[string]string

	// register / unregister channels for concurrency-safe client management.
	register   chan *Client
	unregister chan *Client

	db *pgxpool.Pool
}

// NewHub creates a Hub ready to Run().
func NewHub(db *pgxpool.Pool) *Hub {
	return &Hub{
		rooms:        make(map[string]*Room),
		clients:      make(map[string]*Client),
		profileRoom:  make(map[string]string),
		waitingRooms: make(map[string]string),
		register:     make(chan *Client, 64),
		unregister:   make(chan *Client, 64),
		db:           db,
	}
}

// Run processes register / unregister events. Must be called in its own goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			// If this profile already has an active connection, close the old one.
			if old, ok := h.clients[c.profileID]; ok {
				old.conn.Close()
			}
			h.clients[c.profileID] = c
			h.mu.Unlock()
			log.Printf("ws: registered profile=%s", c.profileID)

		case c := <-h.unregister:
			h.mu.Lock()
			// Only remove if it's the same pointer (not a newer reconnect).
			if cur, ok := h.clients[c.profileID]; ok && cur == c {
				delete(h.clients, c.profileID)
			}
			roomID := c.RoomID()
			h.mu.Unlock()

			if roomID != "" {
				h.mu.RLock()
				room, ok := h.rooms[roomID]
				h.mu.RUnlock()
				if ok {
					room.RemovePlayer(c)
				}
			}
			log.Printf("ws: unregistered profile=%s", c.profileID)
		}
	}
}

// Register queues a client for registration.
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// handleJoinRoom places a client into matchmaking or an existing waiting room.
func (h *Hub) handleJoinRoom(c *Client, msg JoinRoom) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Enforce one active room per profile.
	if existingRoomID, ok := h.profileRoom[c.profileID]; ok {
		if room, exists := h.rooms[existingRoomID]; exists && room.State() != StateEnded {
			c.SendJSON(ErrorMsg{Type: "error", Code: "ALREADY_IN_ROOM", Message: "you are already in a room"})
			return
		}
		// Stale entry — clean up.
		delete(h.profileRoom, c.profileID)
	}

	// Check for a waiting room of the same game type.
	if waitingRoomID, ok := h.waitingRooms[msg.GameType]; ok {
		room, exists := h.rooms[waitingRoomID]
		if exists && room.State() == StateWaiting {
			// Seat as player 2.
			if err := room.AddPlayer(c); err != nil {
				c.SendJSON(ErrorMsg{Type: "error", Code: "JOIN_FAILED", Message: err.Error()})
				return
			}
			h.profileRoom[c.profileID] = waitingRoomID
			delete(h.waitingRooms, msg.GameType)
			return
		}
		// Room was stale; clean up.
		delete(h.waitingRooms, msg.GameType)
	}

	// Create a new waiting room.
	roomID := uuid.New().String()
	room := NewRoom(roomID, msg.GameType, h, h.db, c)
	h.rooms[roomID] = room
	h.profileRoom[c.profileID] = roomID
	h.waitingRooms[msg.GameType] = roomID
	log.Printf("ws: created room=%s type=%s player1=%s", roomID, msg.GameType, c.profileID)
}

// handlePlayerAction routes an action to the client's room.
func (h *Hub) handlePlayerAction(c *Client, msg PlayerAction) {
	roomID := c.RoomID()
	if roomID == "" {
		c.SendJSON(ActionRejected{Type: "action_rejected", Reason: "not in a room"})
		return
	}

	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()
	if !ok {
		c.SendJSON(ActionRejected{Type: "action_rejected", Reason: "room not found"})
		return
	}

	room.HandleAction(c, msg)
}

// handleLeaveRoom removes a client from their room intentionally.
func (h *Hub) handleLeaveRoom(c *Client) {
	roomID := c.RoomID()
	if roomID == "" {
		return
	}

	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()
	if !ok {
		return
	}

	room.RemovePlayer(c)
}

// removeRoom cleans up hub maps for a terminated room.
func (h *Hub) removeRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[roomID]
	if !ok {
		return
	}

	// Clean profile → room index.
	if room.p1ID != "" {
		if h.profileRoom[room.p1ID] == roomID {
			delete(h.profileRoom, room.p1ID)
		}
	}
	if room.p2ID != "" {
		if h.profileRoom[room.p2ID] == roomID {
			delete(h.profileRoom, room.p2ID)
		}
	}

	// Clean waiting room index.
	if wID, wok := h.waitingRooms[room.GameType]; wok && wID == roomID {
		delete(h.waitingRooms, room.GameType)
	}

	delete(h.rooms, roomID)
	log.Printf("ws: removed room=%s", roomID)
}

// GetRoom returns a room by ID (used by reconnect logic).
func (h *Hub) GetRoom(roomID string) (*Room, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	room, ok := h.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room %s not found", roomID)
	}
	return room, nil
}

// RoomCount returns the current number of active rooms (useful for tests / health).
func (h *Hub) RoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
