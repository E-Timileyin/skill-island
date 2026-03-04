# Skill Island — Copilot Instructions

> Primary instructions file. Copilot reads this on every prompt.
> Full game logic: see game-logic-mvp.md and game-logic-deferred.md
> Full data model: see data-model.md

---

## What This Project Is

Skill Island is a **web-first, gamified educational platform** for autistic SEND
teenagers (ages 13–18). It combines single-player cognitive mini-games with
2-player cooperative multiplayer, and provides a read-only analytics dashboard
for parents and educators.

**Core constraint:** Every decision must prioritise safety of the student
experience. This audience includes vulnerable minors. No shortcuts on security.

---

## Monorepo Structure

```
skill-island/
├── apps/
│   └── web/                        # Next.js frontend (TypeScript)
│       ├── app/                    # App Router pages
│       │   ├── (auth)/             # Login, register pages
│       │   ├── (student)/          # Island map, game host, profile
│       │   └── (dashboard)/        # Parent/educator analytics
│       ├── components/             # Shared React components
│       │   └── game/               # UI overlays from Gemini-generated screens
│       ├── game/                   # Phaser 3 game source (TypeScript)
│       │   ├── scenes/             # One file per game zone
│       │   ├── events/             # EventBus — game ↔ Next.js bridge
│       │   └── PhaserGame.tsx      # Canvas mount component
│       └── lib/                    # API client, auth helpers, types
├── services/
│   └── api/                        # Single Go backend service
│       ├── cmd/server/             # main.go entrypoint
│       ├── internal/
│       │   ├── api/                # REST route handlers
│       │   ├── ws/                 # WebSocket hub + room manager
│       │   ├── auth/               # JWT creation, validation, middleware
│       │   ├── db/                 # PostgreSQL pool + query functions
│       │   ├── validator/          # Server-side action + score validation
│       │   └── config/             # Env-based config loading
│       └── migrations/             # SQL migration files (golang-migrate)
├── docker-compose.yml
├── .github/
│   ├── copilot-instructions.md     # THIS FILE
│   ├── game-logic-mvp.md           # Memory Cove, Focus Forest, Team Tower
│   ├── game-logic-deferred.md      # Pattern Plateau, Community Hub
│   └── data-model.md               # Schema, WebSocket, security detail
```

---

## Tech Stack

