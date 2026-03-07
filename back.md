# PHASE 2 CONTINUED — Focus Forest

---

## Issue 17 — Focus Forest Server Validator

```
Title: Phase 2.5 — Focus Forest server-side validator

Read .github/copilot-instructions.md and .github/game-logic-mvp.md before starting.
Focus entirely on the "Game Logic — Focus Forest" section.

Build the complete server-side validator for Focus Forest:

1. internal/validator/focus.go

   Types:

   type SpawnEvent struct {
       TargetID    string  `json:"target_id"`
       TargetType  string  `json:"target_type"`    // "butterfly_blue" | "butterfly_orange" | "butterfly_red" | "bee"
       SpawnTimeMs int     `json:"spawn_time_ms"`  // ms from session start
       PositionX   float64 `json:"position_x"`     // 0.0–1.0 normalised
       PositionY   float64 `json:"position_y"`     // 0.0–1.0 normalised
   }

   type FocusForestAction struct {
       Type            string  `json:"type"`             // always "tap"
       TapX            float64 `json:"tap_x"`            // 0.0–1.0 normalised
       TapY            float64 `json:"tap_y"`            // 0.0–1.0 normalised
       ClientTimestamp int64   `json:"client_timestamp"` // ms since session start
   }

   Functions:

   GenerateSpawnManifest(seed int64, durationMs int, difficultyLevel int) []SpawnEvent
   — Uses seeded random source (deterministic — same seed always = same manifest)
   — Difficulty rules:
       Level 1: 30% bees, speed 80px/s,  spawn every 1200ms
       Level 2: 40% bees, speed 100px/s, spawn every 1000ms
       Level 3: 50% bees, speed 130px/s, spawn every 800ms
       Level 4: 55% bees, speed 160px/s, spawn every 700ms (max)
   — Each target assigned a TargetID (UUID), TargetType, SpawnTimeMs, PositionX, PositionY
   — PositionX: spawn off left edge (0.0); target moves right to 1.0
   — PositionY: random value in [0.1, 0.9] to avoid edges

   GetTargetPositionAtTime(spawn SpawnEvent, timestampMs int, speedPxPerS float64) (x, y float64)
   — Linear interpolation of target X position based on movement speed
   — x = spawn.PositionX + (speedPxPerS / screenWidth) * (timestampMs - spawn.SpawnTimeMs) / 1000
   — y = spawn.PositionY (constant — targets move horizontally only in MVP)
   — Clamp x to [0.0, 1.2] — allow slightly off screen before despawn

   ValidateTaps(actions []FocusForestAction, manifest []SpawnEvent, difficultyLevel int) ValidationResult
   — Constants:
       HIT_RADIUS       = 0.08  // normalised units (server-defined — client cannot change)
       DESPAWN_AFTER_MS = 3000  // target removed after 3 seconds if not tapped
   — For each tap action:
       1. Find all targets where:
          spawn.SpawnTimeMs <= action.ClientTimestamp AND
          action.ClientTimestamp <= spawn.SpawnTimeMs + DESPAWN_AFTER_MS
       2. Lerp each active target's current X position using GetTargetPositionAtTime
       3. Calculate distance = sqrt((tap_x - target_x)^2 + (tap_y - target_y)^2)
       4. Find nearest target within HIT_RADIUS
       5. If butterfly found:
          — correct = true
          — reaction_time_ms = action.ClientTimestamp - spawn.SpawnTimeMs
       6. If bee found:
          — correct = false (distraction event)
          — reaction_time_ms = action.ClientTimestamp - spawn.SpawnTimeMs
          — NO score penalty — SEND-friendly design
       7. If nothing within radius:
          — Record as missed_tap (correct = false, metadata: { was_miss: true })
       8. Write behavioral_metrics row per tap

   CalculateAttentionScore(result ValidationResult) ScoredResult
   — butterfly_hits    = count of correct taps matching butterfly targets
   — bee_hits          = count of taps matching bee targets
   — butterflies_total = total butterfly targets in manifest
   — attention_score   = (butterfly_hits - (bee_hits * 0.5)) / butterflies_total
   — Clamp attention_score to [0.0, 1.0]
   — Stars:
       3 stars → attention_score >= 0.85
       2 stars → attention_score >= 0.65
       1 star  → attention_score >= 0.40
       0 stars → attention_score <  0.40
   — xp_earned = (stars * 10) + floor(attention_score * 50)

2. internal/validator/focus_test.go

   Unit tests (all must pass with go test -race ./internal/validator/...):

   TestGenerateSpawnManifest_Deterministic
   — Same seed + same params always produces identical manifest

   TestGenerateSpawnManifest_DifficultyRatios
   — Level 1 manifest has ~30% bees (±5% tolerance for small manifests)
   — Level 4 manifest has ~55% bees (±5% tolerance)

   TestGetTargetPositionAtTime_Interpolation
   — At SpawnTimeMs: x = spawn.PositionX
   — At SpawnTimeMs + 1000: x has moved by (speed/screenWidth)

   TestValidateTaps_ButterflyHit
   — Tap within HIT_RADIUS of active butterfly → correct = true

   TestValidateTaps_BeeHit
   — Tap within HIT_RADIUS of active bee → correct = false (distraction, not miss)

   TestValidateTaps_Miss
   — Tap with no target within HIT_RADIUS → recorded as missed tap

   TestValidateTaps_ExpiredTarget
   — Tap on target after DESPAWN_AFTER_MS → treated as missed tap (target gone)

   TestCalculateAttentionScore_StarThresholds
   — attention_score 0.90 → 3 stars
   — attention_score 0.70 → 2 stars
   — attention_score 0.45 → 1 star
   — attention_score 0.20 → 0 stars

   TestCalculateAttentionScore_BeeHitsNeverGoNegative
   — All bees hit, no butterflies → attention_score clamped to 0.0 not negative

   TestCalculateAttentionScore_XPCalculation
   — 3 stars + attention_score 0.90 → xp = 30 + floor(0.90 * 50) = 30 + 45 = 75
   — 0 stars + attention_score 0.20 → xp = 0 + floor(0.20 * 50) = 10

All validation functions must be pure — no side effects, no DB calls.
```

---

## Issue 18-BE — Focus Forest API Endpoints

