## Next.js 14 · TypeScript · Phaser 3 · Tailwind

> Pure frontend work only. Assign each issue to @copilot.
> TODO 18-FE, 23-FE, 24-FE, 26-FE, 27-FE, 28-FE, 30-FE, 31-FE each depend on
> their matching backend first before frontend work starts.
> TODO 21 and 22 depend on TODO 20 (backend WS room) being merged.

---

# PHASE 2 CONTINUED — Focus Forest

---

## Issue 18-FE — Focus Forest Phaser Scene, UI & Full Integration

```
Title: Phase 2.6-FE — Focus Forest Phaser scene, UI overlay, and full page integration

Read .github/copilot-instructions.md and .github/game-logic-mvp.md before starting.
Depends on: Issue 18-BE merged (manifest endpoint + session API available).

─────────────────────────────────────────
PART A — Phaser Scene
─────────────────────────────────────────

1. apps/web/game/scenes/FocusForestScene.ts

   On scene create:
   a. Call POST /api/sessions/init with game_type: "focus_forest"yes
   b. Store session_token, seed, difficulty_level
   c. Call GET /api/sessions/manifest?token={session_token}
   d. Store manifest locally for spawn display
   e. Initialise action log: FocusForestAction[]
   f. Initialise session timer: 60000ms countdown

   Spawn loop:
   — Use Phaser time.addEvent with repeating delay based on difficulty spawnIntervalMs
   — On each spawn event: find next unspawned target from manifest
   — Create Phaser Image (placeholder coloured circle for now)
   — Apply Tween for horizontal movement:
     duration = (1.0 / speedPxPerS) * scene.scale.width * 1000
     x: from spawn.position_x * scene.scale.width to scene.scale.width + 100
     ease: 'Linear'
   — Schedule despawn: time.delayedCall(DESPAWN_AFTER_MS, () => destroyTarget(id))
   — Track all active targets in Map<targetId, GameObject>

   On pointer down (tap/click):
   — Calculate normalised position:
     tap_x = pointer.x / scene.scale.width
     tap_y = pointer.y / scene.scale.height
   — Record action: { type: "tap", tap_x, tap_y, client_timestamp: scene.time.now }
   — Append to action log
   — Emit EventBus 'game:action' (debug only)
   — Visual feedback at tap position:
     Butterfly hit: sparkle particle burst → image fades out over 300ms
     Bee hit: shake tween 200ms → image fades out
     Miss: small ripple effect at tap point
   — SEND rule: NO "Wrong!" text, NO negative sound for bee/miss
   — Remove target from active map on hit

   Session timer:
   — Emit 'game:ui-update' every second: { timeRemainingMs, butterfliesHit }
   — At 0ms:
     — Stop spawn loop
     — Cancel all pending despawn timers
     — Destroy all remaining target objects
     — Emit EventBus 'game:session-end' with full payload

   Constants:
   const DESPAWN_AFTER_MS    = 3000
   const SESSION_DURATION_MS = 60000

   SEND rules:
   — No score display during gameplay — result screen only
   — Countdown bar fades colour gently (green → amber → soft coral)
   — No "Time's running out!" text — silent visual only
   — Bee tap = visual feedback only, never punishing language

2. apps/web/game/logic/spawnPatterns.ts

   getSpeedForDifficulty(level: number): number
   — Returns px/s (80 | 100 | 130 | 160)

   getSpawnIntervalForDifficulty(level: number): number
   — Returns interval ms (1200 | 1000 | 800 | 700)

   getTargetColour(targetType: string): number
   — butterfly_blue:   0x4A90E2
   — butterfly_orange: 0xF5A623
   — butterfly_red:    0xD0021B
   — bee:              0xF8E71C

─────────────────────────────────────────
PART B — UI Overlay
─────────────────────────────────────────

3. components/game/FocusForestUI.tsx

   Use Gemini-generated component if it exists in components/game/.
   Otherwise:

   interface FocusForestUIProps {
     timeRemainingMs: number
     totalDurationMs: number
     butterfliesHit: number
     phase: 'playing' | 'complete'
   }

   Renders (pointer-events-none overlay over Phaser canvas):
   — Countdown bar: top of screen, full width
     progress = timeRemainingMs / totalDurationMs
     Colour: green (>50%) → amber (20–50%) → soft coral (<20%)
     No number countdown — bar only (SEND rule: less pressure)
   — Butterfly counter: top right — "🦋 × {butterfliesHit}"
   — Zone title: top left — "Focus Forest"
   — Pause button placeholder: top left corner icon

─────────────────────────────────────────
PART C — Page & Integration
─────────────────────────────────────────

4. app/(student)/game/focus-forest/page.tsx

   a. Verify student profile exists — redirect to /setup if not
   b. Mount <PhaserGame scene="FocusForestScene" />
   c. Mount <FocusForestUI /> receiving state via EventBus 'game:ui-update'
   d. Listen for EventBus 'game:session-end'
   e. On session-end: call lib/api.ts submitSession()
   f. Show loading overlay: "Saving your progress..."
   g. On success: show <SessionResultScreen /> (reuse from Issue 16)
   h. On API error: show "Something went wrong — your progress is saved"
      with retry button + "Back to Island" fallback

5. lib/api.ts additions

   getSessionManifest(token: string): Promise<SpawnEvent[]>
   — GET /api/sessions/manifest?token={token}

6. EventBus updates
   — 'game:ui-update' payload for Focus Forest:
     { timeRemainingMs: number, butterfliesHit: number }

─────────────────────────────────────────
PART D — Island Map Integration
─────────────────────────────────────────

7. app/(student)/island/page.tsx (update)
   — After returning from any session: refresh profile
   — If Focus Forest newly unlocked: pulse animation on Focus Forest card
   — Focus Forest zone card navigates to /game/focus-forest on click

Phase 2-FE is complete when:
— Focus Forest is fully playable end to end
— Tap actions recorded, session submitted, result screen shows
— Island map card links to game and pulses on first unlock
```