| Layer | Technology | Notes |
|---|---|---|
| Frontend | Next.js 14+ (TypeScript) | App Router only; no Pages Router |
| Game Engine | Phaser 3 (TypeScript) | Dynamic import; never SSR |
| Backend | Go 1.22+ | Single binary; REST + WebSocket |
| Database | PostgreSQL 16 | All persistence |
| Auth | JWT in HTTP-only cookies | Never localStorage |
| Containers | Docker + Compose | All services |
| Proxy | Nginx | /api/*, /ws/*, / routing |
| Migrations | golang-migrate | SQL files only |
| WS Library | gorilla/websocket | Go WS |
| HTTP Router | chi | Go router |

---

## Zone Build Status — All 5 Zones Specified

| Zone | Type | Build Status | XP Required |
|---|---|---|---|
| Memory Cove | Solo | BUILD NOW (MVP) | 0 XP |
| Focus Forest | Solo | BUILD NOW (MVP) | 30 XP |
| Team Tower | 2-Player Co-op | BUILD NOW (MVP) | 80 XP |
| Pattern Plateau | Solo | DEFERRED — do not build yet | 150 XP |
| Community Hub | Social | DEFERRED — do not build yet | 250 XP |

> Pattern Plateau and Community Hub are fully specified in game-logic-deferred.md
> so Copilot understands the full system. Do not scaffold or implement them
> until MVP beta with schools is complete.

### Island Map Rule
Always render all 5 zone cards on the island map. Deferred zones show a padlock
and "Coming Soon" label — they are never clickable in the MVP build. Never hide
deferred cards; locked cards build anticipation.

---

## MVP Features — Build These

- JWT auth (student / parent / educator roles)
- Student profile setup (avatar, nickname, play mode)
- Island map — all 5 cards visible; 3 MVP zones unlockable by XP
- Star and XP system (server-calculated only — never client)
- Basic parent/educator dashboard (weekly summary cards)
- Behavioural metrics capture on every session from day one

## Explicitly Deferred — Do Not Build

- Pattern Plateau gameplay
- Community Hub gameplay
- Daily Group Missions
- Focus Forest co-op mode
- Adaptive AI difficulty
- LMS/SCORM integration

---

## Go Backend — Conventions

### Package responsibilities
```
internal/api/       → REST handlers only; no business logic in handlers
internal/ws/        → WebSocket hub, room manager, broadcast loop
internal/auth/      → JWT sign, verify, middleware; nothing else
internal/db/        → DB pool + all query functions; no handlers here
internal/validator/ → Score and action validation; called by api and ws
internal/config/    → Reads env vars; returns typed Config struct
```

### Handler pattern
```go
func (h *Handler) MethodName(w http.ResponseWriter, r *http.Request)

// Always return structured JSON errors — never raw error strings:
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

### Database rules
```go
// Always parameterised — never fmt.Sprintf into SQL
db.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email) // CORRECT

// Always context.Context as first arg to every db call
// Use transactions for writes touching more than one table
```

### Error handling
```go
// Wrap with context: fmt.Errorf("createSession: %w", err)
// Never swallow errors silently
// Log internally; return sanitised message to client
```

---

## Security Rules — Non-Negotiable (Platform Serves Minors)

### Auth
- JWTs in **HTTP-only, Secure, SameSite=Strict cookies only**
- Access token: 1 hour expiry
- Refresh token: 7 days; rotate on every use
- WS upgrade: validate JWT from cookie BEFORE accepting upgrade
- Reject unauthenticated WS with HTTP 401

### Score integrity
- **Client NEVER submits a score value**
- Client submits action stream with timestamps
- `internal/validator` replays actions → calculates authoritative score
- Stars and XP always server-calculated
- Reject sessions with implausible action counts

### Rate limits
| Endpoint | Limit |
|---|---|
| POST /api/auth/login | 5 / 15 min / IP |
| POST /api/auth/register | 3 / hour / IP |
| POST /api/sessions | 30 / min / profile |
| GET /api/analytics/* | 60 / min / user |
| WS /ws/game connect | 3 / min / user |
| WS messages per room | 20 / sec / player |

### General
- Parameterised queries everywhere
- Never return `password_hash` or `refresh_token` in responses
- CORS: strict allow-list only
- CSP headers at Nginx level
- No free-form text input from students anywhere

---

## Next.js Frontend — Conventions

### Route structure
```
app/(auth)/login/page.tsx            Public
app/(auth)/register/page.tsx         Public
app/(student)/island/page.tsx        Protected; student only
app/(student)/game/[zone]/page.tsx   Protected; student only; mounts Phaser
app/(dashboard)/overview/page.tsx    Protected; parent/educator only
```

### Phaser integration
```tsx
// Always dynamic import — Phaser cannot run server-side
const PhaserGame = dynamic(() => import('@/game/PhaserGame'), { ssr: false })

// Phaser ↔ Next.js via EventBus only (see game-logic-mvp.md)
// Next.js listens for 'game:session-end' → POSTs to /api/sessions
```

### Auth
- `useAuth()` hook reads JWT via `GET /api/auth/me`
- Never store identity in localStorage
- Redirect unauthenticated users in Next.js middleware

### API calls
```typescript
// All calls through typed client in lib/api.ts
// Never call fetch() directly in a component
// Always handle loading + error states
```

---

## Environment Variables

```bash
# services/api/.env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/skillisland
JWT_SECRET=change-this-in-production-minimum-32-chars
JWT_REFRESH_SECRET=change-this-in-production-different-from-above
PORT=8080
ENV=development
ALLOWED_ORIGINS=http://localhost:3000

# apps/web/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

---

## Docker Compose (local dev)

- `postgres` — port 5432
- `api` — Go on port 8080; hot-reload via `air`
- `web` — Next.js on port 3000

Start all: `docker compose up`

---

## Good Copilot Prompts

```
"Generate the SQL migration for behavioral_metrics — schema in data-model.md"
"Write POST /api/sessions handler that accepts action log, calls internal/validator"
"Create the WS room manager enforcing 2-player max and WAITING→READY→PLAYING lifecycle"
"Write MemoryCoveScene.ts Phaser scene following the spec in game-logic-mvp.md"
"Write the nightly analytics aggregation SQL for 7-day rolling attention score"
"Scaffold the Go internal/validator/memory.go following game-logic-mvp.md"
```

## Do Not Ask Copilot To

- Calculate scores client-side
- Store JWTs in localStorage
- Skip error handling for brevity
- Add rooms with more than 2 players
- Build Pattern Plateau or Community Hub (deferred)

---

## Phase 0 Checklist — Prove Stack Before Any Game Code

- [ ] Go service starts; `/health` returns 200
- [ ] PostgreSQL connects; migrations run cleanly
- [ ] `POST /api/auth/register` creates user; JWT cookie set
- [ ] `GET /api/auth/me` returns role from cookie
- [ ] 2 clients connect to `/ws/game`; server sends `room_ready`; both get tick
- [ ] Phaser canvas renders in Next.js without SSR errors
- [ ] Phaser emits test event; Next.js receives and logs it
- [ ] `docker compose up` starts all 3 services from cold

**Do not start Phase 1 until every box is checked.**

---

*v2.2 — Split from monolithic AGENTS.md. See sibling files for game logic and schema.*