```
Title: Phase 2.6-BE — Focus Forest session init, manifest, and submission endpoints

Read .github/copilot-instructions.md and .github/game-logic-mvp.md before starting.
This is the backend half of Issue 18. The frontend half runs in parallel.

─────────────────────────────────────────
PART A — Session Init Update
─────────────────────────────────────────

1. POST /api/sessions/init (update from Issue 14)
   Add support for game_type: "focus_forest"

   Additional response field for Focus Forest:
   {
     session_token: string,
     seed: int64,
     difficulty_level: int,   // 1–4; determined server-side from profile history
     session_duration_ms: 60000
   }

   Difficulty level determination:
   — Fetch last 5 focus_forest sessions for this profile
   — avg_attention_score >= 0.85 AND count >= 3 → level up (max 4)
   — avg_attention_score < 0.40 AND count >= 3 → level down (min 1)
   — Otherwise: keep current level (default 1 for new profiles)

2. GET /api/sessions/manifest (new endpoint)
   Query params: token (session_token)

   Flow:
   a. Look up pending_sessions by token
   b. Validate token not expired (expires_at > now())
   c. Regenerate spawn manifest from stored seed + difficulty_level
      (call internal/validator/focus.go GenerateSpawnManifest)
   d. Return manifest array to client
   e. Do NOT consume token yet — token consumed on final POST /api/sessions

   Returns:
   [
     { target_id, target_type, spawn_time_ms, position_x, position_y },
     ...
   ]

   Client uses manifest to position spawns correctly.
   Server regenerates manifest from seed on POST /api/sessions to validate taps.
   Never trust client-submitted positions.

3. POST /api/sessions (update from Issue 14)
   Add handler for game_type: "focus_forest"

   Flow:
   a. Parse SessionSubmission { session_token, game_type, mode, actions, duration_ms }
   b. Look up pending_sessions by token — reject if not found or expired
   c. Regenerate manifest from stored seed + difficulty_level
   d. Call internal/validator/focus.go ValidateTaps(actions, manifest, difficultyLevel)
   e. Reject if actions.length > 300 (implausible for 60s session)
   f. Reject if duration_ms < 5000 (impossibly short session)
   g. Open pgx transaction:
      — Write game_sessions row (server score only)
      — Write behavioral_metrics rows for all tap actions
      — AddXPToProfile
      — AddStarsToProfile
      — Mark pending_sessions token as used
      — Commit
   h. Call CheckUnlockedZones
   i. Return SessionResult

Phase 2-BE is complete when:
— POST /api/sessions/init returns seed + difficulty_level for focus_forest
— GET /api/sessions/manifest returns correct deterministic spawn manifest
— POST /api/sessions for focus_forest validates taps server-side and writes DB
— go test -race ./... passes
```

---

---

# PHASE 3 — Multiplayer (Team Tower)

---

## Issue 19 — Team Tower Server Game Logic

```
Title: Phase 3.1 — Team Tower server-side game logic and physics

Read .github/copilot-instructions.md and .github/game-logic-mvp.md before starting.
Focus entirely on "Game Logic — Team Tower" section.

─────────────────────────────────────────
PART A — Types
─────────────────────────────────────────

1. internal/ws/rooms/teamtower.go

   type Block struct {
       ID        string  `json:"id"`
       Shape     string  `json:"shape"`     // "1x1"|"2x1"|"1x2"|"L_shape"|"T_shape"
       Colour    string  `json:"colour"`    // "blue"|"red"|"green"|"orange"|"grey"
       X         float64 `json:"x"`         // centre position, normalised 0.0–1.0
       Y         float64 `json:"y"`         // height from tower base, normalised
       Rotation  float64 `json:"rotation"`  // degrees; 0 = upright
       PlacedBy  string  `json:"placed_by"` // "player_1" | "player_2"
   }

   type TowerState struct {
       Blocks         []Block `json:"blocks"`
       CurrentHeight  float64 `json:"current_height"`
       TargetHeight   float64 `json:"target_height"`
       ActivePlayer   string  `json:"active_player"`   // "player_1" | "player_2"
       NextBlockShape string  `json:"next_block_shape"`
       Stable         bool    `json:"stable"`           // false = tower fallen
       GroupXP        int     `json:"group_xp"`
       TurnNumber     int     `json:"turn_number"`
   }

   type TeamTowerAction struct {
       Type            string  `json:"type"`             // always "place_block"
       PositionX       float64 `json:"position_x"`       // 0.0–1.0
       ClientTimestamp int64   `json:"client_timestamp"`
   }

   type PlacementResult struct {
       UpdatedState TowerState
       Outcome      string // "" | "win" | "lose"
       Error        error
   }

─────────────────────────────────────────
PART B — Functions
─────────────────────────────────────────

   NewTowerState(seed int64) TowerState
   — TargetHeight = 10.0
   — CurrentHeight = 0.0
   — Stable = true
   — ActivePlayer = "player_1"
   — TurnNumber = 1
   — GroupXP = 0
   — Blocks = []
   — NextBlockShape = GetNextBlockShape(seed, 1)

   GetNextBlockShape(seed int64, turnNumber int) string
   — Deterministic: same seed + turnNumber always returns same shape
   — Weighted distribution:
     "1x1"     → 30%
     "2x1"     → 25%
     "1x2"     → 20%
     "L_shape" → 15%
     "T_shape" → 10%

   GetBlockDimensions(shape string) (width, height float64)
   — "1x1"     → width: 0.10, height: 0.05
   — "2x1"     → width: 0.20, height: 0.05
   — "1x2"     → width: 0.10, height: 0.10
   — "L_shape" → width: 0.15, height: 0.10
   — "T_shape" → width: 0.20, height: 0.08

   GetBlockWeight(shape string) float64
   — "1x1"     → 1.0
   — "2x1"     → 2.0
   — "1x2"     → 1.5
   — "L_shape" → 2.5
   — "T_shape" → 2.5

   ValidatePlacement(state TowerState, action TeamTowerAction, playerRole string, seed int64) PlacementResult

   Step 1 — Turn validation:
   — If playerRole != state.ActivePlayer:
     return PlacementResult{ Error: errors.New("not_your_turn") }

   Step 2 — Bounds validation:
   — If action.PositionX < 0.05 OR action.PositionX > 0.95:
     return PlacementResult{ Error: errors.New("out_of_bounds") }

   Step 3 — Calculate landing Y:
   — Get block dimensions for state.NextBlockShape
   — Find highest existing block whose X range overlaps action.PositionX
   — landingY = highest_block.Y + highest_block.height + new_block.height/2
   — If no existing blocks: landingY = new_block.height / 2 (rests on floor)

   Step 4 — Create new block:
   — newBlock = Block{
       ID:       uuid.New().String(),
       Shape:    state.NextBlockShape,
       Colour:   getRandomColour(seed, state.TurnNumber),
       X:        action.PositionX,
       Y:        landingY,
       Rotation: 0,
       PlacedBy: playerRole,
     }
   — Append to state.Blocks

   Step 5 — Balance check:
   — totalWeight  = sum of GetBlockWeight for all blocks
   — weightedX    = sum of (block.X * GetBlockWeight(block.Shape)) for all blocks
   — centreOfMass = weightedX / totalWeight
   — If abs(centreOfMass - 0.5) > 0.35:
     state.Stable = false
     return PlacementResult{ UpdatedState: state, Outcome: "lose" }

   Step 6 — Update height:
   — state.CurrentHeight = max Y value of all blocks + block.height/2

   Step 7 — Win check:
   — If state.CurrentHeight >= state.TargetHeight:
     return PlacementResult{ UpdatedState: state, Outcome: "win" }

   Step 8 — Advance turn:
   — state.GroupXP += 5
   — state.TurnNumber++
   — Switch state.ActivePlayer: "player_1" ↔ "player_2"
   — state.NextBlockShape = GetNextBlockShape(seed, state.TurnNumber)
   — return PlacementResult{ UpdatedState: state, Outcome: "" }

   CalculateTeamTowerResult(state TowerState, outcome string) ScoredResult
   — Win AND currentHeight >= targetHeight * 1.1 → 3 stars
   — Win (any)                                   → 2 stars
   — Incomplete (disconnect, idle timeout)        → 1 star (SEND rule)
   — Lose (tower fall)                            → 1 star (never 0 in co-op)
   — groupXP = state.GroupXP + (20 if win, else 0)
   — xpPerPlayer = groupXP / 2 (integer division)

─────────────────────────────────────────
PART C — Tests
─────────────────────────────────────────

2. internal/validator/teamtower_test.go

   TestValidatePlacement_WrongTurn
   — player_2 submits when ActivePlayer is "player_1" → error "not_your_turn"

   TestValidatePlacement_OutOfBounds
   — position_x = 0.03 → error "out_of_bounds"
   — position_x = 0.97 → error "out_of_bounds"
   — position_x = 0.50 → no error

   TestValidatePlacement_StacksCorrectly
   — Place block at x=0.5 → Y = block_height/2
   — Place second block at x=0.5 → Y > first block Y

   TestValidatePlacement_BalanceCheck
   — Stack 5 blocks at x=0.9 → Stable = false, Outcome = "lose"
   — Stack 5 blocks alternating x=0.3 and x=0.7 → Stable = true

   TestValidatePlacement_WinCondition
   — Simulate placements until CurrentHeight >= TargetHeight → Outcome = "win"

   TestGetNextBlockShape_Deterministic
   — Same seed + turnNumber always returns same shape

   TestCalculateTeamTowerResult_Stars
   — Win + extra height → 3 stars
   — Win normal → 2 stars
   — Incomplete → 1 star
   — Lose (tower fall) → 1 star (not 0)

   TestCalculateTeamTowerResult_XPSplit
   — 10 placements + win → groupXP = 50 + 20 = 70 → 35 XP per player

All functions pure — no DB calls, no side effects.
go test -race ./... must pass.
```

