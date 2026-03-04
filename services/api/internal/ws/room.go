package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RoomState represents the lifecycle state of a room.
type RoomState int

const (
	StateWaiting RoomState = iota
	StateReady
	StatePlaying
	StatePaused
	StateEnded
)

func (s RoomState) String() string {
	switch s {
	case StateWaiting:
		return "WAITING"
	case StateReady:
		return "READY"
	case StatePlaying:
		return "PLAYING"
	case StatePaused:
		return "PAUSED"
	case StateEnded:
		return "ENDED"
	default:
		return "UNKNOWN"
	}
}

const (
	// TickInterval is the authoritative tick rate (~11 Hz).
	TickInterval = 90 * time.Millisecond
	// ReconnectWindow is how long a room stays paused waiting for reconnect.
	ReconnectWindow = 30 * time.Second
	// IdleWarnAt triggers a warning if no actions for this duration.
	IdleWarnAt = 90 * time.Second
	// IdleCloseAt closes the room if no actions for this duration.
	IdleCloseAt = 120 * time.Second
)

// Room is a 2-player game room that runs in its own goroutine.
type Room struct {
	ID       string
	GameType string

	mu      sync.RWMutex
	players [2]*Client
	state   RoomState
	tick    int64

	// Cached profile IDs so we can write room_sessions after players are nilled.
	p1ID string
	p2ID string

	ctx    context.Context
	cancel context.CancelFunc
	db     *pgxpool.Pool
	hub    *Hub

	lastAction time.Time
	startedAt  time.Time
}

// NewRoom creates a room with the first player already seated.
func NewRoom(id, gameType string, hub *Hub, db *pgxpool.Pool, first *Client) *Room {
	ctx, cancel := context.WithCancel(context.Background())
	now := time.Now()
	r := &Room{
		ID:         id,
		GameType:   gameType,
		state:      StateWaiting,
		ctx:        ctx,
		cancel:     cancel,
		db:         db,
		hub:        hub,
		lastAction: now,
		startedAt:  now,
	}
	r.players[0] = first
	r.p1ID = first.profileID
	first.SetRoomID(id)
	return r
}

// State returns the current lifecycle state.
func (r *Room) State() RoomState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

// PlayerCount returns how many players are connected.
func (r *Room) PlayerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, p := range r.players {
		if p != nil {
			n++
		}
	}
	return n
}

// AddPlayer seats the second player, transitions to READY, and starts the game loop.
func (r *Room) AddPlayer(c *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state != StateWaiting {
		return fmt.Errorf("room %s not in WAITING state (state=%s)", r.ID, r.state)
	}
	if r.players[1] != nil {
		return fmt.Errorf("room %s already full", r.ID)
	}

	r.players[1] = c
	r.p2ID = c.profileID
	c.SetRoomID(r.ID)
	r.state = StateReady

	// Notify both players.
	r.players[0].SendJSON(RoomReady{
		Type:           "room_ready",
		RoomID:         r.ID,
		PlayerRole:     "player_1",
		OpponentAvatar: 0, // TODO: fetch from profile
	})
	r.players[1].SendJSON(RoomReady{
		Type:           "room_ready",
		RoomID:         r.ID,
		PlayerRole:     "player_2",
		OpponentAvatar: 0, // TODO: fetch from profile
	})

	// Transition to PLAYING and start the tick loop.
	r.state = StatePlaying
	r.startedAt = time.Now()
	go r.run()

	return nil
}

// HandleAction processes a validated player action.
func (r *Room) HandleAction(c *Client, msg PlayerAction) {
	r.mu.Lock()
	if r.state != StatePlaying {
		r.mu.Unlock()
		c.SendJSON(ActionRejected{Type: "action_rejected", Reason: "room not in PLAYING state"})
		return
	}
	r.lastAction = time.Now()
	r.mu.Unlock()

	// Broadcast the action to both players (server is source of truth;
	// actual game-logic validation is done by the validator package).
	r.broadcast(msg)
}

// RemovePlayer handles a disconnect or intentional leave.
func (r *Room) RemovePlayer(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	slot := -1
	for i, p := range r.players {
		if p != nil && p.profileID == c.profileID {
			slot = i
			break
		}
	}
	if slot < 0 {
		return
	}

	r.players[slot] = nil
	c.SetRoomID("")

	switch r.state {
	case StateWaiting:
		// Only player left; just end.
		r.endLocked("player_left")

	case StateReady, StatePlaying:
		// Pause for reconnect window.
		r.state = StatePaused
		role := fmt.Sprintf("player_%d", slot+1)
		other := r.otherPlayerLocked(slot)
		if other != nil {
			other.SendJSON(PlayerDisconnected{
				Type:                   "player_disconnected",
				PlayerRole:             role,
				Reason:                 "player_left",
				ReconnectWindowSeconds: int(ReconnectWindow.Seconds()),
			})
		}

		// Start a reconnect timer in a separate goroutine.
		go r.reconnectTimer()

	case StatePaused:
		// Both gone — end immediately.
		r.endLocked("player_left")

	default:
		// ENDED — nothing to do.
	}
}

