# Skill Island — Game Logic (Deferred Zones)

> Pattern Plateau and Community Hub.
> FULLY SPECIFIED — but do not build for MVP.
> These specs exist so Copilot understands the full system and island map
> is built correctly. Only build after MVP beta with schools is complete.

---

## Game Logic — Pattern Plateau ⏸ DEFERRED

> Island map card: visible with padlock + "Coming Soon". Never clickable in MVP.
> Database enum: include "pattern_plateau" in game_type from day one.
> Do not create the Phaser scene or validator until unblocked.

### Overview
Solo pattern recognition game. Server generates number or shape sequences with one
missing element. Player picks the correct answer from multiple-choice options.
Sequences increase in complexity as the player progresses.

### Files (create when unblocked)
```
apps/web/game/scenes/PatternPlateauScene.ts    → Phaser scene
apps/web/game/logic/patternPlateau.ts          → Display helpers only
services/api/internal/validator/pattern.go     → Answer validation + scoring
```

### Pattern Types
```typescript
type PatternType =
  | "number_sequence"      // 4, 7, 8, 9, 10, 11, 12, [?] → 13
  | "arithmetic_sequence"  // 2, 4, 8, 16, [?]             → 32
  | "shape_sequence"       // circle, square, triangle, [?] → circle (repeating)
  | "colour_pattern"       // red, blue, red, blue, [?]     → red
```

### Difficulty Levels
```
Level 1: number_sequence only; gap=1; 4 elements; 3 choices
Level 2: number_sequence; gap=2 or 3; 5 elements; 3 choices
Level 3: arithmetic_sequence; multiply/divide; 5 elements; 4 choices
Level 4: mixed types; 6 elements; 4 choices; two missing elements
Level 5: mixed + shape_sequence; 7 elements; 4 choices; two missing
```

### Client Action Format
```typescript
interface PatternPlateauAction {
  type: "answer_selected"
  challenge_id: string       // Server-generated ID for this challenge
  selected_answer: string    // Value the player selected
  client_timestamp: number   // ms since session start
}
```

### Server Validation (`internal/validator/pattern.go`)
```go
// Per action:
// 1. Regenerate expected answer from challenge_id + session seed
// 2. Compare selected_answer to expected
// 3. Record correct: true/false
// 4. reaction_time_ms = ClientTimestamp - challengeShownAt
// 5. Allow one retry per challenge (SEND-friendly)
//    On retry: increment retry_count; record second attempt separately

// Score (per session):
// score = (correct_first_attempt / total_challenges) * 1000
// 3 stars → first-attempt accuracy >= 0.85
// 2 stars → first-attempt accuracy >= 0.65
// 1 star  → first-attempt accuracy >= 0.40
// 0 stars → accuracy < 0.40
// xp_earned = stars * 10 + challenges_completed * 8
```

### Phaser Scene State Machine
```
IDLE → SHOWING_CHALLENGE → WAITING_FOR_ANSWER → FEEDBACK → NEXT_CHALLENGE → SESSION_END
```

```typescript
// SHOWING_CHALLENGE:   Pattern grid with missing slot (star icon)
//                      Answer pill buttons (3 or 4 options)
// WAITING_FOR_ANSWER:  Player selects pill
// FEEDBACK:
//   Correct → green sparkle on pill; 800ms celebration pause
//   Wrong   → gentle shake on pill; "Try again?" appears (one retry allowed)
//   Retry wrong → reveal correct answer with soft highlight; move on
// NEXT_CHALLENGE: brief transition; next pattern

const CHALLENGE_DISPLAY_MS = 500  // Delay before answer buttons activate
const FEEDBACK_CORRECT_MS  = 800
const FEEDBACK_WRONG_MS    = 600
```

### Behavioural Metrics
```
reaction_time_ms  → challenge display to first answer selection
hesitation_ms     → time between buttons active and first tap
retry_count       → 0 (first attempt) or 1 (needed retry)
correct           → true if correct on first attempt
metadata JSONB    → { pattern_type, difficulty_level, challenge_id, selected, expected }
```

---

## Game Logic — Community Hub ⏸ DEFERRED

> Island map card: visible with padlock + "Coming Soon". Never clickable in MVP.
> Do not create WS hub room, hub_presence_log table, or Phaser scene until unblocked.
> Requires dedicated SEND safety design review before implementation.