---

## Issue 20 — Team Tower WebSocket Room

```
Title: Phase 3.2 — Team Tower WebSocket room integration

Read .github/copilot-instructions.md, .github/data-model.md,
and .github/game-logic-mvp.md before starting.

─────────────────────────────────────────
PART A — Room Structure
─────────────────────────────────────────

1. internal/ws/rooms/teamtower_room.go

   type TeamTowerRoom struct {
       ID               string
       GameType         string // "team_tower"
       State            rooms.TowerState
       Seed             int64
       Players          [2]*ws.Client
       RoomState        string  // "WAITING"|"READY"|"PLAYING"|"PAUSED"|"ENDED"
       Ticker           *time.Ticker
       Ctx              context.Context
       Cancel           context.CancelFunc
       DB               *pgxpool.Pool
       LastActionAt     time.Time
       DisconnectTimers map[string]*time.Timer
       mu               sync.RWMutex
   }

─────────────────────────────────────────
PART B — Room Lifecycle
─────────────────────────────────────────

2. NewTeamTowerRoom(db *pgxpool.Pool) *TeamTowerRoom
   — Generate UUID room ID
   — Generate random seed
   — Initialise TowerState with NewTowerState(seed)
   — Create context with cancel
   — RoomState = "WAITING"

3. AddPlayer(client *ws.Client) error
   — If len(Players) >= 2: return error "room_full"
   — Assign player role: first = "player_1", second = "player_2"
   — If both players present: transition to READY → send room_ready to both
   — room_ready payload:
     {
       type: "room_ready",
       room_id: string,
       player_role: "player_1" | "player_2",
       opponent_avatar: int
     }

4. Run() — main room goroutine
   — Start ticker: time.NewTicker(90 * time.Millisecond)
   — Start idle checker: time.NewTicker(10 * time.Second)

   Loop (select):

   case <-ticker.C:
   — If RoomState != "PLAYING": skip
   — Marshal TowerState as state_update message
   — Send to both connected players simultaneously
   — state_update payload:
     {
       type: "state_update",
       tick: int,
       game_state: TowerState,
       server_timestamp: int64
     }

   case msg := <-incomingActions:
   — Parse message type
   — If type == "place_block":
     — Parse TeamTowerAction
     — Call ValidatePlacement(state, action, playerRole, seed)
     — If error:
       Send action_rejected to that player only:
       { type: "action_rejected", reason: "not_your_turn" | "out_of_bounds" }
     — If Outcome == "lose":
       state.Stable = false; transition to ENDED; call endSession("lose")
     — If Outcome == "win":
       transition to ENDED; call endSession("win")
     — If no outcome: update state; update LastActionAt
   — If type == "heartbeat_ping":
     Update client.LastPingAt
     Send heartbeat_pong immediately
   — If type == "leave_room":
     transition to ENDED; call endSession("incomplete")

   case <-idleChecker.C:
   — If RoomState != "PLAYING": skip
   — idleSeconds = time.Since(LastActionAt).Seconds()
   — If idleSeconds > 90:
     Send idle_warning to both: { type: "idle_warning", seconds_remaining: 30 }
   — If idleSeconds > 120:
     transition to ENDED; call endSession("incomplete")
     disconnect_reason = "idle_timeout"

   case <-ctx.Done():
   — Stop ticker; return

─────────────────────────────────────────
PART C — Session Write
─────────────────────────────────────────

5. endSession(outcome string)

   a. Transition RoomState to "ENDED"
   b. Stop ticker
   c. Calculate result using CalculateTeamTowerResult(state, outcome)
   d. Build session_end message:
      {
        type: "session_end",
        outcome: string,
        group_xp: int,
        stars: int,
        final_state: TowerState,
        room_session_id: string
      }
   e. Send session_end to all connected players
   f. Open pgx transaction:
      — Write room_sessions row:
        { game_type, player_1_profile_id, player_2_profile_id,
          group_xp_earned, completed, started_at, ended_at, disconnect_reason }
      — Write game_sessions row for player_1 (mode: "cooperative")
      — Write game_sessions row for player_2 (mode: "cooperative")
      — Write behavioral_metrics for all placement actions (both players)
      — AddXPToProfile for player_1 (xpPerPlayer)
      — AddXPToProfile for player_2 (xpPerPlayer)
      — AddStarsToProfile for player_1
      — AddStarsToProfile for player_2
      — Commit
   g. If transaction fails: log error with room_id; retry once after 1s
   h. Call cancel() to stop room context
   i. Hub removes room from active rooms map

   completed = (outcome == "win" || outcome == "lose")
   disconnect_reason = NULL if completed;
                       "idle_timeout" | "player_left" | "reconnect_timeout" if not

─────────────────────────────────────────
PART D — Hub Integration
─────────────────────────────────────────

6. internal/ws/hub.go (update)

   Matchmaking queue per game_type:
   — When client sends join_room { game_type: "team_tower" }:
     a. Check if client already in active room → reject with "already_in_room"
     b. Check matchmaking queue for "team_tower"
     c. If queue empty: add client to queue; send:
        { type: "waiting_for_partner" }
     d. If queue has one waiting client:
        — Remove from queue
        — Create new TeamTowerRoom
        — AddPlayer(waitingClient); AddPlayer(newClient)
        — Start room.Run() in goroutine
        — Register room in hub.rooms map

   Profile → Room index: map[profileID]roomID
   — On join: add entry
   — On room ENDED: remove entry
   — Enforce: if profileID already in rooms map → reject new join_room

7. internal/ws/messages.go (add all Team Tower message structs)

   All message structs with JSON tags:
   — JoinRoomMessage            { Type, GameType }
   — PlaceBlockMessage          { Type, PositionX, ClientTimestamp }
   — ActionRejectedMessage      { Type, Reason }
   — WaitingMessage             { Type }
   — RoomReadyMessage           { Type, RoomID, PlayerRole, OpponentAvatar }
   — TowerStateUpdate           { Type, Tick, GameState, ServerTimestamp }
   — SessionEndMessage          { Type, Outcome, GroupXP, Stars, FinalState, RoomSessionID }
   — IdleWarningMessage         { Type, SecondsRemaining }
   — PlayerDisconnectedMessage  { Type, PlayerRole, Reason, ReconnectWindowSeconds }
   — HeartbeatPingMessage       { Type, Timestamp }
   — HeartbeatPongMessage       { Type, Timestamp, ServerTime }

─────────────────────────────────────────
PART E — Race Safety
─────────────────────────────────────────

8. All shared state in TeamTowerRoom protected by sync.RWMutex
   — Read lock for state broadcasts
   — Write lock for state mutations (ValidatePlacement)

9. Every goroutine inside Run():
   — Select on ctx.Done() to exit cleanly
   — Never reference room fields after cancel() is called

10. CI: go test -race ./... must pass with zero race conditions
```