---

---

# PHASE 3 — Multiplayer (Team Tower)

---

## Issue 21 — Team Tower Phaser Scene

```
Title: Phase 3.3 — Team Tower Phaser scene (client rendering)

Read .github/copilot-instructions.md and .github/game-logic-mvp.md before starting.
Depends on: Issue 20 (WebSocket room) merged.

─────────────────────────────────────────
PART A — Scene Setup
─────────────────────────────────────────

1. apps/web/game/scenes/TeamTowerScene.ts

   Properties:
   — playerRole: "player_1" | "player_2"
   — myTurn: boolean
   — currentState: TowerState | null
   — blockObjects: Map<string, Phaser.GameObjects.Rectangle>
   — ws: WebSocket | null
   — heartbeatInterval: number
   — actionLog: TeamTowerAction[]
   — sessionStartTime: number

   On scene create:
   a. Connect WebSocket: new WebSocket(process.env.NEXT_PUBLIC_WS_URL + "/ws/game")
   b. ws.onmessage = handleServerMessage
   c. ws.onclose   = handleDisconnect
   d. ws.onerror   = handleError
   e. On ws.onopen: send { type: "join_room", game_type: "team_tower" }
   f. Show "Finding a partner..." with pulsing animation
   g. Start heartbeat: setInterval(() => sendPing(), HEARTBEAT_INTERVAL_MS)
   h. sessionStartTime = Date.now()

   Constants:
   const LERP_FACTOR           = 0.25
   const HEARTBEAT_INTERVAL_MS = 5000
   const BLOCK_WIDTH_PX        = 80
   const BLOCK_HEIGHT_PX       = 40

─────────────────────────────────────────
PART B — Message Handling
─────────────────────────────────────────

2. handleServerMessage(event: MessageEvent)

   Parse JSON; switch on message.type:

   case "waiting_for_partner":
   — Emit 'game:ui-update' { phase: "waiting" }

   case "room_ready":
   — Store playerRole = message.player_role
   — Show "Get ready!" briefly (1500ms)
   — Emit 'game:ui-update' { phase: "ready", playerRole, opponentAvatar }
   — After 1500ms: transition to playing phase

   case "state_update":
   — Call reconcileState(message.game_state)
   — Emit 'game:ui-update' {
       phase: "playing",
       groupXP: state.group_xp,
       activePlayer: state.active_player,
       myRole: playerRole,
       turnNumber: state.turn_number,
       nextBlockShape: state.next_block_shape
     }

   case "action_rejected":
   — If reason "not_your_turn": shake turn indicator
   — If reason "out_of_bounds": shake screen edges briefly
   — Do NOT show error text to student

   case "player_disconnected":
   — Emit 'game:ui-update' { phase: "partner_disconnected" }
   — Disable all input
   — Show warm message: "Your partner will be right back..."

   case "session_end":
   — clearInterval(heartbeatInterval)
   — ws.close()
   — Emit EventBus 'game:session-end' {
       game_type: "team_tower",
       outcome: message.outcome,
       stars: message.stars,
       group_xp: message.group_xp,
       room_session_id: message.room_session_id,
       duration_ms: Date.now() - sessionStartTime
     }

   case "idle_warning":
   — Emit 'game:ui-update' { phase: "idle_warning", secondsRemaining: message.seconds_remaining }

   case "heartbeat_pong":
   — Record RTT for monitoring (optional)

─────────────────────────────────────────
PART C — State Reconciliation
─────────────────────────────────────────

3. reconcileState(newState: TowerState)

   For each block in newState.blocks:
   — If block.id not in blockObjects:
     Create Phaser.GameObjects.Rectangle
     Position: (block.x * scene.scale.width, scene.scale.height - block.y * BLOCK_HEIGHT_PX)
     Fill colour from getBlockColour(block.colour)
     Dimensions from getBlockDimensions(block.shape)
     Add to blockObjects map
     Spawn animation: scale 0.1 → 1.0 over 150ms
   — If block exists: lerp position each update cycle with LERP_FACTOR = 0.25

   If newState.stable === false AND currentState.stable === true:
   — Tower fall animation:
     1. Camera shake: this.cameras.main.shake(300, 0.01)
     2. All blocks: tween rotation +/- 45 degrees over 600ms
     3. All blocks: tween y += 200 over 800ms with gravity ease
     4. After 1200ms: emit 'game:session-end' { outcome: "lose" }
   — SEND rule: slow dramatic lean, NOT instant disappear

   If newState.current_height >= newState.target_height:
   — Win animation:
     1. All blocks: rainbow colour cycle tween
     2. Particle burst from top of tower
     3. After 1500ms: emit 'game:session-end' { outcome: "win" }

   currentState = newState

   CoM warning (visual only):
   — Calculate CoM: sum(block.x * weight) / totalWeight
   — If abs(CoM - 0.5) > 0.25: subtle camera sway (pre-warning)
   — If abs(CoM - 0.5) > 0.30: increase sway intensity
   — Never reaches 0.35 on client — server triggers fall first

─────────────────────────────────────────
PART D — Input
─────────────────────────────────────────

4. Input handling

   On pointer down:
   — If RoomState != "playing": ignore
   — If currentState.active_player != playerRole: ignore
   — normalised position_x = pointer.x / scene.scale.width; clamp [0.05, 0.95]
   — Record + append action: { type: "place_block", position_x, client_timestamp }
   — Send to server immediately via ws.send()
   — Show placement preview: dashed vertical line at position_x (fades after 500ms)
   — Block preview: ghost image of NextBlockShape at position_x

   Drop zone guide:
   — Valid drop area (x: 0.05–0.95): subtle background rectangle
   — MY turn: guide visible and bright
   — NOT my turn: guide hidden entirely

5. apps/web/game/logic/teamTower.ts

   lerpBlock(current, target, factor): { x, y }
   getBlockColour(colour: string): number   // Phaser hex values
   getBlockDimensions(shape: string): { w: number, h: number }  // pixels
   calculateCoM(blocks: Block[]): number    // normalised centre of mass x
```

