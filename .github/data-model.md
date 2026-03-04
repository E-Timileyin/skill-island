# Skill Island — Data Model & Real-Time Protocol

> PostgreSQL schema, WebSocket spec, and security detail.
> Core rules and conventions: see copilot-instructions.md

---

## Database Schema

### Naming Conventions
- Tables: `snake_case` plural
- Columns: `snake_case`
- PKs: `UUID` via `gen_random_uuid()`
- FKs: `{table_singular}_id` e.g. `profile_id`
- Timestamps: always `TIMESTAMPTZ`, never `TIMESTAMP`

---

### Table: users
```sql
CREATE TABLE users (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email          TEXT UNIQUE NOT NULL,
  role           TEXT NOT NULL CHECK (role IN ('student', 'parent', 'educator')),
  password_hash  TEXT NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_login_at  TIMESTAMPTZ
);
-- password_hash NEVER returned in any API response
-- email NEVER exposed to student-role sessions
```

### Table: student_profiles
```sql
CREATE TABLE student_profiles (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  nickname     TEXT NOT NULL,
  avatar_id    INTEGER NOT NULL DEFAULT 0,
  total_stars  INTEGER NOT NULL DEFAULT 0,
  total_xp     INTEGER NOT NULL DEFAULT 0,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id)  -- one profile per student account
);
```

### Table: game_sessions
```sql
CREATE TYPE game_zone AS ENUM (
  'memory_cove', 'focus_forest', 'team_tower', 'pattern_plateau'
);
-- Note: community_hub uses hub_presence_log, not game_sessions

CREATE TYPE play_mode AS ENUM ('solo', 'cooperative');

CREATE TABLE game_sessions (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  profile_id       UUID NOT NULL REFERENCES student_profiles(id),
  game_type        game_zone NOT NULL,
  mode             play_mode NOT NULL DEFAULT 'solo',
  score            INTEGER NOT NULL DEFAULT 0,
  duration_seconds INTEGER NOT NULL,
  accuracy         FLOAT NOT NULL DEFAULT 0.0,
  stars_earned     INTEGER NOT NULL DEFAULT 0 CHECK (stars_earned BETWEEN 0 AND 3),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_game_sessions_profile ON game_sessions(profile_id);
CREATE INDEX idx_game_sessions_created ON game_sessions(created_at);
```

### Table: room_sessions
```sql
-- Tracks every co-op room — write on room close, always, no exceptions
CREATE TABLE room_sessions (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  game_type             game_zone NOT NULL,
  player_1_profile_id   UUID NOT NULL REFERENCES student_profiles(id),
  player_2_profile_id   UUID NOT NULL REFERENCES student_profiles(id),
  group_xp_earned       INTEGER NOT NULL DEFAULT 0,
  completed             BOOLEAN NOT NULL DEFAULT false,
  started_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at              TIMESTAMPTZ,
  disconnect_reason     TEXT  -- 'idle_timeout' | 'player_left' | 'error' | NULL
);
```

### Table: behavioral_metrics
```sql
-- Collect from day one — future adaptive AI depends on this data
CREATE TYPE behavior_event AS ENUM ('action', 'hesitation', 'retry', 'distraction');

CREATE TABLE behavioral_metrics (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id          UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
  event_type          behavior_event NOT NULL,
  reaction_time_ms    INTEGER,      -- ms from stimulus to action; NULL if not applicable
  hesitation_ms       INTEGER,      -- ms of pause before acting; NULL if not applicable
  retry_count         INTEGER NOT NULL DEFAULT 0,
  correct             BOOLEAN NOT NULL,
  timestamp_offset_ms INTEGER NOT NULL,  -- ms since session start
  metadata            JSONB         -- game-specific context
);
CREATE INDEX idx_behavioral_session ON behavioral_metrics(session_id);

-- CRITICAL: always write behavioral_metrics atomically with game_sessions
-- Use a single transaction — never write a session without its events
```