---

## Issue 23-BE — Server Heartbeat, Reconnect Window & Idle Timeout

```
Title: Phase 3.5-BE — WebSocket heartbeat, reconnect window, and idle timeout (server)

Read .github/copilot-instructions.md and .github/data-model.md before starting.
This is the backend half of Issue 23. The frontend reconnect client runs in parallel.

─────────────────────────────────────────
PART A — Server Heartbeat (internal/ws/client.go)
─────────────────────────────────────────

1. Client struct additions:
   LastPingAt   time.Time
   WriteTimeout time.Duration  // default 10s

2. writePump goroutine:
   — Set write deadline before every send:
     conn.SetWriteDeadline(time.Now().Add(client.WriteTimeout))
   — On write error (timeout or connection closed):
     log.Warn("client write failed", profileID, error)
     Notify hub to disconnect client cleanly

3. Background stale checker (per client, runs in goroutine):
   — Ticker: every 5 seconds
   — If time.Since(LastPingAt) > 15 * time.Second:
     log.Info("client stale — no ping", profileID)
     conn.Close()
     Notify room of disconnect
   — Exits on ctx.Done()

4. On heartbeat_ping received:
   — client.LastPingAt = time.Now()
   — Send heartbeat_pong immediately:
     { type: "heartbeat_pong", timestamp: ping.Timestamp, server_time: time.Now().UnixMilli() }

─────────────────────────────────────────
PART B — Reconnect Window (teamtower_room.go update)
─────────────────────────────────────────

5. On player disconnect detected (write error or stale):
   a. Identify which player disconnected by profileID
   b. Set Players[index] = nil
   c. Transition RoomState to "PAUSED"
   d. Stop main game ticker (preserve state)
   e. Send to remaining connected player:
      {
        type: "player_disconnected",
        player_role: "player_1" | "player_2",
        reason: "connection_lost",
        reconnect_window_seconds: 30
      }
   f. Start reconnect timer: time.AfterFunc(30 * time.Second, func() {
        if Players[index] == nil {
          endSession("incomplete") with disconnect_reason = "reconnect_timeout"
        }
      })
   g. Register disconnected profileID in hub reconnect index:
      hub.reconnecting[profileID] = roomID

6. On client reconnect (sends join_room for same game_type):
   a. Hub checks hub.reconnecting[profileID]
   b. If roomID found AND room is in PAUSED state:
      — Attach client to room Players slot
      — Delete from hub.reconnecting
      — Cancel reconnect timer
      — Transition RoomState to "PLAYING"
      — Resume main game ticker
      — Send full state snapshot to reconnected player:
        { type: "state_update", tick: currentTick, game_state: fullTowerState, ... }
      — Send to other player: { type: "partner_reconnected" }
   c. If no roomID found: treat as new matchmaking join

─────────────────────────────────────────
PART C — Idle Timeout (teamtower_room.go update)
─────────────────────────────────────────

7. LastActionAt: time.Time — updated on every place_block action

8. Idle checker goroutine (every 10 seconds):
   a. idleDuration = time.Since(room.LastActionAt)
   b. If idleDuration > 90s AND RoomState == "PLAYING":
      — Send to both: { type: "idle_warning", seconds_remaining: 30 }
      — Set idleWarningSent = true
   c. If idleDuration > 120s AND RoomState == "PLAYING":
      — Call endSession("incomplete")
      — disconnect_reason = "idle_timeout"
   d. On any place_block: reset LastActionAt; set idleWarningSent = false

─────────────────────────────────────────
PART D — Goroutine Safety
─────────────────────────────────────────

9. All goroutines (ticker, idle checker, stale checker, reconnect timer):
   — Must select on ctx.Done() to exit
   — Must not access room fields after cancel() called
   — Must not send on closed channels

10. CI: go test -race ./internal/ws/... must pass with zero data races
    Add race condition test: TestRoomConcurrentActions
    — Simulate 2 goroutines sending actions simultaneously
    — Verify no race condition on TowerState mutation
```

