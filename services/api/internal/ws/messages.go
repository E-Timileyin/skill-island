package ws

import "encoding/json"

// ---------- Envelope for all messages ----------

// Envelope wraps every message with a type discriminator for JSON routing.
type Envelope struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"` // decoded payload; not re-serialised
}

// ---------- Client → Server ----------

// JoinRoom requests matchmaking for a specific game zone.
type JoinRoom struct {
	Type     string `json:"type"`      // "join_room"
	GameType string `json:"game_type"` // e.g. "team_tower"
}

// PlayerAction relays a game action to the server.
type PlayerAction struct {
	Type            string          `json:"type"`             // "player_action"
	Payload         json.RawMessage `json:"payload"`          // zone-specific; opaque at WS layer
	ClientTimestamp int64           `json:"client_timestamp"` // ms since session start
}

// HeartbeatPing keeps the connection alive.
type HeartbeatPing struct {
	Type      string `json:"type"`      // "heartbeat_ping"
	Timestamp int64  `json:"timestamp"` // client clock ms
}

// LeaveRoom is an intentional room departure.
type LeaveRoom struct {
	Type string `json:"type"` // "leave_room"
}

// ---------- Server → Client ----------

// RoomReady is sent when both players have connected.
type RoomReady struct {
	Type           string `json:"type"`            // "room_ready"
	RoomID         string `json:"room_id"`         //
	PlayerRole     string `json:"player_role"`     // "player_1" | "player_2"
	OpponentAvatar int    `json:"opponent_avatar"` //
}

// StateUpdate is the authoritative tick broadcast.
type StateUpdate struct {
	Type            string          `json:"type"`             // "state_update"
	Tick            int64           `json:"tick"`             //
	GameState       json.RawMessage `json:"game_state"`       // zone-specific state
	ServerTimestamp int64           `json:"server_timestamp"` // unix ms
}

// PlayerDisconnected notifies the remaining player.
type PlayerDisconnected struct {
	Type                   string `json:"type"`                     // "player_disconnected"
	PlayerRole             string `json:"player_role"`              // role of the player who left
	Reason                 string `json:"reason"`                   //
	ReconnectWindowSeconds int    `json:"reconnect_window_seconds"` // 30
}

// IdleWarning warns players of impending session closure.
type IdleWarning struct {
	Type             string `json:"type"`              // "idle_warning"
	SecondsRemaining int    `json:"seconds_remaining"` // 30
}

// SessionEnd signals that the game session is over.
type SessionEnd struct {
	Type          string          `json:"type"`            // "session_end"
	Outcome       string          `json:"outcome"`         // "win" | "lose" | "incomplete"
	GroupXP       int             `json:"group_xp"`        //
	Stars         int             `json:"stars"`           //
	FinalState    json.RawMessage `json:"final_state"`     // zone-specific
	RoomSessionID string          `json:"room_session_id"` //
}

// HeartbeatPong responds to a client ping.
type HeartbeatPong struct {
	Type       string `json:"type"`        // "heartbeat_pong"
	Timestamp  int64  `json:"timestamp"`   // echoed from client
	ServerTime int64  `json:"server_time"` // server unix ms
}

// ActionRejected tells the client an action was invalid.
type ActionRejected struct {
	Type   string `json:"type"`   // "action_rejected"
	Reason string `json:"reason"` //
}

// ErrorMsg is a generic error pushed to the client.
type ErrorMsg struct {
	Type    string `json:"type"`    // "error"
	Code    string `json:"code"`    //
	Message string `json:"message"` //
}