---

## Issue 22 — Team Tower UI & Full Integration

```
Title: Phase 3.4 — Team Tower UI overlay, matchmaking screen, and full integration

Read .github/copilot-instructions.md and .github/game-logic-mvp.md before starting.
Depends on: Issue 21 (Phaser scene) merged.

─────────────────────────────────────────
PART A — Page
─────────────────────────────────────────

1. app/(student)/game/team-tower/page.tsx

   On load:
   a. Verify student profile exists — redirect to /setup if not
   b. Check profile.play_mode === "team"
      If play_mode === "solo":
      — Do NOT mount the game
      — Show friendly message:
        "Team Tower needs two players! Switch to Team mode in your profile to play."
      — Link to /setup with "Update My Profile" button
   c. Mount <PhaserGame scene="TeamTowerScene" />
   d. Mount <TeamTowerUI /> receiving state via EventBus 'game:ui-update'
   e. Mount <WaitingForPartner /> conditionally (phase === "waiting")
   f. Listen for EventBus 'game:session-end'
   g. On session-end:
      — Unmount game
      — Call lib/api.ts submitCoopSession()
      — Show loading: "Saving your teamwork..."
      — On success: show <CoopSessionResultScreen />
      — On error: show "Something went wrong — your progress is saved"

─────────────────────────────────────────
PART B — UI Overlay
─────────────────────────────────────────

2. components/game/TeamTowerUI.tsx

   Use Gemini-generated component if available, otherwise:

   interface TeamTowerUIProps {
     phase: 'waiting' | 'ready' | 'playing' | 'partner_disconnected' | 'idle_warning' | 'complete'
     groupXP: number
     groupXPTarget: number
     activePlayer: string
     myRole: string
     turnNumber: number
     opponentAvatarId: number
     partnerStatus: 'connected' | 'disconnected'
     idleSecondsRemaining?: number
   }

   When phase === "playing":
   — Turn banner (top centre, large):
     MY TURN:       bright green pill — "Your Turn! 🎯"
     PARTNER'S TURN: soft blue pill  — "Partner's Turn..."
   — Group XP bar (bottom, full width): groupXP / groupXPTarget
   — Turn counter (top right): "Turn {turnNumber}"
   — Opponent avatar (top right corner): placeholder circle
   — Next block preview (bottom right): NextBlockShape outline

   When phase === "partner_disconnected":
   — Semi-transparent warm amber overlay
   — "Your partner will be right back... ⏳"
   — "The game is paused for 30 seconds"
   — No cancel button during reconnect window

   When phase === "idle_warning":
   — Gentle pulsing border around game area
   — Bottom banner: "Still there? Place a block to continue! 😊"
   — No countdown number — SEND rule

─────────────────────────────────────────
PART C — Waiting Screen
─────────────────────────────────────────

3. components/game/WaitingForPartner.tsx

   interface WaitingForPartnerProps {
     onCancel: () => void
     waitingSeconds: number
   }

   Shows:
   — Animated pulsing icon (two avatar placeholders side by side)
   — "Finding a partner..."
   — "This usually takes just a moment 🏝️"
   — "Cancel" button → onCancel → sends leave_room → redirects to /island

   If waitingSeconds > 60:
   — "No partner found right now"
   — "Try again in a few minutes! Other players might be online soon."
   — Cancel changes to "Back to Island"
   — SEND rule: never show frustration language

─────────────────────────────────────────
PART D — Result Screen
─────────────────────────────────────────

4. components/game/CoopSessionResultScreen.tsx

   interface CoopSessionResultScreenProps {
     outcome: 'win' | 'lose' | 'incomplete'
     starsEarned: number
     groupXPEarned: number
     myXPEarned: number
     totalXP: number
     unlockedZones: string[]
     onPlayAgain: () => void
     onGoToIsland: () => void
   }

   All messages positive — SEND rule:
   — "win"        → "Amazing Teamwork! 🏆" + confetti animation
   — "lose"       → "Great Effort! The tower wobbled — try again! 🏗️"
   — "incomplete" → "Good Try! Your partner disconnected this time. 💪"

   Stars display:
   — Minimum 1 star always shown (even for lose/incomplete)
   — Stars animate in sequence: 0.3s gap between each
   — Earned: bright gold ⭐ | Not earned: soft grey outline

   XP display:
   — "+{myXPEarned} XP each" with bounce animation
   — "Your team earned {groupXPEarned} XP together!"

   If unlockedZones.length > 0:
   — Full-screen flash: "🔓 New Zone Unlocked: {zoneName}!" before result fades in

   Buttons: "Play Again" | "Back to Island"

─────────────────────────────────────────
PART E — API Client
─────────────────────────────────────────

5. lib/api.ts additions

   submitCoopSession(data: {
     game_type: "team_tower",
     mode: "cooperative",
     room_session_id: string,
     outcome: string,
     duration_ms: number
   }): Promise<SessionResult>

   Note: co-op sessions do not submit an action log here.
   Scoring is handled server-side by the WS room.
   This endpoint only confirms the result and fetches updated profile totals.
```