---

## Issue 24-BE — Stats API, Achievement Checker & Session History

```
Title: Phase 3.6-BE — Stats API, achievement checker, and session history endpoints

Read .github/copilot-instructions.md and .github/data-model.md before starting.
This is the backend half of Issue 24. The frontend progress screen runs in parallel.

─────────────────────────────────────────
PART A — Stats API
─────────────────────────────────────────

1. GET /api/profiles/me/stats (internal/api/stats_handler.go)

   Requires RequireAuth + student role

   Returns:
   {
     total_stars: number,
     total_xp: number,
     sessions_total: number,
     sessions_this_week: number,
     favourite_zone: string | null,
     best_accuracy: number,
     cooperative_sessions: number,
     current_streak_days: number,
     unlocked_zones: string[],
     achievements: Achievement[]
   }

2. internal/db/stats_queries.go

   GetProfileStats(ctx, profileID) → ProfileStats, error:
   — Single SQL query joining game_sessions + student_profiles
   — sessions_this_week: WHERE created_at >= NOW() - INTERVAL '7 days'
   — favourite_zone: GROUP BY game_type ORDER BY COUNT(*) DESC LIMIT 1
   — best_accuracy: MAX(accuracy) across all sessions

   GetCurrentStreak(ctx, profileID) → int, error:
   — Fetch distinct session dates for last 30 days
   — Count consecutive days ending today
   — Day with no sessions = streak broken

   GetAchievements(ctx, profileID) → []Achievement, error

3. internal/validator/achievements.go

   CheckAndAwardAchievements(ctx, db, profileID) error:
   Fetch current stats; check each condition:

   "great_job"        → any session with stars_earned == 3
   "top_team"         → any room_sessions with completed = true where both stars = 3
   "island_explorers" → has sessions in all 3 MVP zones
   "treasure_found"   → total_stars >= 100
   "skills_unlocked"  → total_xp >= 80 (all 3 MVP zones unlocked)
   "island_success"   → sessions_total >= 20
   "quest_completed"  → current_streak_days >= 7

   For each achievement not yet in achievements table:
   — INSERT INTO achievements (profile_id, type, earned_at)

   Idempotent: never duplicate-award same achievement to same profile.
   Call CheckAndAwardAchievements after every session write in session handler.

─────────────────────────────────────────
PART B — Session History API
─────────────────────────────────────────

4. GET /api/profiles/me/sessions

   Query params:
   — limit: int (default 10, max 50)
   — offset: int (default 0)
   — game_type: string (optional filter)

   Returns:
   {
     sessions: [
       { id, game_type, mode, score, accuracy, stars_earned, duration_seconds, created_at }
     ],
     total: number,
     limit: number,
     offset: number
   }

   Never return behavioral_metrics in this response.
   Never return room_sessions data directly.

5. internal/db/session_queries.go (additions)

   GetSessionHistory(ctx, profileID, limit, offset, gameType) → []SessionSummary, error
   GetSessionCount(ctx, profileID, gameType) → int, error
```

---

---

# PHASE 4 — Analytics, Polish & Beta Launch

---

## Issue 25 — Nightly Analytics Aggregation Job

```
Title: Phase 4.1 — Nightly analytics aggregation job

Read .github/copilot-instructions.md and .github/data-model.md before starting.
Focus on "Analytics Aggregation" section in data-model.md.

─────────────────────────────────────────
PART A — Job Implementation
─────────────────────────────────────────

1. internal/jobs/aggregation.go

   type AggregationJob struct {
       DB  *pgxpool.Pool
       Log *slog.Logger
   }

   type AggregationResult struct {
       ProfilesProcessed int
       ProfilesFailed    int
       Duration          time.Duration
       Errors            []AggregationError
   }

   type AggregationError struct {
       ProfileID string
       Error     string
   }

2. AggregateProfile(ctx context.Context, profileID string, date time.Time) error

   Runs in a single transaction — all-or-nothing per profile:

   a. Fetch sessions from last 7 days:
      SELECT game_type, accuracy, mode, created_at
      FROM game_sessions
      WHERE profile_id = $1
        AND created_at >= $2 - INTERVAL '7 days'
        AND created_at < $2 + INTERVAL '1 day'

   b. Fetch behavioral_metrics for those sessions:
      SELECT bm.reaction_time_ms, bm.hesitation_ms, bm.retry_count
      FROM behavioral_metrics bm
      JOIN game_sessions gs ON gs.id = bm.session_id
      WHERE gs.profile_id = $1
        AND gs.created_at >= $2 - INTERVAL '7 days'

   c. Calculate metrics:
      attention_score         = AVG(accuracy) WHERE game_type = 'focus_forest'
      memory_score            = AVG(accuracy) WHERE game_type = 'memory_cove'
      engagement_frequency    = COUNT(*) all sessions
      coop_participation_rate = COUNT(*) WHERE mode = 'cooperative' / COUNT(*) total
                                (return 0.0 if no sessions)
      avg_reaction_time_ms    = AVG(reaction_time_ms) WHERE reaction_time_ms IS NOT NULL
      avg_hesitation_ms       = AVG(hesitation_ms) WHERE hesitation_ms IS NOT NULL
      retry_rate              = AVG(retry_count)

   d. Upsert into analytics_snapshots:
      INSERT INTO analytics_snapshots (
        profile_id, snapshot_date,
        attention_score, memory_score, engagement_frequency,
        coop_participation_rate, avg_reaction_time_ms,
        avg_hesitation_ms, retry_rate
      ) VALUES ($1, $2, ...)
      ON CONFLICT (profile_id, snapshot_date) DO UPDATE SET
        attention_score = EXCLUDED.attention_score,
        memory_score = EXCLUDED.memory_score,
        -- all other fields

   Idempotent: running twice for same profile + date produces same result.

3. RunAll(ctx context.Context) AggregationResult

   a. Fetch all profile IDs:
      SELECT id FROM student_profiles WHERE deleted_at IS NULL
   b. For each profileID:
      — Run AggregateProfile; on error: log, add to errors list, CONTINUE
      — Never let one profile failure stop the entire run
   c. Return AggregationResult with counts and duration
   Total runtime target: < 30 seconds for 1000 profiles

─────────────────────────────────────────
PART B — Scheduler
─────────────────────────────────────────

4. cmd/server/main.go (update)

   func scheduleNightlyAggregation(job *jobs.AggregationJob) {
       now := time.Now().UTC()
       next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 5, 0, 0, time.UTC)
       initialDelay := time.Until(next)

       time.AfterFunc(initialDelay, func() {
           runAggregation(job)
           ticker := time.NewTicker(24 * time.Hour)
           for range ticker.C {
               runAggregation(job)
           }
       })
   }

   func runAggregation(job *jobs.AggregationJob) {
       ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
       defer cancel()
       result := job.RunAll(ctx)
       slog.Info("aggregation complete",
           "profiles_processed", result.ProfilesProcessed,
           "profiles_failed", result.ProfilesFailed,
           "duration_ms", result.Duration.Milliseconds(),
       )
   }

   Call scheduleNightlyAggregation in main() after server starts.
   Run in a goroutine — must not block server startup.
   Cancel job context on SIGTERM.

─────────────────────────────────────────
PART C — Manual Trigger Endpoint
─────────────────────────────────────────

5. POST /api/internal/jobs/aggregate

   Protected by X-Internal-Key header (not JWT):
   — Read expected key from INTERNAL_API_KEY env var
   — If header missing or wrong: return 401
   — Disable in production if ALLOW_INTERNAL_ENDPOINTS=false

   Handler:
   — Trigger RunAll in background goroutine
   — Return 202 Accepted immediately:
     { "message": "Aggregation started", "started_at": "..." }

   Add INTERNAL_API_KEY to .env.example

─────────────────────────────────────────
PART D — Tests
─────────────────────────────────────────

6. internal/jobs/aggregation_test.go

   TestAggregateProfile_CorrectValues
   TestAggregateProfile_Idempotent
   TestAggregateProfile_NoSessions — no panic, NULL/0 snapshot inserted
   TestRunAll_ContinuesOnError — DB error for one profile; others still processed
   TestRunAll_EmptyDatabase — 0 processed, no errors, no panic
```

