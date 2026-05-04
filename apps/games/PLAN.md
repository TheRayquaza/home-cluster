# Games Platform — Plan

2-player in-browser mini-games. One game per day (gamemaster picks). Go backend, React frontend, PostgreSQL.

---

## Stack

| Layer | Tech |
|-------|------|
| Backend | Go 1.25, stdlib HTTP, gorilla/websocket |
| Frontend | React 19 + Vite + TypeScript |
| Database | PostgreSQL (sidecar in K8s) |
| Auth | JWT via HTTP-only cookie |
| K8s | Namespace `games`, HTTPRoute `games.internal.rayq.app` |

---

## Directory Layout

```
apps/games/
├── backend/
│   ├── main.go
│   ├── go.mod
│   ├── Dockerfile
│   └── internal/
│       ├── config/      env-based config
│       ├── db/          postgres connection + migrations
│       ├── models/      User, DailyGame, GameSession
│       ├── auth/        JWT middleware (cookie-based)
│       ├── hub/         WebSocket hub + room management
│       └── games/
│           ├── interface.go    Game interface
│           ├── registry.go     game_id → Game map
│           ├── tictactoe/
│           ├── connect4/
│           ├── memory/
│           ├── nim/
│           ├── pong/
│           ├── lightcycles/
│           └── wordle/
└── frontend/
    ├── vite.config.ts
    └── src/
        ├── api/         REST + WebSocket client
        ├── hooks/       useWebSocket, useGame, useAuth
        ├── components/
        │   ├── games/   one component per game
        │   └── ui/      Lobby, RoomCode, Scoreboard
        └── pages/       Login, Home, Game, Admin
```

---

## Game Interface (Go)

```go
type Game interface {
    ID()        string
    Name()      string
    RealTime()  bool          // pong + lightcycles only
    Init(seed int64) State
    Apply(state State, playerIdx int, action Action) (State, error)
    IsOver(state State) bool
    Winner(state State) int   // 0=p1, 1=p2, -1=draw, -2=ongoing
}

// Real-time games also implement:
type TickableGame interface {
    Game
    Tick(state State, dt float64) State
}
```

---

## WebSocket Protocol

```
Client → Server
  { "type": "join",   "payload": { "room": "XXXX" } }
  { "type": "ready" }
  { "type": "action", "payload": { /* game-specific */ } }

Server → Client
  { "type": "joined",    "payload": { "playerIdx": 0|1, "game": "tictactoe" } }
  { "type": "state",     "payload": { /* full game state */ } }
  { "type": "game_over", "payload": { "winner": 0|1|-1 } }
  { "type": "error",     "payload": { "message": "..." } }
```

---

## REST API

```
POST /api/auth/login
POST /api/auth/register
POST /api/auth/logout
GET  /api/me

GET  /api/daily          → today's game
GET  /api/games          → all available games
GET  /api/leaderboard

POST /api/rooms          → create room (uses today's game)
GET  /api/rooms/:code

PUT  /api/admin/daily    { game_id }   gamemaster only
GET  /api/admin/sessions

WS   /ws?room=XXXX
```

---

## PostgreSQL Schema

```sql
CREATE TABLE users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username      VARCHAR(50) UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  role          VARCHAR(20) NOT NULL DEFAULT 'player',
  created_at    TIMESTAMP DEFAULT NOW()
);

CREATE TABLE daily_games (
  date    DATE PRIMARY KEY,
  game_id VARCHAR(50) NOT NULL
);

CREATE TABLE game_sessions (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  game_id    VARCHAR(50) NOT NULL,
  player1_id UUID REFERENCES users(id),
  player2_id UUID REFERENCES users(id),
  winner_idx INT,   -- 0=p1, 1=p2, -1=draw
  played_at  TIMESTAMP DEFAULT NOW()
);
```

---

## The 7 Games

### 1. Tic Tac Toe — turn-based
- 3×3 grid, X vs O
- Action: `{ "cell": 0-8 }`
- Win: 3-in-a-row (8 patterns)

