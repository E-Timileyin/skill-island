# Skill Island — API Documentation

> **Base URL (local dev):** `http://localhost:8080`  
> **Content-Type:** All request/response bodies are `application/json`  
> **Auth mechanism:** HTTP-only cookies (`access_token`, `refresh_token`)

---

## Table of Contents

1. [Overview](#overview)
2. [Authentication & Cookies](#authentication--cookies)
3. [Error Format](#error-format)
4. [Route Map](#route-map)
5. [Auth Routes](#auth-routes)
   - [POST /api/auth/register](#post-apiauthregister)
   - [POST /api/auth/login](#post-apiauthlogin)
   - [POST /api/auth/logout](#post-apiauthlogout)
   - [POST /api/auth/refresh](#post-apiauthrefresh)
   - [GET /api/auth/me](#get-apiauthme)
6. [Profile Routes](#profile-routes)
   - [POST /api/profiles](#post-apiprofiles)
   - [GET /api/profiles/me](#get-apiprofilesme)
   - [PATCH /api/profiles/me](#patch-apiprofilesme)
7. [Session Routes](#session-routes)
   - [POST /api/sessions](#post-apisessions)
8. [Analytics Routes](#analytics-routes)
   - [GET /api/analytics/overview](#get-apianalyticsoverview)
9. [WebSocket](#websocket)
   - [GET /ws/game](#get-wsgame)
10. [Health Check](#health-check)
11. [Game Types & Actions](#game-types--actions)
    - [Memory Cove](#memory-cove-memory_cove)
    - [Focus Forest](#focus-forest-focus_forest)
    - [Team Tower](#team-tower-team_tower)
12. [XP & Zone Unlocks](#xp--zone-unlocks)
13. [Roles & Permissions](#roles--permissions)
14. [Project Structure](#project-structure)

---

## Overview

Skill Island is a gamified learning platform. The API handles:

- **Auth** — registration, login, token refresh, current user
- **Profiles** — student profiles (nickname, avatar, play mode, XP, stars)
- **Sessions** — server-side game session validation and scoring
- **Analytics** — weekly performance snapshots for parents/educators
- **WebSocket** — real-time cooperative game support via a persistent hub

All scoring happens **server-side**. Clients submit raw action logs; the server validates and scores them.

---

## Authentication & Cookies

After a successful **register**, **login**, or **refresh**, the server sets two HTTP-only cookies:

| Cookie | Purpose | Lifetime |
|---|---|---|
| `access_token` | Short-lived JWT for API auth | ~15 min |
| `refresh_token` | Long-lived JWT for obtaining a new access token | ~7 days |

Both cookies are `HttpOnly` and `SameSite=Strict`. In production (`ENV != development`) they are also `Secure`.

Include cookies automatically if your client is a browser. For non-browser clients use a cookie jar or set the `Cookie` header manually.

---

## Error Format

All error responses share the same structure:

```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable description"
}
```

### Common Error Codes

| Code | HTTP Status | Meaning |
|---|---|---|
| `BAD_REQUEST` | 400 | Malformed JSON body |
| `VALIDATION_ERROR` | 400 | Field-level validation failure |
| `UNAUTHORIZED` | 401 | Missing or invalid access token / cookie |
| `INVALID_CREDENTIALS` | 401 | Wrong email or password |
| `FORBIDDEN` | 403 | Authenticated but wrong role |
| `NOT_FOUND` | 404 | Resource does not exist |
| `DUPLICATE_EMAIL` | 409 | Email already registered |
| `DUPLICATE_PROFILE` | 409 | Student profile already exists for user |
| `SESSION_REJECTED` | 422 | Server-side validation rejected the session |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Route Map

```
GET    /health
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/logout
POST   /api/auth/refresh
GET    /api/auth/me             🔒 access_token required
POST   /api/sessions            🔒 student role only
GET    /api/analytics/overview  🔒 parent / educator role only
GET    /ws/game                 🔒 access_token cookie required
POST   /api/profiles            🔒 student role only
GET    /api/profiles/me         🔒 student role only
PATCH  /api/profiles/me         🔒 student role only
```

---

## Auth Routes

### `POST /api/auth/register`

Creates a new user account and immediately issues auth cookies.

**Request Body**

```json
{
  "email": "student@example.com",
  "password": "supersecret123",
  "role": "student"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `email` | string | yes | Normalised to lowercase + trimmed |
| `password` | string | yes | Minimum 8 characters |
| `role` | string | yes | `student`, `parent`, or `educator` |

**Success Response — `201 Created`**

```json
{
  "id": "uuid-of-user",
  "email": "student@example.com",
  "role": "student"
}
```

**Sets Cookies:** `access_token`, `refresh_token`

**Error Responses**

| Status | Code | When |
|---|---|---|
| 400 | `VALIDATION_ERROR` | Missing field, password < 8 chars, invalid role |
| 409 | `DUPLICATE_EMAIL` | Email already exists |
| 500 | `INTERNAL_ERROR` | DB or token generation failure |

---

### `POST /api/auth/login`

Authenticates an existing user and issues auth cookies.

**Request Body**

```json
{
  "email": "student@example.com",
  "password": "supersecret123"
}
```

**Success Response — `200 OK`**

```json
{
  "id": "uuid-of-user",
  "email": "student@example.com",
  "role": "student"
}
```

**Sets Cookies:** `access_token`, `refresh_token`

**Error Responses**

| Status | Code | When |
|---|---|---|
| 400 | `VALIDATION_ERROR` | Empty email or password |
| 401 | `INVALID_CREDENTIALS` | Wrong email or password |
| 500 | `INTERNAL_ERROR` | DB or token failure |

---

### `POST /api/auth/logout`

Clears auth cookies. No request body required.

**Success Response — `200 OK`**

```json
{ "message": "logged out" }
```

---

### `POST /api/auth/refresh`

Validates the `refresh_token` cookie and issues a new `access_token` + `refresh_token`.

**Cookies Required:** `refresh_token`

**Success Response — `200 OK`**

```json
{ "message": "tokens refreshed" }
```

**Sets Cookies:** `access_token`, `refresh_token` (rotated)

**Error Responses**

| Status | Code | When |
|---|---|---|
| 401 | `UNAUTHORIZED` | Missing, invalid, or expired `refresh_token` |
| 500 | `INTERNAL_ERROR` | DB or token failure |

---

### `GET /api/auth/me`

> **Auth required:** `access_token` cookie

Returns the currently authenticated user. For students, includes embedded profile data if a profile exists.

**Success Response — `200 OK` (non-student)**

```json
{
  "id": "uuid-of-user",
  "email": "parent@example.com",
  "role": "parent"
}
```

**Success Response — `200 OK` (student with profile)**

```json
{
  "id": "uuid-of-user",
  "email": "student@example.com",
  "role": "student",
  "profile": {
    "id": "uuid-of-profile",
    "nickname": "StarBlaster",
    "avatar_id": 3,
    "total_stars": 42,
    "total_xp": 210,
    "play_mode": "solo"
  }
}
```

**Error Responses**

| Status | Code | When |
|---|---|---|
| 401 | `UNAUTHORIZED` | Missing or invalid `access_token` |
| 404 | `NOT_FOUND` | User record deleted from DB |

---

## Profile Routes

All profile routes require a valid `access_token` cookie and the `student` role.

### `POST /api/profiles`

Creates the student profile for the authenticated user. Only one profile per user allowed.

**Request Body**

```json
{
  "nickname": "StarBlaster",
  "avatar_id": 3,
  "play_mode": "solo"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `nickname` | string | yes | Display name |
| `avatar_id` | int | no | Defaults to `0` |
| `play_mode` | string | no | `solo` (default) or `team` |

**Success Response — `201 Created`**

```json
{
  "id": "uuid-of-profile",
  "nickname": "StarBlaster",
  "avatar_id": 3,
  "total_stars": 0,
  "total_xp": 0,
  "play_mode": "solo",
  "created_at": "2026-03-07T15:00:00Z"
}
```

**Error Responses**

| Status | Code | When |
|---|---|---|
| 400 | `VALIDATION_ERROR` | Missing nickname, invalid play_mode |
| 403 | `FORBIDDEN` | Not a student |
| 409 | `DUPLICATE_PROFILE` | Profile already exists |

---

### `GET /api/profiles/me`

Returns the student profile for the authenticated user.

**Success Response — `200 OK`**

```json
{
  "id": "uuid-of-profile",
  "nickname": "StarBlaster",
  "avatar_id": 3,
  "total_stars": 42,
  "total_xp": 210,
  "play_mode": "solo",
  "created_at": "2026-03-07T15:00:00Z"
}
```

**Error Responses**

| Status | Code | When |
|---|---|---|
| 403 | `FORBIDDEN` | Not a student |
| 404 | `NOT_FOUND` | No profile created yet |

---

### `PATCH /api/profiles/me`

Updates one or more profile fields. At least one field must be provided.

**Request Body** (all optional, but at least one required)

```json
{
  "nickname": "NewName",
  "avatar_id": 5,
  "play_mode": "team"
}
```

| Field | Type | Notes |
|---|---|---|
| `nickname` | string | Cannot be set to empty string |
| `avatar_id` | int | Any non-negative integer |
| `play_mode` | string | `solo` or `team` |

**Success Response — `200 OK`** — same shape as `GET /api/profiles/me`

**Error Responses**

| Status | Code | When |
|---|---|---|
| 400 | `VALIDATION_ERROR` | No fields provided, empty nickname, or invalid play_mode |
| 403 | `FORBIDDEN` | Not a student |
| 404 | `NOT_FOUND` | No profile exists |

---

## Session Routes

### `POST /api/sessions`

> **Auth required:** `access_token` cookie, `student` role, and a student profile

Submits a completed game session. **Never send scores from the client.** All scoring is computed server-side from raw actions.

**Request Body**

```json
{
  "game_type": "memory_cove",
  "mode": "solo",
  "duration_ms": 45000,
  "room_session_id": "",
  "actions": [...]
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `game_type` | string | yes | `memory_cove`, `focus_forest`, or `team_tower` |
| `mode` | string | yes | `solo` or `cooperative` |
| `duration_ms` | int | yes | Must be > 0 |
| `actions` | array | yes | 1 – 500 action objects |
| `room_session_id` | string | no | Only for cooperative sessions |

**Success Response — `201 Created`**

```json
{
  "score": 800,
  "accuracy": 0.80,
  "stars_earned": 2,
  "xp_earned": 20,
  "total_xp": 230,
  "unlocked_zones": ["memory_cove", "focus_forest"],
  "behavioral_metrics_count": 10
}
```

| Field | Type | Notes |
|---|---|---|
| `score` | int | Raw score (game-type specific) |
| `accuracy` | float | 0.0 – 1.0 |
| `stars_earned` | int | 0, 1, 2, or 3 |
| `xp_earned` | int | XP awarded this session |
| `total_xp` | int | Cumulative XP after this session |
| `unlocked_zones` | string[] | All currently unlocked zones |
| `behavioral_metrics_count` | int | Behavioral data points saved |

**Error Responses**

| Status | Code | When |
|---|---|---|
| 400 | `VALIDATION_ERROR` | Invalid game_type, mode, duration, or empty actions |
| 401 | `UNAUTHORIZED` | Missing access token |
| 403 | `FORBIDDEN` | Not a student, or no profile |
| 422 | `SESSION_REJECTED` | Session is implausible (action count out of range, etc.) |

---

## Analytics Routes

### `GET /api/analytics/overview`

> **Auth required:** `access_token` cookie, `parent` or `educator` role

Returns the latest weekly performance snapshot for a student.

**Query Parameters**

| Param | Required | Description |
|---|---|---|
| `profile_id` | yes | UUID of the student profile to query |

**Example:** `GET /api/analytics/overview?profile_id=uuid-here`

**Cache-Control:** `private, max-age=900` (15-minute client cache)

**Success Response — `200 OK` (data available)**

```json
{
  "attention_score": 0.78,
  "memory_score": 0.65,
  "engagement_frequency": 5,
  "coop_participation_rate": 0.40,
  "avg_reaction_time_ms": 412.5,
  "total_stars": 42,
  "total_xp": 210,
  "sessions_this_week": 5,
  "snapshot_date": "2026-03-07"
}
```

**Success Response — `200 OK` (no data yet)**

```json
{
  "attention_score": null,
  "memory_score": null,
  "engagement_frequency": 0,
  "coop_participation_rate": null,
  "avg_reaction_time_ms": null,
  "total_stars": 0,
  "total_xp": 0,
  "sessions_this_week": 0,
  "snapshot_date": "",
  "message": "No data yet"
}
```

| Field | Type | Notes |
|---|---|---|
| `attention_score` | float or null | Average from Focus Forest sessions |
| `memory_score` | float or null | Average from Memory Cove sessions |
| `engagement_frequency` | int | Sessions in the snapshot period |
| `coop_participation_rate` | float or null | Fraction of cooperative sessions |
| `avg_reaction_time_ms` | float or null | Average tap reaction time (ms) |
| `total_stars` | int | Lifetime total |
| `total_xp` | int | Lifetime total |
| `sessions_this_week` | int | Alias of `engagement_frequency` |
| `snapshot_date` | string | ISO date `YYYY-MM-DD` |
| `message` | string | Present only when no data exists |

**Error Responses**

| Status | Code | When |
|---|---|---|
| 400 | `VALIDATION_ERROR` | Missing `profile_id` query param |
| 401 | `UNAUTHORIZED` | Missing access token |
| 403 | `FORBIDDEN` | Not a parent or educator |

---

## WebSocket

### `GET /ws/game`

> **Auth required:** `access_token` cookie (validated before HTTP upgrade)

Upgrades the HTTP connection to a persistent WebSocket for real-time cooperative gameplay.

**Connection Flow**

1. Client connects with valid `access_token` cookie
2. Server validates JWT **before** the WebSocket upgrade
3. On invalid or missing cookie: HTTP `401` is returned (not a WS error frame)
4. On success: connection is registered in the in-memory WebSocket hub

**On Auth Failure — HTTP `401`**

```json
{
  "code": "UNAUTHORIZED",
  "message": "missing or invalid access token"
}
```

> The WebSocket message protocol (room join, game sync events, etc.) is handled by `internal/ws` and documented separately.

---

## Health Check

### `GET /health`

Public. No authentication required.

**Success Response — `200 OK`**

```json
{ "status": "ok" }
```

---

## Game Types & Actions

### Memory Cove (`memory_cove`)

A sequence memory game. The server generates a deterministic shape-colour sequence from a seed. The client must press buttons in correct sequence order.

**Action shape**

```json
{
  "type": "press",
  "button_id": "circle-red",
  "element_index": 0,
  "client_timestamp": 1200
}
```

| Field | Type | Description |
|---|---|---|
| `type` | string | Always `"press"` |
| `button_id` | string | `"{shape}-{colour}"` e.g. `"circle-red"`, `"square-blue"` |
| `element_index` | int | 0-based index of the sequence element |
| `client_timestamp` | int64 | Client ms since session start |

**Valid shapes:** `circle`, `square`, `triangle`, `star`  
**Valid colours:** `red`, `blue`, `green`, `yellow`

**Scoring**

| Accuracy | Stars |
|---|---|
| >= 90% | 3 stars |
| >= 70% | 2 stars |
| >= 50% | 1 star |
| < 50% | 0 stars |

---

### Focus Forest (`focus_forest`)

An attention game. Players tap butterflies and avoid bees. Targets are spawned from a server-generated deterministic manifest.

**Action shape**

```json
{
  "type": "tap",
  "tap_x": 0.45,
  "tap_y": 0.62,
  "client_timestamp": 3200
}
```

| Field | Type | Description |
|---|---|---|
| `type` | string | Always `"tap"` |
| `tap_x` | float | Normalised X position (0.0 – 1.0) |
| `tap_y` | float | Normalised Y position (0.0 – 1.0) |
| `client_timestamp` | int64 | Client ms since session start |

**Scoring — Attention Score**

Formula: `attention = (butterfly_hits - bee_hits × 0.5) / total_butterflies`

| Attention Score | Stars |
|---|---|
| >= 0.85 | 3 stars |
| >= 0.60 | 2 stars |
| >= 0.30 | 1 star |
| < 0.30 | 0 stars |

---

### Team Tower (`team_tower`)

A cooperative game mode. Requires `"mode": "cooperative"` and `room_session_id`. Server-side scoring uses a generic fallback until the full validator is implemented.

> Full action schema and scoring for Team Tower will be documented when the validator is complete.

---

## XP & Zone Unlocks

XP accumulates on the student profile. As total XP crosses thresholds, new zones unlock.

### XP Earned per Session

| Game Type | Stars | XP |
|---|---|---|
| `memory_cove` | any | `stars × 10 + rounds_completed × 5` |
| `focus_forest` or `team_tower` | 0 | 0 |
| `focus_forest` or `team_tower` | 1 | 10 |
| `focus_forest` or `team_tower` | 2 | 20 |
| `focus_forest` or `team_tower` | 3 | 35 |

### Zone Unlock Thresholds

| Zone | XP Required |
|---|---|
| `memory_cove` | 0 (always unlocked) |
| `focus_forest` | 30 XP |
| `team_tower` | 80 XP |
| `pattern_plateau` | 150 XP |
| `community_hub` | 250 XP |

The session response always includes the full `unlocked_zones` list reflecting the student's current total XP.

---

## Roles & Permissions

| Route | student | parent | educator |
|---|---|---|---|
| `POST /api/auth/register` | yes | yes | yes |
| `POST /api/auth/login` | yes | yes | yes |
| `POST /api/auth/logout` | yes | yes | yes |
| `POST /api/auth/refresh` | yes | yes | yes |
| `GET /api/auth/me` | yes | yes | yes |
| `POST /api/profiles` | yes | no | no |
| `GET /api/profiles/me` | yes | no | no |
| `PATCH /api/profiles/me` | yes | no | no |
| `POST /api/sessions` | yes | no | no |
| `GET /api/analytics/overview` | no | yes | yes |
| `GET /ws/game` | yes | no | no |

---

## Project Structure

```
services/api/
├── cmd/
│   └── server/
│       └── main.go                  # Entry: DB connect, migrations, start server
├── internal/
│   ├── api/
│   │   ├── handlers/                # HTTP handler implementations
│   │   │   ├── base.go              # Handler struct, Health endpoint, writeJSON()
│   │   │   ├── auth.go              # Register, Login, Logout, Refresh, Me
│   │   │   ├── profile_handler.go   # CreateProfile, GetProfile, UpdateProfile
│   │   │   ├── session_handler.go   # SubmitSession + server-side action validation
│   │   │   ├── analytics_handler.go # AnalyticsOverview
│   │   │   ├── ws_handler.go        # WebSocket upgrade handler (ServeWS)
│   │   │   └── types.go             # Shared request/response structs
│   │   └── routes/
│   │       └── routes.go            # SetupRouter — all routes in one place
│   ├── auth/                        # JWT generation, validation, cookie helpers
│   ├── config/                      # Env-based config loader
│   ├── db/                          # DB queries and transaction helpers (pgx/v5)
│   ├── validator/                   # Server-side game scoring logic
│   │   ├── memory.go                # Memory Cove sequence generation & validation
│   │   ├── focus.go                 # Focus Forest spawn manifest & tap validation
│   │   ├── xp.go                   # XP table, zone unlock thresholds
│   │   └── validator.go             # Shared ValidationResult / BehavioralMetric types
│   └── ws/                         # WebSocket hub for real-time co-op
├── migrations/                      # SQL migration files (applied on startup)
└── API_DOCS.md                      # This documentation file
```