---

## Issue 26-BE — Analytics API Update

```
Title: Phase 4.2-BE — Analytics overview API with trend data and deltas

Read .github/copilot-instructions.md and .github/data-model.md before starting.
This is the backend half of Issue 26. The dashboard components run in parallel.

─────────────────────────────────────────
PART A — API Update
─────────────────────────────────────────

1. GET /api/analytics/overview (update from Phase 1.5)

   New response shape:
   {
     current: {
       attention_score: number | null,
       memory_score: number | null,
       engagement_frequency: number,
       coop_participation_rate: number,
       avg_reaction_time_ms: number | null,
       total_stars: number,
       total_xp: number,
       sessions_this_week: number,
       snapshot_date: string
     },
     trend: [
       { date: string, attention_score: number | null, memory_score: number | null,
         engagement_frequency: number }
     ],  // last 7 days, oldest first
     deltas: {
       attention_delta: number | null,
       memory_delta: number | null,
       engagement_delta: number
     },
     child_nickname: string,
     child_avatar_id: number,
     data_available: boolean
   }

2. internal/db/analytics_queries.go (update)

   GetLatestSnapshot(ctx, profileID) → AnalyticsSnapshot, error

   GetSnapshotRange(ctx, profileID, days int) → []AnalyticsSnapshot, error
   — Returns last {days} daily snapshots, ordered by date ASC
   — Missing dates: fill with { date: X, attention_score: null, ... }

   GetWeeklyDelta(ctx, profileID) → AnalyticsDeltas, error
   — Compare last 7 days avg vs previous 7 days avg
   — Positive = improving; negative = declining

3. GET /api/analytics/sessions

   Requires parent/educator role
   Returns last 30 sessions grouped by zone:
   {
     by_zone: { memory_cove: number, focus_forest: number, team_tower: number },
     sessions: [{ game_type, stars_earned, created_at }]
   }
   Cache-Control: max-age=900 (15 minutes)
```

---

## Issue 27-BE — PDF Report Generation & API

```
Title: Phase 4.3-BE — Weekly progress report PDF generation and API endpoint

Read .github/copilot-instructions.md before starting.
This is the backend half of Issue 27. The ExportButton frontend component runs in parallel.

─────────────────────────────────────────
PART A — Go PDF Generation
─────────────────────────────────────────

1. Add to go.mod:
   github.com/go-pdf/fpdf v2.7.0

2. internal/reports/weekly_report.go

   type WeeklyReportData struct {
       ChildNickname   string
       WeekStartDate   time.Time
       WeekEndDate     time.Time
       CurrentSnapshot AnalyticsSnapshot
       TrendSnapshots  []AnalyticsSnapshot  // last 30 days
       RecentSessions  []SessionSummary     // last 7 days
       TotalStats      ProfileStats
   }

   GenerateWeeklyReport(ctx context.Context, db *pgxpool.Pool, profileID string) ([]byte, error)

   PAGE 1 — WEEKLY SUMMARY:
   — Header band (warm teal): "Skill Island Progress Report", nickname, date range
   — 4 Metric boxes (2×2 grid): Attention %, Memory %, Social %, Total XP
     Each box: value, trend arrow (↑ ↓ →) + delta, encouraging label
   — Bottom row: sessions count, stars earned, total XP, streak
   — Encouraging sign-off (rotate through 5 messages based on week number)

   PAGE 2 — SESSION DETAIL:
   — Table: Day | Zone | Mode | Stars | Accuracy | Duration
   — Subtotal row per zone
   — If no sessions: "No sessions played this week — check back next week!"

   PAGE 3 — PROGRESS OVERVIEW:
   — Plain English 30-day trend description per metric
   — Next milestone XP info
   — Achievements earned this week
   — Footer: "Skill Island — {date}" + clinical disclaimer on every page

   Language rules enforced throughout:
   — All language positive — no exceptions
   — No "failed", "poor", "low", "declined"
   — Declining metrics → "There's room to grow here 🌱"
   — Colours: amber/green only, never red

3. internal/db/report_queries.go
   FetchWeeklyReportData(ctx, profileID) → WeeklyReportData, error

─────────────────────────────────────────
PART B — API Endpoint
─────────────────────────────────────────

4. GET /api/reports/weekly (internal/api/report_handler.go)

   Requires RequireAuth + parent or educator role

   Flow:
   a. Get profileID linked to this parent (from parent_student_links table)
   b. Call GenerateWeeklyReport(ctx, db, profileID)
   c. On success: return PDF bytes with headers:
      Content-Type: application/pdf
      Content-Disposition: attachment; filename="skillisland-{nickname}-{date}.pdf"
      Cache-Control: no-store
   d. On error: return 500 with generic message
      Never expose internal error details to client

5. Parent → Student relationship:

   Migration: 009_create_parent_links.up.sql

   CREATE TABLE parent_student_links (
     id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     parent_id  UUID NOT NULL REFERENCES users(id),
     student_id UUID NOT NULL REFERENCES student_profiles(id),
     created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
     UNIQUE(parent_id, student_id)
   );

   POST /api/admin/link-student (body: { parent_email, student_nickname })
   Protected by INTERNAL_API_KEY header.
   Only active when ALLOW_INTERNAL_ENDPOINTS=true.
```