---

## Issue 23-FE — Client Reconnect Handling

```
Title: Phase 3.5-FE — TeamTowerScene client reconnect and reconnecting UI

Read .github/copilot-instructions.md before starting.
Depends on: Issue 23-BE (server reconnect window) merged.

─────────────────────────────────────────
PART A — Client Reconnect (TeamTowerScene.ts update)
─────────────────────────────────────────

1. handleDisconnect():
   — clearInterval(heartbeatInterval)
   — If RoomState == "playing" or "paused":
     — Emit 'game:ui-update' { phase: "reconnecting" }
     — Wait 2000ms
     — Attempt reconnect: new WebSocket(wsUrl)
     — On reconnect success: send join_room (hub routes to existing room via reconnect index)
     — On reconnect failure after 2 attempts:
       Emit 'game:session-end' { outcome: "incomplete", stars: 1, duration_ms: ... }
   — If RoomState == "ended": do nothing

─────────────────────────────────────────
PART B — Reconnecting UI
─────────────────────────────────────────

2. TeamTowerUI.tsx (update) — add phase: "reconnecting"

   When phase === "reconnecting":
   — Warm overlay: "Reconnecting... 🔄"
   — Gentle spinner animation
   — SEND rule: never alarming, never blame language
   — No dismiss button (auto-resolves)

   When phase === "partner_reconnected" (server sends partner_reconnected):
   — Remove disconnected overlay
   — Flash brief "Partner is back! ▶️" banner for 1500ms
   — Resume game state (already synced by server state_update)
```

---

## Issue 24-FE — Student Progress Screen