// Reconnect re-seats a returning player during the PAUSED window.
func (r *Room) Reconnect(c *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state != StatePaused {
		return fmt.Errorf("room %s not in PAUSED state", r.ID)
	}

	// Find the empty slot.
	slot := -1
	for i, p := range r.players {
		if p == nil {
			slot = i
			break
		}
	}
	if slot < 0 {
		return fmt.Errorf("room %s has no empty slot", r.ID)
	}

	r.players[slot] = c
	c.SetRoomID(r.ID)
	r.state = StatePlaying
	r.lastAction = time.Now()

	// Send full state snapshot to reconnected player.
	c.SendJSON(StateUpdate{
		Type:            "state_update",
		Tick:            r.tick,
		GameState:       json.RawMessage(`{}`), // TODO: real game state
		ServerTimestamp: time.Now().UnixMilli(),
	})

	return nil
}

// ---------- internal ----------

// run is the main tick loop; runs in its own goroutine.
func (r *Room) run() {
	ticker := time.NewTicker(TickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return

		case <-ticker.C:
			r.mu.Lock()
			if r.state != StatePlaying {
				continue
			}

			r.tick++
			tick := r.tick
			lastAction := r.lastAction
			r.mu.Unlock()

			// Idle check.
			idle := time.Since(lastAction)
			if idle >= IdleCloseAt {
				r.End("idle_timeout")
				return
			}
			if idle >= IdleWarnAt {
				r.broadcast(ErrorMsg{Type: "error", Code: "IDLE_WARNING", Message: "no actions detected; room will close soon"})
			}

			// Broadcast authoritative state.
			r.broadcast(StateUpdate{
				Type:            "state_update",
				Tick:            tick,
				GameState:       json.RawMessage(`{}`), // TODO: real game state
				ServerTimestamp: time.Now().UnixMilli(),
			})
		}
	}
}

// reconnectTimer waits for ReconnectWindow, then ends the room if still paused.
func (r *Room) reconnectTimer() {
	select {
	case <-time.After(ReconnectWindow):
		r.mu.RLock()
		st := r.state
		r.mu.RUnlock()
		if st == StatePaused {
			r.End("reconnect_timeout")
		}
	case <-r.ctx.Done():
		return
	}
}

// End terminates the room (thread-safe entry point).
func (r *Room) End(reason string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.endLocked(reason)
}

// endLocked does the actual teardown; caller MUST hold r.mu.
func (r *Room) endLocked(reason string) {
	if r.state == StateEnded {
		return
	}
	r.state = StateEnded

	// Notify connected players.
	for i, p := range r.players {
		if p == nil {
			continue
		}
		p.SendJSON(SessionEnd{
			Type:          "session_end",
			Outcome:       "incomplete",
			GroupXP:       0,
			Stars:         0,
			FinalState:    json.RawMessage(`{}`),
			RoomSessionID: r.ID,
		})
		p.SetRoomID("")
		r.players[i] = nil
	}

	// Write room_sessions row.
	go r.writeRoomSession(reason)

	// Clean up hub maps and cancel context.
	r.hub.removeRoom(r.ID)
	r.cancel()
}

// writeRoomSession persists the room outcome to the database.
func (r *Room) writeRoomSession(reason string) {
	if r.db == nil {
		return
	}

	// Collect profile IDs captured before players were nilled.
	// We store them at creation / add time — for now use a best-effort approach.
	// In a production build, cache profile IDs in the Room struct.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var disconnectReason *string
	if reason != "" {
		disconnectReason = &reason
	}

	_, err := r.db.Exec(ctx,
		`INSERT INTO room_sessions
			(id, game_type, player_1_profile_id, player_2_profile_id, completed, ended_at, disconnect_reason)
		 VALUES ($1, $2::game_zone, $3, $4, $5, now(), $6)`,
		r.ID, r.GameType,
		r.player1ProfileID(), r.player2ProfileID(),
		reason == "", // completed = true only if no abnormal reason
		disconnectReason,
	)
	if err != nil {
		log.Printf("ws: writeRoomSession room=%s error: %v", r.ID, err)
	}
}

// player1ProfileID / player2ProfileID return cached IDs; safe to call post-endLocked.
func (r *Room) player1ProfileID() string { return r.p1ID }
func (r *Room) player2ProfileID() string { return r.p2ID }

// broadcast sends v (JSON-marshalled) to every connected player in the room.
func (r *Room) broadcast(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("ws: broadcast marshal error room=%s: %v", r.ID, err)
		return
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.players {
		if p != nil {
			p.Send(data)
		}
	}
}

func (r *Room) otherPlayerLocked(slot int) *Client {
	return r.players[1-slot]
}