---

## Issue 28-BE — Backend Performance Optimisation

```
Title: Phase 4.4-BE — Go backend performance and connection pool optimisation

Read .github/copilot-instructions.md before starting.
This is the backend portion of Issue 28. Frontend Phaser optimisation runs in parallel.

─────────────────────────────────────────
PART A — Database Connection Pool
─────────────────────────────────────────

1. services/api/internal/db/db.go — connection pool config

   pgxpool.ParseConfig additions:
   config.MaxConns         = 20
   config.MinConns         = 2
   config.MaxConnLifetime  = time.Hour
   config.MaxConnIdleTime  = 30 * time.Minute
   config.HealthCheckPeriod = time.Minute

─────────────────────────────────────────
PART B — pprof (dev only)
─────────────────────────────────────────

2. cmd/server/main.go:

   if cfg.Env == "development" {
     go func() {
       log.Println(http.ListenAndServe("localhost:6060", nil))
     }()
     import _ "net/http/pprof"
   }

   Profile POST /api/sessions:
   go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
   Target: < 5ms per session submission (excluding network)

─────────────────────────────────────────
PART C — Performance Test Script
─────────────────────────────────────────

3. scripts/perf-test.sh

   Requires: k6 (brew install k6 or choco install k6)

   a. POST /api/sessions — 100 concurrent users, 30s
      Pass: P95 < 200ms, P99 < 500ms, 0% error rate

   b. GET /api/analytics/overview — 50 concurrent users
      Pass: P95 < 100ms (cached), 0% error rate

   c. WebSocket room — 20 concurrent room pairs
      Pass: state_update latency P95 < 100ms

   Run before marking complete. Paste P95 results as PR comment.
```

---

## Issue 29-BE — Security Audit & Hardening

```
Title: Phase 4.5-BE — Security audit and hardening (platform serves minors)

Read .github/copilot-instructions.md and .github/data-model.md before starting.

─────────────────────────────────────────
PART A — Score Manipulation Tests
─────────────────────────────────────────

1. internal/validator/security_test.go

   TestRejectImplausiblyLargeActionCount
   — Submit 501 actions for Memory Cove → expect 422

   TestRejectOutOfOrderTimestamps
   — ClientTimestamp decreases between actions → clamp to 0, no panic/reject

   TestRejectEmptyActionLog
   — actions: [] → 0 stars, 0 XP, 200 OK (valid zero session)

   TestRejectNegativeReactionTime
   — ClientTimestamp before session start → reaction_time_ms clamped to 0

   TestRejectReplayedSessionToken
   — Submit valid token twice → second → 422 "session_token_already_used"

   TestRejectExpiredSessionToken
   — expires_at in past → 422 "session_token_expired"

   TestRejectClientSubmittedScore
   — Client submits { score: 9999 } → DB stores server-calculated value, not 9999

─────────────────────────────────────────
PART B — Auth Edge Case Tests
─────────────────────────────────────────

2. internal/auth/auth_test.go (additions)

   TestExpiredAccessToken       — exp=1s; wait 2s; use token → 401
   TestTamperedJWTSignature     — modified payload → 401
   TestMissingCookie            — no cookies → 401
   TestRefreshTokenRotation     — reuse refresh token → 401
   TestRoleEnforcement_StudentOnParentEndpoint — GET /api/analytics/overview → 403
   TestRoleEnforcement_ParentOnStudentEndpoint — POST /api/sessions → 403

─────────────────────────────────────────
PART C — WebSocket Security
─────────────────────────────────────────

3. internal/ws/ws_handler.go (updates)

   conn.SetReadLimit(4096)  // 4KB max per WS message
   — On message > 4KB: conn.Close(); log profileID + message size

   Malformed JSON:
   — recover from panic on parse error
   — On parse error: log warning; send action_rejected; do NOT disconnect

   Already-ended game:
   — If room.RoomState == "ENDED" and client sends place_block:
     Return { type: "action_rejected", reason: "game_ended" }

4. internal/ws/security_test.go

   TestWSUpgradeWithoutJWT       — GET /ws/game with no cookie → HTTP 401
   TestWSMaxMessageSize          — send 5KB message → connection closed cleanly
   TestWSMalformedJSON           — "not json" → client receives action_rejected (no disconnect)
   TestWSPlaceBlockAfterSessionEnd — place_block after ENDED → action_rejected "game_ended"

─────────────────────────────────────────
PART D — HTTP Security Headers
─────────────────────────────────────────

5. nginx/nginx.conf (create if not exists)

   add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
   add_header X-Content-Type-Options "nosniff" always;
   add_header X-Frame-Options "DENY" always;
   add_header Referrer-Policy "no-referrer" always;
   add_header Permissions-Policy "camera=(), microphone=(), geolocation=()" always;
   add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self' ws: wss:;" always;

─────────────────────────────────────────
PART E — Dependency Scanning CI
─────────────────────────────────────────

6. .github/workflows/security.yml

   name: Security Scan
   triggers: push and PR to main

   jobs:
   go-vuln:
   — govulncheck ./...
   — Fail on HIGH or CRITICAL

   npm-audit:
   — npm audit --audit-level=high
   — Fail on HIGH or CRITICAL

─────────────────────────────────────────
PART F — Rate Limit Verification
─────────────────────────────────────────

7. scripts/rate-limit-test.sh

   Test 1 — Brute force: 10 rapid POST /api/auth/login → requests 6–10 return 429
   Test 2 — Session limit: 35 rapid POST /api/sessions → requests 31–35 return 429
   Test 3 — WS limit: 5 WS connections for same user → connections 4–5 return 429

   Run script; paste results as PR comment. All 3 must show expected 429s.
```