```
Title: Phase 3.6-FE — Student progress page, skill bars, and achievement badges

Read .github/copilot-instructions.md before starting.
Depends on: Issue 24-BE (stats API and session history API) merged.

─────────────────────────────────────────
PART A — Progress Page
─────────────────────────────────────────

1. app/(student)/progress/page.tsx

   On load:
   — Fetch GET /api/profiles/me/stats (parallel with sessions)
   — Fetch GET /api/profiles/me/sessions?limit=10
   — Show skeleton loader while fetching
   — Render <ProgressScreen stats={stats} recentSessions={sessions} />

─────────────────────────────────────────
PART B — Progress Screen Component
─────────────────────────────────────────

2. components/game/ProgressScreen.tsx

   Use Gemini-generated LeaderboardScreen component if available, otherwise:

   interface ProgressScreenProps {
     stats: ProfileStats
     recentSessions: SessionSummary[]
     isLoading: boolean
   }

   SKILL BARS (top section):
   — Memory Skills:   avg accuracy from memory_cove sessions (0–100%)
   — Focus Skills:    avg attention from focus_forest sessions (0–100%)
   — Team Skills:     cooperative_sessions count as bar (target: 10)
   — Animated fill on load; colour matched to zone theme
   — Label: skill name + percentage / count
   — SEND rule: personal bests only — no comparison to other students

   ACHIEVEMENTS (middle section):
   — Grid of achievement badge cards
   — Earned: full colour + name + "Earned!" label
   — Not yet earned: greyed out + "???" name + hint text
   — Tap earned achievement: tooltip with description
   — Show all possible achievements (locked ones build aspiration)

   RECENT SESSIONS (bottom section):
   — Last 10 sessions as card list
   — Each card: zone emoji + zone name + stars (⭐⭐⭐) + date
   — Tap card: expand to show accuracy + duration
   — "View All" button: loads more via offset pagination

   STREAK (top right corner):
   — streak >= 2: "🔥 {n} day streak!"
   — streak == 1: "🌱 Keep it up!"
   — streak == 0: "▶️ Play today to start a streak!"

─────────────────────────────────────────
PART C — Island Map Link
─────────────────────────────────────────

3. Island map (app/(student)/island/page.tsx update):
   — Add "My Progress 📊" button to PlayerHUD component
   — Navigates to /progress
```

---

---

# PHASE 4 — Analytics, Polish & Beta Launch

---

## Issue 26-FE — Parent Dashboard Components

```
Title: Phase 4.2-FE — Parent dashboard metric cards, sparklines, and session breakdown

Read .github/copilot-instructions.md before starting.
Depends on: Issue 26-BE (analytics overview API with trend data) merged.

─────────────────────────────────────────
PART A — Dashboard Components
─────────────────────────────────────────

1. components/dashboard/ParentDashboard.tsx (full rebuild from Phase 1.5 shell)

   Use Gemini-generated component if available, otherwise:

   interface ParentDashboardProps {
     data: AnalyticsOverview | null
     isLoading: boolean
     error: string | null
   }

   Layout:
   — Header: "Welcome back! Here's {nickname}'s progress this week 🌟"
   — If data_available === false:
     "No games played yet — check back after {nickname} plays!"
     Show all 4 metric card shells (greyed out, placeholder values)

   4 Metric Cards (warm pastel backgrounds):

   ATTENTION CARD (blue/purple):
   — Icon: 🦉  |  Label: "Concentration"
   — Value: attention_score as "X%" or "--"
   — Trend sparkline: 7-day SVG (see TrendSparkline component)
   — Delta badge (see DeltaBadge component)
   — Encouraging sub-label:
     >= 80%: "Excellent focus this week! 🌟"
     >= 60%: "Good concentration 👍"
     >= 40%: "Building focus skills 🌱"
     <  40%: "Keep practising! 💪"

   MEMORY CARD (green):
   — Icon: 🐘  |  Encouraging labels same threshold pattern

   SOCIAL ENGAGEMENT CARD (orange):
   — Icon: 🤝
   — Value: coop_participation_rate as "%"
   — Sessions breakdown mini bar (SessionBreakdown component)

   PROGRESS CARD (yellow/gold):
   — Icon: ⭐
   — Value: total_xp XP
   — Stars this week + sessions this week
   — Next milestone: "X more XP to unlock [zone name]!"

2. components/dashboard/TrendSparkline.tsx (pure SVG — no external chart library)

   interface TrendSparklineProps {
     data: (number | null)[]   // 7 values
     colour: string
     width?: number   // default 120
     height?: number  // default 40
     label: string    // for aria-label
   }

   — Calculate min/max from non-null values; normalise to height
   — Draw smooth bezier SVG path through data points
   — null points: dashed gap in line
   — Dot at each data point (r=2)
   — Filled area under curve with 20% opacity
   — Accessibility: role="img", aria-label, <title> element

3. components/dashboard/SessionBreakdown.tsx (pure SVG — no external chart library)

   interface SessionBreakdownProps {
     data: { zone: string, count: number, colour: string }[]
     totalSessions: number
   }

   — Horizontal bar chart; one bar per zone
   — Memory Cove: blue | Focus Forest: green | Team Tower: orange
   — Bar width proportional to count / totalSessions
   — If 0 sessions: 4px minimum width bar
   — role="img", aria-label summarising the data

4. components/dashboard/DeltaBadge.tsx

   Props: delta: number | null, metric: string

   delta > 0.05  → green badge: "↑ {delta*100 toFixed 0}%"
   delta < -0.05 → amber badge: "↓ {abs(delta)*100 toFixed 0}%"
   otherwise     → grey badge:  "→ Steady"
   null          → grey badge:  "Not enough data"

   SEND rule: never red — amber is the most alarming colour in the entire dashboard.

─────────────────────────────────────────
PART B — Dashboard Page Update
─────────────────────────────────────────

5. app/(dashboard)/overview/page.tsx (update)

   — Fetch GET /api/analytics/overview and GET /api/analytics/sessions in parallel
   — Show skeleton loader while loading
   — Handle error state gracefully
   — Refresh data on window focus (stale-while-revalidate)
   — Add "Last updated: X minutes ago" indicator
```

