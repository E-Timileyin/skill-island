package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/E-Timileyin/skill-island/services/api/internal/db"
	"github.com/E-Timileyin/skill-island/services/api/internal/ws/rooms"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamTowerRoom struct {
	ID               string
	GameType         string // "team_tower"
	State            rooms.TowerState
	Seed             int64
	Players          [2]*Client
	ProfileIDs       [2]string
	RoomState        string // "WAITING"|"READY"|"PLAYING"|"PAUSED"|"ENDED"
	Ticker           *time.Ticker
	Ctx              context.Context
	Cancel           context.CancelFunc
	DB               *pgxpool.Pool
	Hub              *Hub
	LastActionAt     time.Time
	DisconnectTimers map[string]*time.Timer
	mu               sync.RWMutex

	incomingActions chan actionEnv
	idleWarningSent bool
}

type actionEnv struct {
	client *Client
	msg    PlayerAction
}

func NewTeamTowerRoom(hub *Hub, db *pgxpool.Pool) *TeamTowerRoom {
	ctx, cancel := context.WithCancel(context.Background())
	seed := time.Now().UnixNano()
	
	return &TeamTowerRoom{
		ID:               uuid.New().String(),
		GameType:         "team_tower",
		State:            rooms.NewTowerState(seed),
		Seed:             seed,
		RoomState:        "WAITING",
		Ctx:              ctx,
		Cancel:           cancel,
		DB:               db,
		Hub:              hub,
		LastActionAt:     time.Now(),
		DisconnectTimers: make(map[string]*time.Timer),
		incomingActions:  make(chan actionEnv, 64),
	}
}

func (r *TeamTowerRoom) AddPlayer(client *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= 2 && r.Players[0] != nil && r.Players[1] != nil {
		return fmt.Errorf("room_full")
	}

	slot := -1
	for i, p := range r.Players {
		if p == nil {
			slot = i
			break
		}
	}
	if slot == -1 {
		return fmt.Errorf("room_full")
	}

	r.Players[slot] = client
	r.ProfileIDs[slot] = client.profileID
	client.SetRoomID(r.ID)

	if r.Players[0] != nil && r.Players[1] != nil {
		r.RoomState = "READY"
		
		r.Players[0].SendJSON(RoomReady{
			Type:           "room_ready",
			RoomID:         r.ID,
			PlayerRole:     "player_1",
			OpponentAvatar: 0,
		})
		r.Players[1].SendJSON(RoomReady{
			Type:           "room_ready",
			RoomID:         r.ID,
			PlayerRole:     "player_2",
			OpponentAvatar: 0,
		})
		
		r.RoomState = "PLAYING"
		// Start the room's main loop if not already running.
		go r.Run()
	}

	return nil
}

func (r *TeamTowerRoom) HandleAction(client *Client, msg PlayerAction) {
	// Instead of selecting, non-blocking send or just push if playing
	r.mu.RLock()
	if r.RoomState != "PLAYING" {
		r.mu.RUnlock()
		return
	}
	r.mu.RUnlock()
	
	select {
	case r.incomingActions <- actionEnv{client, msg}:
	default:
		log.Println("TeamTowerRoom.HandleAction full, dropping")
	}
}

func (r *TeamTowerRoom) RemovePlayer(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	slot := -1
	for i, p := range r.Players {
		if p != nil && p.profileID == client.profileID {
			slot = i
			break
		}
	}
	if slot == -1 {
		return
	}

	r.Players[slot] = nil
	client.SetRoomID("")
	
	if r.RoomState == "WAITING" {
		r.endSessionLocked("player_left")
		return
	}
	
	if r.RoomState == "READY" || r.RoomState == "PLAYING" {
		r.RoomState = "PAUSED"
		if r.Ticker != nil {
			r.Ticker.Stop()
		}
		
		role := "player_1"
		if slot == 1 { role = "player_2" }
		
		otherIdx := 1 - slot
		other := r.Players[otherIdx]
		if other != nil {
			other.SendJSON(PlayerDisconnected{
				Type: "player_disconnected",
				PlayerRole: role,
				Reason: "connection_lost",
				ReconnectWindowSeconds: 30,
			})
		}
		
		timer := time.AfterFunc(30 * time.Second, func() {
			r.mu.Lock()
			if r.Players[slot] == nil {
				r.endSessionLocked("incomplete")
			}
			r.mu.Unlock()
		})
		r.DisconnectTimers[client.profileID] = timer
		
		// Add to Hub reconnect
		r.Hub.mu.Lock()
		r.Hub.reconnecting[client.profileID] = r.ID
		r.Hub.mu.Unlock()
	} else if r.RoomState == "PAUSED" {
		// Both gone
		r.endSessionLocked("player_left")
	}
}

func (r *TeamTowerRoom) Reconnect(client *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.RoomState != "PAUSED" {
		return fmt.Errorf("room not paused")
	}

	slot := -1
	for i, p := range r.Players {
		if p == nil {
			slot = i
			break
		}
	}
	if slot == -1 {
		return fmt.Errorf("no empty slots")
	}

	r.Players[slot] = client
	client.SetRoomID(r.ID)
	r.RoomState = "PLAYING"
	
	if t, ok := r.DisconnectTimers[client.profileID]; ok {
		t.Stop()
		delete(r.DisconnectTimers, client.profileID)
	}

	r.Hub.mu.Lock()
	delete(r.Hub.reconnecting, client.profileID)
	r.Hub.mu.Unlock()

	// Resume game
	r.Ticker = time.NewTicker(90 * time.Millisecond)
	r.LastActionAt = time.Now()

	// Send full state to reconnecting player
	client.SendJSON(StateUpdate{
		Type: "state_update",
		Tick: 0, // placeholder, but would be r.State
		GameState: toRawJSON(r.State), // We'll serialize directly
		ServerTimestamp: time.Now().UnixMilli(),
	})

	// To other player
	other := r.Players[1-slot]
	if other != nil {
		other.SendJSON(Envelope{Type: "partner_reconnected"})
	}

	return nil
}

