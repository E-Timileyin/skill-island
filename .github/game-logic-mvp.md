# Skill Island — Game Logic (MVP Zones)

> Memory Cove, Focus Forest, Team Tower.
> These are the only zones to build for MVP.
> For deferred zones see game-logic-deferred.md

---

## Game Logic — Memory Cove

### Overview
Solo sequence-memory game. Server generates a sequence of coloured shapes, displays
them one by one, hides them. Player repeats the sequence via button presses. Server
validates every action and calculates the final score.

### Files
```
apps/web/game/scenes/MemoryCoveScene.ts    → Phaser scene
apps/web/game/logic/memoryCove.ts          → Display helpers only (no scoring)
services/api/internal/validator/memory.go  → Sequence validation + scoring
services/api/internal/api/sessions.go      → POST /api/sessions handler
```

### Sequence Generation (Server)
```go
// Shapes:  "circle" | "square" | "triangle" | "star"
// Colours: "red" | "blue" | "green" | "yellow"
// Each element = { Shape, Colour } pair e.g. { "circle", "red" }

type SequenceElement struct {
    Shape  string `json:"shape"`
    Colour string `json:"colour"`
}

// Server uses seeded random — seed stored with session.
// On validation: regenerate expected sequence from seed.
// NEVER trust client to send sequence back.
func GenerateSequence(seed int64, length int) []SequenceElement
```

### Difficulty Scaling (Server)
```
Rounds 1–2:   length = 3
Rounds 3–4:   length = 4
Rounds 5–6:   length = 5
Rounds 7–8:   length = 6
Rounds 9–10:  length = 7
Rounds 11+:   length = 8 (max)

Scale UP only if accuracy >= 0.80 over last 3 rounds.
Scale DOWN if accuracy < 0.50 over last 2 rounds (min = 3).
Never change length mid-round.
```

### Client Action Format
```typescript
interface MemoryCoveAction {
  type: "button_press"
  element_index: number     // 0-based position in sequence
  button_id: string         // e.g. "circle_red"
  client_timestamp: number  // ms since session start
}

interface SessionEndSignal {
  type: "session_end"
  rounds_completed: number
  total_actions: number
}
```

### Server Validation (`internal/validator/memory.go`)
```go
// Per action:
// 1. Regenerate expected element at ElementIndex from seed
// 2. Compare ButtonID to expected — record correct: true/false
// 3. reaction_time_ms = ClientTimestamp - stimulusShownAt
// 4. Write behavioral_metrics row immediately

// Score at session end:
// score     = (correct_count / total_actions) * 1000
// 3 stars   → accuracy >= 0.90
// 2 stars   → accuracy >= 0.70
// 1 star    → accuracy >= 0.50
// 0 stars   → accuracy <  0.50
// xp_earned = stars * 10 + rounds_completed * 5
```

### Phaser Scene State Machine
```
IDLE → SHOWING_SEQUENCE → WAITING_FOR_INPUT → CHECKING → ROUND_COMPLETE → GAME_OVER
```

```typescript
// IDLE:             "Watch carefully!" prompt
// SHOWING_SEQUENCE: Each element shown for 800ms; 300ms gap between
// WAITING_FOR_INPUT: Buttons active; player taps in order
// CHECKING:         Emit player_action per tap; await next
// ROUND_COMPLETE:   Sparkle (correct) / shake (wrong); 600ms pause; next round
// GAME_OVER:        Emit 'game:session-end'; show stars; disable input

const ELEMENT_DISPLAY_MS = 800
const ELEMENT_GAP_MS     = 300
const FEEDBACK_PAUSE_MS  = 600

// SEND rule: wrong answer does NOT end session.
// Mark incorrect, continue — no sudden failure screens.
```

### Behavioural Metrics
```
reaction_time_ms  → input-phase-start to button press
hesitation_ms     → pause before first press per round
retry_count       → always 0 (no retries in Memory Cove)
correct           → button matched expected element
metadata JSONB    → { sequence_length, round_number, element_index, expected, actual }
```