---

## Issue 27-FE — PDF Export Button

```
Title: Phase 4.3-FE — PDF export button component

Read .github/copilot-instructions.md before starting.
Depends on: Issue 27-BE (GET /api/reports/weekly endpoint) merged.

1. components/dashboard/ExportButton.tsx (upgrade from Phase 1.5)

   interface ExportButtonProps {
     childNickname: string
   }

   States:
   — Idle:    "Download Weekly Report 📄"
   — Loading: "Generating report..." with spinner
   — Success: brief green tick animation → back to idle after 2s
   — Error:   "Report unavailable — try again" (soft amber, not red)

   On click:
   a. Set loading state
   b. fetch GET /api/reports/weekly with credentials: 'include'
   c. PDF blob download:
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `skillisland-report-${date}.pdf`
      a.click()
      URL.revokeObjectURL(url)
   d. Set success state for 2s
   e. On fetch error: set error state
```

---

## Issue 28-FE — Phaser Performance Optimisation & Bundle Analysis

```
Title: Phase 4.4-FE — Phaser object pooling, config optimisation, and Next.js bundle

Read .github/copilot-instructions.md before starting.
Target: 60 FPS sustained on Intel Celeron N4000, 4GB RAM, Chrome 90+.
Depends on: Issue 28-BE (connection pool + perf test script) merged.

─────────────────────────────────────────
PART A — Object Pooling
─────────────────────────────────────────

1. Focus Forest — TargetPool (FocusForestScene.ts update)

   class TargetPool {
     private pool: Phaser.GameObjects.Image[] = []
     private scene: Phaser.Scene

     acquire(targetType: string, x: number, y: number): Phaser.GameObjects.Image
     — Pool has inactive object: reactivate + reposition + return
     — Pool empty: create new object

     release(target: Phaser.GameObjects.Image): void
     — Deactivate + hide; return to pool

     prewarm(count: number): void
     — Called on scene create: pre-create {count} objects (default 20)
   }

   Replace all scene.add.image() calls with pool.acquire()
   Replace all target.destroy() calls with pool.release()
   Call pool.prewarm(20) in scene create

2. Memory Cove — object pool for shape elements
   Same TargetPool pattern; prewarm(8) for max sequence length

─────────────────────────────────────────
PART B — Phaser Config Optimisation
─────────────────────────────────────────

3. apps/web/game/PhaserGame.tsx — replace Phaser config with:

   {
     type: Phaser.WEBGL,
     backgroundColor: '#1a1a2e',
     antialias: false,          // disable for low-end GPU
     pixelArt: false,
     powerPreference: 'low-power',
     fps: {
       target: 60,
       forceSetTimeOut: false,
       smoothStep: true
     },
     scale: {
       mode: Phaser.Scale.RESIZE,
       autoCenter: Phaser.Scale.CENTER_BOTH
     },
     render: {
       batchSize: 2048,
       maxLights: 0
     }
   }

   Fallback: Phaser.CANVAS if WebGL unavailable (automatic via Phaser.AUTO)

─────────────────────────────────────────
PART C — Next.js Bundle
─────────────────────────────────────────

4. apps/web/next.config.ts

   const withBundleAnalyzer = require('@next/bundle-analyzer')({
     enabled: process.env.ANALYZE === 'true'
   })

   Verify with: ANALYZE=true npm run build

   Bundle targets:
   — Phaser (3MB+) only in /game/* route bundles
   — Dashboard pages: zero Phaser in bundle
   — First load JS for /island: < 150KB

5. apps/web/app/(dashboard)/overview/page.tsx — lazy load chart components:

   const TrendSparkline = dynamic(() => import('@/components/dashboard/TrendSparkline'), {
     loading: () => <div className="h-10 bg-gray-100 animate-pulse rounded" />
   })
   const SessionBreakdown = dynamic(() => import('@/components/dashboard/SessionBreakdown'))

6. Verify no Phaser in server bundle:
   grep -r "import.*phaser" apps/web/app/
   — Must return zero results (all Phaser via dynamic import in /game/)

Paste ANALYZE=true bundle screenshot as PR comment.
Paste manual Chromebook FPS test result as PR comment (60 FPS target).
```