### 2. Connect 4 — turn-based
- 7 cols × 6 rows, pieces fall to lowest empty row
- Action: `{ "col": 0-6 }`
- Win: 4-in-a-row horizontal / vertical / diagonal

### 3. Memory Match — turn-based
- 16 cards (8 pairs), face-down, shuffled
- Action: `{ "card": 0-15 }` — reveal 2 per turn; match = extra turn
- Win: most pairs when all matched

### 4. Nim — turn-based
- 21 sticks; take 1–3 per turn; last to take **loses**
- Action: `{ "take": 1|2|3 }`

### 5. Pong — real-time (60 fps server loop)
- Ball + 2 paddles; server-authoritative physics
- Input: `{ "dir": "up"|"down"|"stop" }`
- Win: first to 7 points

### 6. Light Cycles — real-time (10 fps server loop)
- 40×30 grid; each player leaves a trail; hit trail = lose
- Input: `{ "dir": "up"|"down"|"left"|"right" }`
- Win: opponent crashes first

### 7. Wordle — async (both guess simultaneously)
- Same secret 5-letter word for both players
- Action: `{ "guess": "CRANE" }` — server responds with color hints
- Win: first to solve; tie-break = fewer attempts
- Word list: ~2000 common words embedded in binary

---

## Daily Game Logic

```go
// GET /api/daily
func getDailyGame(db *sql.DB, registry map[string]Game) (string, error) {
    today := time.Now().Format("2006-01-02")
    // try SELECT game_id FROM daily_games WHERE date = today
    // if not found: hash today's date → deterministic index → insert + return
}
```

Gamemaster can override via `PUT /api/admin/daily`.

---

## Room Flow

1. P1 calls `POST /api/rooms` → receives 4-char `room_code`
2. P1 shares code with P2
3. Both connect: `WS /ws?room=XXXX`
4. Both send `{ "type": "ready" }` → server starts game, broadcasts initial state
5. Game runs until `game_over` → server records session in DB

---

## Build Sequence

| Step | What |
|------|------|
| 1 | `backend/internal/db` + schema + migrations |
| 2 | `backend/internal/auth` (JWT, same as kommande) |
| 3 | `backend/internal/games/interface.go` + registry |
| 4 | 4 turn-based games: tictactoe, connect4, nim, memory |
| 5 | `backend/internal/hub` (rooms, turn dispatch) |
| 6 | 2 real-time games: pong, lightcycles + tick loop |
| 7 | wordle (embedded word list + hint logic) |
| 8 | REST handlers + WS handler wired in main.go |
| 9 | Frontend: auth + lobby + room pages |
| 10 | Frontend: 7 game components |
| 11 | Dockerfile + docker-compose.yaml |
| 12 | k8s/games/ manifests + k8s/apps/games.yaml |

---

## K8s Layout

```
k8s/games/
├── kustomization.yaml
├── namespace.yaml
├── deployment.yaml    # games container + postgres sidecar
├── pvc.yaml           # 5Gi for postgres data
├── configmap.yaml     # DB_URL, PORT, ALLOWED_ORIGIN
├── secret.yaml        # JWT_SECRET, DB_PASSWORD
├── service.yaml       # ClusterIP :8080
├── route.yaml         # HTTPRoute games.internal.rayq.app
└── certificate.yaml
```

ArgoCD app: `k8s/apps/games.yaml` (sync-wave 1, same as kommande).

---

## Verification Checklist

- [ ] `docker-compose up` → backend :8080, frontend :5173
- [ ] Register + login → JWT cookie set
- [ ] `GET /api/daily` → game_id returned, row inserted in `daily_games`
- [ ] Create room → 4-char code returned
- [ ] Join room in two tabs → both receive `joined` message
- [ ] Play all 7 games end-to-end
- [ ] Gamemaster override daily game
- [ ] Leaderboard reflects finished sessions