### EventBus Emission
```typescript
EventBus.emit('game:session-end', {
  game_type: 'memory_cove',
  actions: MemoryCoveAction[],
  rounds_completed: number,
  total_actions: number,
  duration_ms: number
})
// Next.js catches this and POSTs to /api/sessions
```

---

## Game Logic — Focus Forest

### Overview
Solo sustained-attention game. Butterflies and bees move across a forest background.
Tap butterflies; avoid bees. Fixed 60-second session. Server pre-generates a spawn
manifest from a seed and validates taps against it.

### Files
```
apps/web/game/scenes/FocusForestScene.ts    → Phaser scene
apps/web/game/logic/spawnPatterns.ts        → Client spawn display (not validation)
services/api/internal/validator/focus.go    → Tap validation + attention score
```

### Target Types
```typescript
type TargetType = "butterfly_blue" | "butterfly_orange" | "butterfly_red" | "bee"
// Butterfly tap = correct (+1)
// Bee tap       = distraction (recorded; NO score penalty — SEND-friendly)
// Missed target = recorded as missed; no penalty
```

### Session Structure
```
Duration: 60 seconds (fixed)

Difficulty (auto-scales between sessions, not within):
  Level 1: 30% bees, speed 80px/s,  spawn every 1200ms
  Level 2: 40% bees, speed 100px/s, spawn every 1000ms
  Level 3: 50% bees, speed 130px/s, spawn every 800ms
  Level 4: 55% bees, speed 160px/s, spawn every 700ms (max)
```

### Server Spawn Manifest
```go
type SpawnEvent struct {
    TargetID    string  `json:"target_id"`
    TargetType  string  `json:"target_type"`    // "butterfly_*" or "bee"
    SpawnTimeMs int     `json:"spawn_time_ms"`  // ms from session start
    PositionX   float64 `json:"position_x"`     // 0.0–1.0 normalised
    PositionY   float64 `json:"position_y"`     // 0.0–1.0 normalised
}
// Seed stored with session; server regenerates manifest to validate taps.
// Client receives seed at session start to display identical spawn positions.
```

### Client Action Format
```typescript
interface FocusForestAction {
  type: "tap"
  tap_x: number            // 0.0–1.0 normalised
  tap_y: number            // 0.0–1.0 normalised
  client_timestamp: number // ms since session start
}
// Client does NOT say what it tapped — server determines from manifest.
// Hit radius: 0.08 normalised units (server-defined).
```

### Server Validation (`internal/validator/focus.go`)
```go
// Per tap:
// 1. Find active targets at action.ClientTimestamp (spawned, not despawned)
// 2. Lerp each target's position to current timestamp
// 3. Find nearest target within hit_radius (0.08)
// 4. butterfly → correct: true | bee → correct: false (distraction)
// 5. reaction_time_ms = tap_timestamp - target.SpawnTimeMs

// Score:
// attention_score = (butterfly_hits - bee_hits * 0.5) / butterflies_total
//                   clamped [0.0, 1.0]
// 3 stars → attention_score >= 0.85
// 2 stars → attention_score >= 0.65
// 1 star  → attention_score >= 0.40
// 0 stars → attention_score <  0.40
// xp_earned = stars * 10 + floor(attention_score * 50)
```

### Phaser Scene Behaviour
```typescript
// Targets: Phaser Image + Tween (NOT physics velocity)
// Tween duration = screen_width / speed_px_per_s * 1000

// On tap:
// 1. Record position + timestamp; emit player_action immediately
// 2. Butterfly hit → sparkle + fade tween
// 3. Bee hit       → shake tween + buzz sound
// 4. No "Wrong!" text — visual-only feedback (SEND rule)

// Spawn: time.addEvent with delay = spawnIntervalMs
// Client spawn is display-only — server regenerates from seed for truth

const DESPAWN_AFTER_MS = 3000  // Remove if not tapped within 3s

// Session timer: countdown bar top of screen
// At 0 → emit 'game:session-end' with full action log
```

### Behavioural Metrics
```
reaction_time_ms  → target spawn to tap
hesitation_ms     → null (not applicable)
retry_count       → 0
correct           → true = butterfly hit, false = bee/miss
metadata JSONB    → { target_type, difficulty_level, was_miss: bool }
```