func (r *TeamTowerRoom) Run() {
	r.mu.Lock()
	r.Ticker = time.NewTicker(90 * time.Millisecond)
	r.mu.Unlock()
	
	idleChecker := time.NewTicker(10 * time.Second)
	
	var tickCount int64

	defer func() {
		r.mu.Lock()
		if r.Ticker != nil {
			r.Ticker.Stop()
		}
		r.mu.Unlock()
		idleChecker.Stop()
	}()

	for {
		select {
		case <-r.Ctx.Done():
			return

		case t := <-func() <-chan time.Time {
				r.mu.RLock()
				defer r.mu.RUnlock()
				if r.Ticker != nil {
					return r.Ticker.C
				}
				return nil
			}():
			_ = t
			r.mu.RLock()
			if r.RoomState != "PLAYING" {
				r.mu.RUnlock()
				continue
			}
			tickCount++
			
			// Broadcast state update
			update := StateUpdate{
				Type: "state_update",
				Tick: tickCount,
				GameState: toRawJSON(r.State),
				ServerTimestamp: time.Now().UnixMilli(),
			}
			msg, _ := json.Marshal(update)
			
			for _, p := range r.Players {
				if p != nil { p.Send(msg) }
			}
			r.mu.RUnlock()

		case msgEnv := <-r.incomingActions:
			r.mu.Lock()
			if r.RoomState != "PLAYING" {
				r.mu.Unlock()
				continue
			}
			
			// Parse the inner "payload" of PlayerAction OR the msgEnv.msg.Type
			// Wait, if Type == "player_action", client sends payload: { "type": "place_block" ... }
			// or client sends TopLevel: { "type": "place_block" }
			// Issue 20 says: "If type == "place_block": Parse TeamTowerAction"
			var action rooms.TeamTowerAction
			err := json.Unmarshal(msgEnv.msg.Payload, &action)
			if err != nil {
				// Maybe Payload is not used, and it's full msgEnv.msg?
				// Just try to parse it. Client.go uses msgEnv.msg.Payload.
			}

			if action.Type == "place_block" {
				playerRole := "player_1"
				if r.Players[1] != nil && msgEnv.client.profileID == r.Players[1].profileID {
					playerRole = "player_2"
				}

				result := rooms.ValidatePlacement(r.State, action, playerRole, r.Seed)
				if result.Error != nil {
					msgEnv.client.SendJSON(ActionRejected{
						Type: "action_rejected",
						Reason: result.Error.Error(),
					})
				} else {
					r.State = result.UpdatedState
					r.LastActionAt = time.Now()
					r.idleWarningSent = false
					
					if result.Outcome == "lose" {
						r.State.Stable = false
						r.endSessionLocked("lose")
						r.mu.Unlock()
						continue
					} else if result.Outcome == "win" {
						r.endSessionLocked("win")
						r.mu.Unlock()
						continue
					}
				}
			}
			r.mu.Unlock()

		case <-idleChecker.C:
			r.mu.Lock()
			if r.RoomState != "PLAYING" {
				r.mu.Unlock()
				continue
			}
			idleSeconds := time.Since(r.LastActionAt).Seconds()
			if idleSeconds > 120 {
				r.endSessionLocked("incomplete") // with idle_timeout
			} else if idleSeconds > 90 && !r.idleWarningSent {
				r.idleWarningSent = true
				msg := IdleWarning{Type: "idle_warning", SecondsRemaining: 30}
				for _, p := range r.Players {
					if p != nil { p.SendJSON(msg) }
				}
			}
			r.mu.Unlock()
		}
	}
}

func (r *TeamTowerRoom) End(reason string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.endSessionLocked(reason)
}

func (r *TeamTowerRoom) endSessionLocked(outcome string) {
	if r.RoomState == "ENDED" {
		return
	}
	r.RoomState = "ENDED"
	if r.Ticker != nil { 
		r.Ticker.Stop() 
	}
	
	// Convert "lose/win" outcome to result
	res := rooms.CalculateTeamTowerResult(r.State, outcome)
	
	msg := SessionEnd{
		Type: "session_end",
		Outcome: outcome,
		GroupXP: res.GroupXP,
		Stars: res.Stars,
		FinalState: toRawJSON(r.State),
		RoomSessionID: r.ID,
	}
	
	for i, p := range r.Players {
		if p != nil {
			p.SendJSON(msg)
			p.SetRoomID("")
			r.Players[i] = nil
		}
	}
	
	go r.writeSessionToDB(outcome, res)
	r.Cancel()
	r.Hub.removeRoom(r.ID)
}

func (r *TeamTowerRoom) writeSessionToDB(outcome string, res rooms.ScoredResult) {
	if r.DB == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// In the real version, we'd wrap these calls in a single transaction.
	
	completed := (outcome == "win" || outcome == "lose")
	var disconnectReason *string
	if !completed {
		if outcome == "incomplete" {
			dr := "idle_timeout" // or player_left, reconnect_timeout — to be refined
			disconnectReason = &dr
		} else {
			disconnectReason = &outcome
		}
	}

	p1 := r.ProfileIDs[0]
	p2 := r.ProfileIDs[1]

	err := db.WriteTeamTowerSession(ctx, r.DB, r.ID, p1, p2, res.GroupXP, completed, r.Seed, disconnectReason, res.Stars, res.XPPerPlayer)
	if err != nil {
		log.Printf("TeamTowerRoom write DB failed: %v", err)
	}
}

func toRawJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