---

## Issue 30-FE — Accessibility Audit & SEND Sensory Review

```
Title: Phase 4.6-FE — Accessibility audit (WCAG 2.1 AA) and SEND sensory review

Read .github/copilot-instructions.md before starting.
Depends on: Issue 30-BE (slow_mode migration + API) merged.

─────────────────────────────────────────
PART A — Dashboard WCAG 2.1 AA
─────────────────────────────────────────

1. Install and run axe-core audit:
   npm install --save-dev @axe-core/cli
   npx axe http://localhost:3000/dashboard/overview --tags wcag2a,wcag2aa

   Fix ALL failures before marking complete. Common fixes:

   Colour contrast:
   — Normal text: minimum 4.5:1
   — Large text (>= 18pt or 14pt bold): minimum 3:1
   — Check all metric card values, labels, delta badges
   — Update Tailwind colours as needed

2. Keyboard navigation audit (manual):
   — Every interactive element reachable by Tab
   — Tab order: logical left-to-right, top-to-bottom
   — Focus indicator visible on every element:
     Add to globals.css:
     *:focus-visible {
       outline: 2px solid #4A90E2;
       outline-offset: 2px;
     }
   — No keyboard traps

3. Screen reader fixes:
   — Icon-only buttons: add aria-label="Download report" etc.
   — Form inputs: <label htmlFor> associated with input id
   — Charts/SVGs: role="img" + aria-label describing the data

   Data table fallback for each chart:
   <details>
     <summary>View {metric} data as table</summary>
     <table>
       <thead><tr><th>Date</th><th>Value</th></tr></thead>
       <tbody>{data.map(d => <tr><td>{d.date}</td><td>{d.value}</td></tr>)}</tbody>
     </table>
   </details>

4. components/dashboard/TrendSparkline.tsx (accessibility update)
   — Add <title>{label}: {trendDescription}</title> inside SVG
   — Add <desc>Line chart showing {metric} values over 7 days</desc>
   — Add role="img" to SVG wrapper
   — Data table fallback as above

─────────────────────────────────────────
PART B — Game Zones SEND Sensory Review
─────────────────────────────────────────

5. Animation safety audit — enforce in all Phaser scenes:

   Focus Forest:
   — No animation > 3 flashes/second
   — Butterfly despawn: fade over >= 300ms (not instant)
   — Bee shake: <= 200ms, <= 5px amplitude
   — Spawn tween: fade in over 200ms

   Memory Cove:
   — Wrong answer shake: <= 300ms, <= 8px amplitude
   — Sparkle: soft burst, not strobe

   Team Tower:
   — Tower fall: >= 600ms lean before collapse
   — Win celebration: particle burst (not flashing)

6. Reduced motion support — all Phaser scenes:

   apps/web/game/PhaserGame.tsx:
   const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
   EventBus.emit('game:config', { reducedMotion: prefersReducedMotion })

   In each scene, on 'game:config' received:
   — If reducedMotion:
     Memory Cove:    instant element switch (no slide tween)
     Focus Forest:   targets appear/disappear with fade only (no movement)
     Team Tower:     blocks snap to position (LERP disabled)

7. Slow Mode toggle — setup page:

   app/(student)/setup/page.tsx — add toggle:
   Label: "Show sequences more slowly 🐢"
   Sub-label: "Gives more time to watch each shape"
   — Saves slow_mode: boolean via PATCH /api/profiles/me

   MemoryCoveScene.ts — on scene create, if sessionConfig.slowMode:
   ELEMENT_DISPLAY_MS = 1200  // was 800
   ELEMENT_GAP_MS     = 500   // was 300
   FEEDBACK_PAUSE_MS  = 800   // was 600

─────────────────────────────────────────
PART C — Language Audit
─────────────────────────────────────────

8. Banned language grep:

   grep -ri "wrong\|failed\|game over\|you lost\|error\|bad\|incorrect\|oops" \
     apps/web/components apps/web/app apps/web/game \
     --include="*.tsx" --include="*.ts"

   For every match:
   — UI text shown to student: replace with approved language from game-logic-mvp.md
   — Code variable/comment: acceptable, leave as-is
   — Document all replacements in PR

9. components/game/SessionResultScreen.tsx (0-star audit):
   — 0 stars: NEVER show only 3 grey empty stars
   — Show: "Great effort! Keep practising! 💪" with 1 dim participation star
   — Star graphic: 1 dim gold star, not 3 grey empties

─────────────────────────────────────────
PART D — Accessibility Report
─────────────────────────────────────────

10. Create ACCESSIBILITY.md at repo root:

    # Skill Island Accessibility Report

    ## WCAG 2.1 AA Compliance (Dashboard)
    - Audit date: {date}
    - Tool: axe-core {version}
    - Issues found: {n}
    - Issues fixed: {n}
    - Outstanding known gaps: {list any with justification}

    ## SEND Sensory Review (Game Zones)
    - Reduced motion: supported via OS preference
    - Animation limits: all within 3 flash/second threshold
    - Slow Mode: available for Memory Cove
    - Feedback language: all negative language removed

    ## Recommended Further Review
    - User testing with SEND school students (planned for beta)
    - Review with SEND specialist educator
    - Formal SEND-specific audit post-beta
```