### Attention Score in Dashboard
```
Derived from Focus Forest sessions.
7-day rolling average of per-session attention_score.
1.0 = perfect butterfly accuracy, zero bee distractions.
Display as: (attention_score * 100).toFixed(0) + "%"
```

---

## Game Logic — Team Tower (2-Player Co-op)

### Overview
Real-time 2-player co-op. Players alternate placing falling blocks to build a shared
tower. Server is authoritative for all physics and placement. Clients render server
state using interpolation.

### Files
```
apps/web/game/scenes/TeamTowerScene.ts           → Phaser scene
apps/web/game/logic/teamTower.ts                 → Client interpolation helpers
services/api/internal/ws/rooms/teamtower.go      → Server game logic
services/api/internal/ws/hub.go                  → Room lifecycle
services/api/internal/ws/broadcast.go            → Tick broadcast loop
```

### Game Structure
```
Players:        Exactly 2 (enforced by hub — not just the handler)
Turn order:     Alternate; Player 1 places first
Win condition:  Tower height >= target (10 blocks, level 1)
Lose condition: Any block tilts > 45 degrees (CoM deviation > 0.35)
Session end:    Win | Lose | Disconnect without reconnect in 30s window
```

### Block Types
```typescript
type BlockShape = "1x1" | "2x1" | "1x2" | "L_shape" | "T_shape"
// Colours: blue, red, green, orange, grey — cosmetic only
// Shape sequence: server-seeded; next shape shown to both players before placement
```

### Server Game State
```go
type Block struct {
    ID       string  `json:"id"`
    Shape    string  `json:"shape"`
    Colour   string  `json:"colour"`
    X        float64 `json:"x"`        // centre, normalised 0.0–1.0
    Y        float64 `json:"y"`        // from tower bottom
    Rotation float64 `json:"rotation"` // degrees; 0 = upright
    PlacedBy string  `json:"placed_by"` // "player_1" | "player_2"
}

type TowerState struct {
    Blocks          []Block `json:"blocks"`
    CurrentHeight   float64 `json:"current_height"`
    TargetHeight    float64 `json:"target_height"`
    ActivePlayer    string  `json:"active_player"`
    NextBlockShape  string  `json:"next_block_shape"`
    Stable          bool    `json:"stable"`   // false = tower fallen
    GroupXP         int     `json:"group_xp"`
    TurnNumber      int     `json:"turn_number"`
}
```

### Client Action Format
```typescript
interface TeamTowerAction {
  type: "place_block"
  position_x: number        // 0.0–1.0; where player drops block
  client_timestamp: number
}
// Client sends ONLY position_x.
// Server determines Y, validates turn, checks physics, updates state.
```

### Server Physics (`teamtower.go`)
```go
// On place_block:
// 1. Validate it is this player's turn — reject if not
// 2. Validate position_x in [0.05, 0.95] — reject if not
// 3. Calculate landing Y (stack on highest block at position_x)
// 4. Centre-of-mass check:
//    CoM = sum(block.X * weight) / sum(weight)
//    If abs(CoM - 0.5) > 0.35 → tower falls → Stable = false
// 5. Win check: current_height >= target_height
// 6. Award +5 group_xp per placement
// 7. Switch active_player; select next block from seed
// 8. Broadcast updated TowerState on next tick

// Block weights:
// "1x1" = 1.0 | "2x1" = 2.0 | "1x2" = 1.5 | "L_shape" = 2.5 | "T_shape" = 2.5
```

### WebSocket Messages (Team Tower)
```typescript
// Client → Server
{ type: "place_block", position_x: 0.47, client_timestamp: 12450 }

// Server → Client (tick broadcast)
{ type: "state_update", tick: 142, game_state: TowerState, server_timestamp: number }

// Server → Client (validation error)
{ type: "action_rejected", reason: "not_your_turn" | "out_of_bounds" | "game_ended" }

// Server → Client (session end)
{
  type: "session_end",
  outcome: "win" | "lose" | "incomplete",
  group_xp: number,
  stars: number,
  final_state: TowerState,
  room_session_id: string
}
```