---

## Issue 30-BE — Slow Mode Migration & API

```
Title: Phase 4.6-BE — Slow Mode database migration and API support

Read .github/copilot-instructions.md before starting.
This is the backend portion of Issue 30 (accessibility). The frontend WCAG audit runs in parallel.

─────────────────────────────────────────
PART A — Migration
─────────────────────────────────────────

1. migrations/009_add_slow_mode.up.sql:
   ALTER TABLE student_profiles ADD COLUMN slow_mode BOOLEAN NOT NULL DEFAULT false;

   migrations/009_add_slow_mode.down.sql:
   ALTER TABLE student_profiles DROP COLUMN slow_mode;

─────────────────────────────────────────
PART B — API Updates
─────────────────────────────────────────

2. PATCH /api/profiles/me — add slow_mode field to update handler
   — Accept { slow_mode: boolean } in request body
   — Validate and save to student_profiles

3. GET /api/sessions/init — return slow_mode in response for memory_cove sessions
   Response additions (game_type = "memory_cove" only):
   {
     ...existing fields...,
     slow_mode: boolean
   }
   Client uses this to configure display timings — never trusts client-sent value.
```

---

## Issue 31-BE — Beta Production Readiness

```
Title: Phase 4.7-BE — Backend production readiness and beta launch

Read .github/copilot-instructions.md before starting.
This is the backend half of Issue 31. Frontend error boundaries run in parallel.

─────────────────────────────────────────
PART A — Production Config Validation
─────────────────────────────────────────

1. cmd/server/main.go — startup validation

   validateConfig(cfg *config.Config) error:
   — Panic with clear message if any required env var is empty:
     DATABASE_URL, JWT_SECRET, JWT_REFRESH_SECRET, ALLOWED_ORIGINS, PORT
   — Panic if JWT_SECRET length < 32 characters
   — Panic if JWT_SECRET == "change-this-in-production-minimum-32-chars"
   — Log all non-secret config values on startup
   — Never log JWT_SECRET or DATABASE_URL

   Dockerfile:
   ARG BUILD_VERSION=dev
   ENV BUILD_VERSION=$BUILD_VERSION
   Include BUILD_VERSION in /health response.

2. Disable pprof in production:
   if cfg.Env == "development" { /* pprof from Issue 28 */ }
   // Never expose in production

3. Gate internal endpoints:
   if cfg.AllowInternalEndpoints {
     r.Post("/api/internal/jobs/aggregate", h.TriggerAggregation)
     r.Post("/api/admin/link-student", h.LinkStudent)
   }
   Add ALLOW_INTERNAL_ENDPOINTS=false to production env.

─────────────────────────────────────────
PART B — Structured Logging
─────────────────────────────────────────

4. services/api — replace all fmt.Println and log.Printf with slog

   In cmd/server/main.go:
   if cfg.Env == "production" {
     logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
   } else {
     logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
   }
   slog.SetDefault(logger)

   Log every request (middleware):
   slog.Info("request", "method", "path", "status", "duration_ms", "profile_id")

   Log every room event:
   slog.Info("room_event", "room_id", "event", "player_count", "outcome")

   Never log: passwords, JWT tokens, full email addresses, raw SQL queries

─────────────────────────────────────────
PART C — Graceful Shutdown
─────────────────────────────────────────

5. cmd/server/main.go — graceful shutdown

   signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

   On signal:
   1. server.Shutdown(ctx with 30s timeout)
   2. hub.Shutdown()  — sends session_end to all clients; writes DB
   3. jobCancel()     — cancel nightly job if running
   4. dbPool.Close()

6. internal/ws/hub.go — Hub.Shutdown():
   — Cancel all room contexts
   — Wait up to 10 seconds for rooms to finish writing DB
   — Force close remaining connections after timeout

─────────────────────────────────────────
PART D — Health Check Upgrade
─────────────────────────────────────────

7. GET /health (update from Phase 0.2)

   Response:
   {
     "status": "ok" | "degraded",
     "db": "ok" | "error",
     "version": "{BUILD_VERSION}",
     "uptime_seconds": number,
     "active_rooms": number,
     "connected_clients": number
   }

   Return HTTP 200 if db == "ok"
   Return HTTP 503 if db == "error"
   DB check: run SELECT 1 with 2-second timeout

─────────────────────────────────────────
PART E — Seed Script & Deploy Workflow
─────────────────────────────────────────

8. scripts/seed-beta.sh

   Creates:
   — 1 educator account: educator@beta.skillisland.com
   — 3 student profiles: Alex (avatar 0), Jordan (avatar 2), Sam (avatar 4)
   — 10 sample game sessions per student across all 3 zones
   — Triggers POST /api/internal/jobs/aggregate to populate snapshots
   — Creates parent_student_links for demonstration

   Usage:
   BASE_URL=http://localhost:8080 INTERNAL_API_KEY=xxx bash scripts/seed-beta.sh

9. .github/workflows/deploy.yml

   Trigger: push to main (after all CI + security checks pass)

   Steps:
   a. go test -race ./...
   b. govulncheck ./... + npm audit
   c. Build Go binary with BUILD_VERSION=${{ github.sha }}
   d. npm run build
   e. docker build -t skillisland-api:$SHA -t skillisland-web:$SHA
   f. Push to GitHub Container Registry
   g. SSH to VPS: docker compose pull; docker compose up -d --no-deps api web
   h. Health check: curl -f https://skillisland.example.com/health (retry 10×)
   i. Notify on success/failure

   Rollback: if health check fails → deploy previous image tag

─────────────────────────────────────────
PART F — Backend PR Checklist
─────────────────────────────────────────

Before this can be completed:

- [ ] All required env vars validated on startup; server panics clearly on missing vars
- [ ] pprof disabled in production (cfg.Env check confirmed)
- [ ] Internal endpoints gated behind ALLOW_INTERNAL_ENDPOINTS flag
- [ ] Graceful shutdown tested: SIGTERM closes rooms + writes DB cleanly
- [ ] Health check returns correct db status and BUILD_VERSION
- [ ] Structured JSON logging confirmed in production mode
- [ ] go test -race ./... — all pass
- [ ] Security workflow — zero HIGH/CRITICAL vulnerabilities
- [ ] Rate limit test — 429s confirmed
- [ ] Deploy workflow runs successfully on merge to main
- [ ] Health check passes after deploy
- [ ] Seed script runs successfully against beta environment
```

---