---

## Issue 31-FE — Error Boundaries, BETA.md & Frontend Checklist

```
Title: Phase 4.7-FE — Error boundaries, fallback UI, and BETA.md

Read .github/copilot-instructions.md before starting.
Depends on: Issue 31-BE (production config + deploy) merged.

─────────────────────────────────────────
PART A — Error Boundary
─────────────────────────────────────────

1. components/ErrorBoundary.tsx

   class ErrorBoundary extends React.Component<
     { children: React.ReactNode, fallback?: React.ReactNode },
     { hasError: boolean }
   > {
     componentDidCatch(error, info) {
       console.error('Skill Island error:', error, info)
       // Future: send to error monitoring service
     }
     render() {
       if (this.state.hasError) {
         return this.props.fallback || <DefaultErrorFallback />
       }
       return this.props.children
     }
   }

2. components/DefaultErrorFallback.tsx

   — Warm, calm design — never alarming
   — Message: "Oops! Something went a bit wrong 🌊"
   — Sub-message: "Don't worry — your progress is saved!"
   — Button: "Go back to the Island 🏝️" → redirects to /island
   — Auto-redirect after 8 seconds (visible countdown)
   — SEND rule: no scary technical language, no error codes shown to student

3. Wrap all game pages:

   app/(student)/game/[zone]/page.tsx:
   <ErrorBoundary fallback={<DefaultErrorFallback />}>
     <PhaserGame ... />
   </ErrorBoundary>

─────────────────────────────────────────
PART B — BETA.md for Schools
─────────────────────────────────────────

4. Create BETA.md at repo root:

   # Skill Island — Beta Information for Schools

   ## Welcome
   Thank you for participating in the Skill Island closed beta.
   This document covers everything your IT team and staff need to know.

   ## System Requirements

   Students:
   - Browser: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+
   - Device: Any laptop, Chromebook, or tablet with a modern browser
   - Network: Standard school WiFi; no special ports required except port 443

   Important — WebSocket requirement:
   Skill Island uses WebSocket connections for the Team Tower multiplayer game.
   Please ensure your school network does not block WebSocket connections on port 443.
   If students cannot reach Team Tower, this is the most likely cause.
   Contact us and we can help your IT team verify the configuration.

   ## Staff Accounts
   - Educator accounts: provided by the Skill Island team on request
   - Link educator account to student profiles: contact us with student nicknames
   - Dashboard: read-only; educators see session summaries but not game content

   ## Parent Accounts
   - Parents register at [url]/register with role "parent"
   - Link to student: contact us with parent email and student nickname
   - Parents see: weekly progress summary, session history, downloadable reports
   - Parents cannot see: real-time gameplay, other students' data

   ## Data & Privacy
   What we store:
   - Student nickname and avatar (no real name required)
   - Game session results (scores, accuracy, stars)
   - Behavioural metrics (anonymised reaction times — used only for adaptive features)
   - Parent/educator email address (for login)

   What we do NOT store:
   - Real student names
   - Photos or biometric data
   - Location data
   - Any communication content (there is no student-to-student communication)

   Data retention: session data retained for 2 years; deleted on request.
   Data location: EU-based servers.
   GDPR: contact [email] for data access or deletion requests.

   ## Known Limitations in Beta
   - Pattern Plateau and Community Hub zones are coming soon (locked in current build)
   - PDF reports show data only after the first 24 hours of play
   - Team Tower matchmaking requires another student to be online simultaneously
   - Mobile portrait supported; landscape is optimal for Team Tower

   ## Feedback
   Please report issues to: [email]
   Include: student age, device type, browser, description of what happened.

   Thank you for helping make Skill Island the best it can be for SEND students. 🏝️

─────────────────────────────────────────
PART C — Frontend  Checklist
─────────────────────────────────────────

Before this can be completed:

- [ ] ErrorBoundary wrapping all game pages (student-facing)
- [ ] DefaultErrorFallback never shows raw error messages or stack traces
- [ ] No console.log statements in production build (npm run build clean)
- [ ] BETA.md complete and reviewed by a non-technical reader
- [ ] ACCESSIBILITY.md complete (from Issue 30-FE)
- [ ] npm run build — no TypeScript errors, no warnings
- [ ] Security workflow — zero HIGH/CRITICAL npm vulnerabilities
- [ ] One full student play session verified manually on beta server
  (Memory Cove solo + Focus Forest solo + Team Tower co-op)
```

---