### Overview
Shared social presence space. Students move avatars around a village environment.
No free-form communication of any kind. Social interaction is strictly limited to:
avatar proximity, shared Daily Group Missions, and Team Achievements.

### Non-Negotiable Safety Constraints
```
NO free-form text chat
NO voice communication
NO direct player-to-player messages
NO full usernames visible to others (first initial only)
NO player count display (avoids social anxiety)
Interaction ONLY via: avatar movement, mission contribution, proximity emotes
```

### Public Avatar State
```typescript
// What other clients receive — profile_id is NEVER broadcast to others
interface PublicAvatarState {
  avatar_id: number
  display_label: string  // first initial only e.g. "A"
  x: number              // 0.0–1.0 normalised
  y: number              // 0.0–1.0 normalised
  emote: string | null   // "wave" | "jump" | "dance" | null
}
```

### Daily Group Mission
```typescript
type MissionType =
  | "complete_n_puzzles"     // counts Memory Cove + Pattern Plateau sessions
  | "complete_n_activities"  // counts any game session
  | "team_n_sessions"        // counts Team Tower sessions

interface DailyMission {
  id: string
  date: string             // YYYY-MM-DD
  mission_type: MissionType
  goal_count: number
  current_progress: number // community-wide; individual contributions NOT shown
  completed: boolean
}
// Progress bar = current_progress / goal_count
// Individual contributions hidden — prevents comparison anxiety
// When completed: celebration animation for all connected players
```

### Server Presence Architecture
```go
// Hub room: persistent; no player cap (unlike Team Tower)
// One hub room per day; resets at midnight UTC
// Tick rate: 5 Hz (200ms) — movement is slow and decorative
// Broadcast: PublicAvatarState[] for all connected players
// Max rendered per client: 20 nearest avatars

// Does NOT write to game_sessions
// Writes to hub_presence_log (engagement analytics only)
```

### Hub Presence Table (add to schema when building)
```sql
hub_presence_log (
  id               UUID PK,
  profile_id       UUID FK → student_profiles.id,
  session_date     DATE,
  duration_seconds INT,
  emotes_used      INT,
  created_at       TIMESTAMPTZ
)
```

### Team Achievements
```typescript
type AchievementType =
  | "mission_complete"   // contributed to completed Daily Group Mission
  | "great_job"          // 3-star session in any zone
  | "top_team"           // won Team Tower with 3 stars
  | "island_explorers"   // played all 3 MVP zones at least once
  | "treasure_found"     // reached 100 total stars
  | "skills_unlocked"    // unlocked 3 zones
  | "island_success"     // completed 20 sessions total
  | "quest_completed"    // completed a full week of daily missions

// Achievements checked and awarded server-side after every session write
// Displayed as badge icons in Community Hub and on the leaderboard screen
```

### Avatar Movement
```typescript
// Click-to-move only (not WASD — accessibility rule)
// Walking speed: 80px/s normalised — slow and calm; no running

interface HubMoveAction {
  type: "move"
  target_x: number  // 0.0–1.0
  target_y: number  // 0.0–1.0
}

// Client interpolates own avatar immediately (optimistic)
// Other avatars lerped on state_update from server

const HUB_MOVE_SPEED         = 80    // px/sec normalised
const EMOTE_DURATION_MS      = 2000
const EMOTE_BROADCAST_RADIUS = 0.3   // emotes only to nearby players

// Emotes: wave | jump | dance — all positive, SEND-safe
// Emote button cycles through 3 options; duration 2s then clears
```

### Mission Progress API
```typescript
// GET /api/missions/today → DailyMission
// Progress auto-updated server-side when qualifying game_sessions written
// No separate client call needed for progress increment

// WS broadcast to hub room on progress change:
{ type: "mission_progress", current_progress: number, goal_count: number, completed: boolean }
```

### EventBus (deferred — do not implement in MVP)
```typescript
// Add these ONLY when building Community Hub:
EventBus.emit('game:hub-state', {
  avatars: PublicAvatarState[],
  mission: DailyMission
})
EventBus.emit('game:mission-progress', {
  current_progress: number,
  completed: boolean
})
```

---

*v2.2 — Deferred zone specs. Build only after MVP beta is stable.*