### Table: analytics_snapshots
```sql
-- Written by nightly aggregation job — dashboard reads only this table
CREATE TABLE analytics_snapshots (
  profile_id              UUID NOT NULL REFERENCES student_profiles(id),
  snapshot_date           DATE NOT NULL,
  attention_score         FLOAT,   -- 7-day rolling avg (Focus Forest)
  memory_score            FLOAT,   -- 7-day rolling avg accuracy (Memory Cove)
  engagement_frequency    INTEGER, -- session count for period
  coop_participation_rate FLOAT,   -- coop sessions / total sessions
  avg_reaction_time_ms    FLOAT,   -- 7-day avg from behavioral_metrics
  avg_hesitation_ms       FLOAT,   -- 7-day avg; cognitive load proxy
  retry_rate              FLOAT,   -- retries per challenge
  PRIMARY KEY (profile_id, snapshot_date)
);
```

### Table: achievements
```sql
CREATE TABLE achievements (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  profile_id   UUID NOT NULL REFERENCES student_profiles(id),
  type         TEXT NOT NULL,  -- see AchievementType in game-logic-deferred.md
  earned_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_achievements_profile ON achievements(profile_id);
```

---

## Key Data Rules

```go
// 1. Write behavioral_metrics + game_sessions in one transaction
// 2. Never return password_hash or refresh_token in any response
// 3. analytics_snapshots: never query raw session tables from dashboard
// 4. behavioral_metrics: partition by month at scale (post-MVP)
// 5. Soft delete only — never hard delete student data
//    Add deleted_at TIMESTAMPTZ NULL to users; filter WHERE deleted_at IS NULL
```

---

## WebSocket — Full Protocol

### Architecture Rules
- Server is **single source of truth** for all game state
- Client sends **intentions** (actions), never state
- Server validates → applies → broadcasts delta
- Client **interpolates only** — never predicts physics
- Tick rate: **10–12 Hz** for Team Tower (`time.NewTicker(90 * time.Millisecond)`)
- Community Hub: **5 Hz** (200ms) — movement is decorative

### Connection Lifecycle
```
1. Client requests WS upgrade to /ws/game
2. Server extracts JWT from HTTP-only cookie
3. Server validates JWT — reject with HTTP 401 if invalid (not a WS error)
4. Server creates client connection; assigns to room or matchmaking queue
5. Server sends room_ready when both players connected
6. Game loop begins; ticks broadcast on ticker
7. On disconnect: room enters PAUSED; 30s reconnect window
8. On reconnect: full state snapshot sent (not delta)
9. On room end: session written; goroutine exits cleanly
```

### Message Schemas

#### Client → Server
```typescript
// Join matchmaking
{ type: "join_room", game_type: string }

// Game action (all zones)
{ type: "player_action", payload: Action, client_timestamp: number }

// Heartbeat
{ type: "heartbeat_ping", timestamp: number }

// Intentional leave
{ type: "leave_room" }
```

#### Server → Client
```typescript
// Both players connected
{
  type: "room_ready",
  room_id: string,
  player_role: "player_1" | "player_2",
  opponent_avatar: number
}

// Authoritative state (every tick)
{
  type: "state_update",
  tick: number,
  game_state: TowerState,  // or zone-specific state
  server_timestamp: number
}

// Partner disconnected
{
  type: "player_disconnected",
  player_role: "player_1" | "player_2",
  reason: string,
  reconnect_window_seconds: 30
}

// Session complete
{
  type: "session_end",
  outcome: "win" | "lose" | "incomplete",
  group_xp: number,
  stars: number,
  final_state: object,
  room_session_id: string
}

// Heartbeat response
{ type: "heartbeat_pong", timestamp: number, server_time: number }

// Action rejected
{ type: "action_rejected", reason: string }
```

### Room Lifecycle State Machine
```
WAITING  → player 1 connects; awaiting player 2
READY    → both connected; room_ready sent to both clients
PLAYING  → active game; ticks broadcasting at 10-12 Hz
PAUSED   → one player disconnected; 30s reconnect window open
ENDED    → session_end sent; room_sessions written; goroutine exits
```