### Broadcast Tick
```go
// Tick rate: 10–12 Hz → time.NewTicker(90 * time.Millisecond)
// Never broadcast on every client message — tick only
// Delta broadcast if full state > 512 bytes
// Always full state on: round start, reconnect, turn change
```

### Client Rendering (`TeamTowerScene.ts`)
```typescript
const LERP_FACTOR = 0.25  // Smooth between server state updates

// On state_update:
// - New block: spawn at received position
// - Existing: lerp to received position over 90ms
// - stable === false: tower fall animation → emit session_end
// - height >= target: win animation → emit session_end

// Input: only accept tap when state.active_player === this client's role
// Show "Your Turn" / "Partner's Turn" at ALL times
// Disable tap when not this player's turn — no frustration (SEND rule)

// Visual:
// Good placement → gentle bounce tween
// Tower wobble   → subtle camera shake when CoM near limit
// Fall           → slow dramatic lean, NOT sudden flash (SEND rule)
```

### Stars & Score
```go
// Stars (session end):
// Win + height >= 110% target  → 3 stars
// Win                          → 2 stars
// Incomplete (disconnect)      → 1 star (effort reward — SEND rule)
// Tower fall                   → 1 star (never 0 stars in co-op)

// XP: +5 per placement; +20 bonus on win; split equally between players

// Write room_sessions: completed = (outcome == "win" || outcome == "lose")
// Write 2 game_sessions rows (one per player); mode = "cooperative"
```

### Behavioural Metrics
```
reaction_time_ms  → "Your Turn" shown to place_block action
hesitation_ms     → time without action after turn became active (> 3s threshold)
retry_count       → 0 (placement is final)
correct           → true if placement did not cause fall
metadata JSONB    → { position_x, turn_number, tower_height_at_placement, block_shape }
```

---

## Shared Rules — All MVP Zones

### XP Thresholds
```
Memory Cove:     0 XP  — BUILD NOW
Focus Forest:   30 XP  — BUILD NOW
Team Tower:     80 XP  — BUILD NOW
Pattern Plateau: 150 XP — DEFERRED
Community Hub:  250 XP  — DEFERRED
```

### Star → XP Conversion
```
0 stars → 0 XP
1 star  → 10 XP
2 stars → 20 XP
3 stars → 35 XP  (mastery bonus)
```

### XP Unlock Check (server, after every session write)
```go
// Fetch updated total_xp; check against thresholds.
// Return unlocked_zones[] in session response if new zone unlocked.
// Client triggers unlock animation on island map.
```

### Feedback Language (apply to every zone — no exceptions)
```
NEVER:  "Wrong!" | "Failed" | "Game Over" | "You lost"
ALWAYS: "Nice try!" | "Almost!" | "Keep going!" | "Great effort!"
3 stars: "Amazing!" | "You nailed it!" | "Perfect!"
Co-op disconnect: "Your partner will be right back..." (never blame)
```

### Session POST endpoint
```typescript
// POST /api/sessions — called by Next.js on 'game:session-end'
interface SessionSubmission {
  game_type: "memory_cove" | "focus_forest" | "team_tower" | "pattern_plateau"
  mode: "solo" | "cooperative"
  actions: Action[]
  duration_ms: number
  room_session_id?: string  // required if cooperative
}

interface SessionResult {
  score: number
  accuracy: number
  stars_earned: number
  xp_earned: number
  total_xp: number
  unlocked_zones: string[]
  behavioral_metrics_count: number
}
```

### Phaser EventBus (all scenes)
```typescript
// apps/web/game/events/EventBus.ts — single shared EventEmitter

// Game → Next.js:
EventBus.emit('game:ready',       { scene: string })
EventBus.emit('game:action',      { action: Action })      // debug only
EventBus.emit('game:session-end', SessionSubmission)

// Next.js → Game:
EventBus.emit('game:profile-loaded', { nickname: string, avatar_id: number })
EventBus.emit('game:ws-state',       { state: TowerState }) // Team Tower only

// Deferred (do not implement in MVP):
// EventBus.emit('game:hub-state', ...)
// EventBus.emit('game:mission-progress', ...)
```

---

*v2.2 — MVP game logic. See game-logic-deferred.md for Pattern Plateau and Community Hub.*