### Heartbeat & Timeouts
```go
const (
    HeartbeatInterval   = 5  * time.Second  // client pings every 5s
    StaleThreshold      = 15 * time.Second  // server closes stale if no ping in 15s
    ReconnectWindow     = 30 * time.Second  // hold room after disconnect
    IdleWarnAt          = 90 * time.Second  // no actions for 90s → warn
    IdleCloseAt         = 120 * time.Second // no actions for 120s → close room
)

// On idle close: record room_sessions with disconnect_reason = "idle_timeout"
// On reconnect: send FULL state snapshot (not delta) to resync client
```

### Go Room Implementation Notes
```go
// Each room runs in a dedicated goroutine
// Use context.WithCancel for clean shutdown — propagates to all sub-goroutines
// Run race detector in CI: go test -race ./...
// Enforce 2-player max in the hub register function, not just the HTTP handler

// Room map in hub: map[string]*Room — protected by sync.RWMutex
// Player active room index: map[profileID]roomID — enforce one room per profile

// On ENDED:
// 1. Write room_sessions row (always — completed or not)
// 2. Write game_sessions rows for both players
// 3. Write behavioral_metrics rows (in same transaction as game_sessions)
// 4. Release room from hub map
// 5. Cancel room context → goroutine exits
```

---

## Analytics Aggregation

### Nightly Job (runs at 00:05 UTC)
```sql
-- 7-day attention score per profile
INSERT INTO analytics_snapshots (profile_id, snapshot_date, attention_score, ...)
SELECT
  gs.profile_id,
  CURRENT_DATE,
  AVG(gs.accuracy) FILTER (WHERE gs.game_type = 'focus_forest'),
  AVG(gs.accuracy) FILTER (WHERE gs.game_type = 'memory_cove'),
  COUNT(*) as engagement_frequency,
  COUNT(*) FILTER (WHERE gs.mode = 'cooperative')::float / NULLIF(COUNT(*), 0),
  AVG(bm.reaction_time_ms),
  AVG(bm.hesitation_ms),
  AVG(bm.retry_count)
FROM game_sessions gs
LEFT JOIN behavioral_metrics bm ON bm.session_id = gs.id
WHERE gs.created_at >= now() - interval '7 days'
GROUP BY gs.profile_id
ON CONFLICT (profile_id, snapshot_date) DO UPDATE SET
  attention_score = EXCLUDED.attention_score,
  -- ... other fields
  ;
```

### Dashboard API Rules
```go
// GET /api/analytics/overview — reads analytics_snapshots ONLY
// Never run live aggregation queries from dashboard endpoints
// Cache responses: 15-minute TTL; invalidate after nightly job completes
// Parent role: sees own child's data only (filter by parent → child relationship)
// Educator role: sees all profiles in their linked school (anonymised by default)
```

---

## Security Detail

### JWT Structure
```go
type Claims struct {
    ProfileID string `json:"profile_id,omitempty"` // students only
    Role      string `json:"role"`                  // student|parent|educator
    jwt.RegisteredClaims
}
// Access token:  exp = 1 hour
// Refresh token: exp = 7 days; jti stored in DB for rotation check
```

### Cookie Configuration
```go
http.SetCookie(w, &http.Cookie{
    Name:     "access_token",
    Value:    tokenString,
    HttpOnly: true,        // JS cannot read
    Secure:   true,        // HTTPS only
    SameSite: http.SameSiteStrictMode,
    Path:     "/",
    MaxAge:   3600,        // 1 hour
})
```

### Server Score Validation Pattern
```go
// internal/validator interface — implement per game type
type GameValidator interface {
    ValidateActions(actions []Action, seed int64) ValidationResult
}

type ValidationResult struct {
    Score       int
    Accuracy    float64
    StarsEarned int
    XPEarned    int
    Metrics     []BehavioralMetric
    Rejected    bool
    RejectReason string
}

// Always call validator BEFORE writing to DB
// If Rejected == true: log with profile_id + action count; return 422
```

---

*v2.2 — Full schema, WS protocol, security detail.*
