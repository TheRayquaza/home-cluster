package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"games/internal/auth"
	"games/internal/config"
	"games/internal/db"
	"games/internal/games"
	"games/internal/hub"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DBURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("db migrate: %v", err)
	}

	h := hub.NewHub(database)

	mux := http.NewServeMux()

	// Health probe — no DB hit, instant 200
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Auth routes
	mux.HandleFunc("POST /api/auth/register", handleRegister(database))
	mux.HandleFunc("POST /api/auth/login", handleLogin(database, cfg.JWTSecret))
	mux.HandleFunc("POST /api/auth/logout", handleLogout())
	mux.Handle("GET /api/me", chain(handleMe(), auth.RequireAuth(cfg.JWTSecret)))

	// Public game routes
	mux.HandleFunc("GET /api/daily", handleDaily(database))
	mux.HandleFunc("GET /api/games", handleGames())
	mux.HandleFunc("GET /api/leaderboard", handleLeaderboard(database))

	// Room routes (auth required)
	mux.Handle("POST /api/rooms", chain(handleCreateRoom(h, database), auth.RequireAuth(cfg.JWTSecret)))
	mux.Handle("GET /api/rooms/{code}", chain(handleGetRoom(h), auth.RequireAuth(cfg.JWTSecret)))

	// Admin routes
	mux.Handle("PUT /api/admin/daily", chain(handleSetDaily(database), auth.RequireGamemaster(cfg.JWTSecret)))
	mux.Handle("GET /api/admin/sessions", chain(handleAdminSessions(database), auth.RequireGamemaster(cfg.JWTSecret)))

	// WebSocket
	mux.HandleFunc("GET /ws", handleWS(h, cfg.JWTSecret, cfg.AllowedOrigin))

	handler := corsMiddleware(cfg.AllowedOrigin)(mux)

	log.Printf("server listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

// chain applies middleware in order (first middleware = outermost).
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// corsMiddleware sets CORS headers for the allowed origin.
func corsMiddleware(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// jsonOK writes a JSON 200 response.
func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("jsonOK encode: %v", err)
	}
}

// jsonErr writes a JSON error response.
func jsonErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// ---- Auth handlers ----

func handleRegister(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
		body.Username = strings.TrimSpace(body.Username)
		if body.Username == "" || body.Password == "" {
			jsonErr(w, http.StatusBadRequest, "username and password are required")
			return
		}
		if len(body.Username) > 50 {
			jsonErr(w, http.StatusBadRequest, "username too long")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("bcrypt error: %v", err)
			jsonErr(w, http.StatusInternalServerError, "internal error")
			return
		}

		var userID string
		err = database.QueryRow(
			`INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`,
			body.Username, string(hash),
		).Scan(&userID)
		if err != nil {
			if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
				jsonErr(w, http.StatusConflict, "username already taken")
				return
			}
			log.Printf("register error: %v", err)
			jsonErr(w, http.StatusInternalServerError, "internal error")
			return
		}

		jsonOK(w, map[string]string{"id": userID, "username": body.Username})
	}
}

func handleLogin(database *sql.DB, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonErr(w, http.StatusBadRequest, "invalid request body")
			return
		}

		var userID, hash, role string
		err := database.QueryRow(
			`SELECT id, password_hash, role FROM users WHERE username = $1`,
			body.Username,
		).Scan(&userID, &hash, &role)
		if err == sql.ErrNoRows {
			jsonErr(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		if err != nil {
			log.Printf("login query: %v", err)
			jsonErr(w, http.StatusInternalServerError, "internal error")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)); err != nil {
			jsonErr(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		token, err := auth.GenerateToken(secret, userID, body.Username, role)
		if err != nil {
			log.Printf("generate token: %v", err)
			jsonErr(w, http.StatusInternalServerError, "internal error")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			HttpOnly: true,
			Path:     "/",
			MaxAge:   7 * 24 * 3600,
			SameSite: http.SameSiteLaxMode,
		})
		jsonOK(w, map[string]string{
			"id":       userID,
			"username": body.Username,
			"role":     role,
		})
	}
}

func handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			HttpOnly: true,
			Path:     "/",
			MaxAge:   -1,
		})
		jsonOK(w, map[string]string{"message": "logged out"})
	}
}

func handleMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.UserFromContext(r.Context())
		if claims == nil {
			jsonErr(w, http.StatusUnauthorized, "not authenticated")
			return
		}
		jsonOK(w, map[string]string{
			"id":       claims.UserID,
			"username": claims.Username,
			"role":     claims.Role,
		})
	}
}

// ---- Game routes ----

func handleGames() http.HandlerFunc {
	type gameInfo = games.GameMeta
	result := make([]gameInfo, 0, len(games.Meta))
	for _, m := range games.Meta {
		result = append(result, m)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return func(w http.ResponseWriter, r *http.Request) {
		jsonOK(w, result)
	}
}

func handleDaily(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameID, err := getDailyGame(database)
		if err != nil {
			log.Printf("daily game: %v", err)
			jsonErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		meta, ok := games.Meta[gameID]
		if !ok {
			jsonErr(w, http.StatusInternalServerError, "unknown game")
			return
		}
		jsonOK(w, meta)
	}
}

// getDailyGame returns today's game ID, inserting a deterministic default if absent.
func getDailyGame(database *sql.DB) (string, error) {
	today := time.Now().Format("2006-01-02")

	var gameID string
	err := database.QueryRow(`SELECT game_id FROM daily_games WHERE date = $1`, today).Scan(&gameID)
	if err == nil {
		return gameID, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("query daily_games: %w", err)
	}

	// Deterministic selection from date hash
	gameIDs := gameIDList()
	h := sha256.Sum256([]byte(today))
	idx := int(binary.BigEndian.Uint64(h[:8])) % len(gameIDs)
	if idx < 0 {
		idx = -idx
	}
	gameID = gameIDs[idx]

	_, insertErr := database.Exec(
		`INSERT INTO daily_games (date, game_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		today, gameID,
	)
	if insertErr != nil {
		log.Printf("insert daily_games: %v", insertErr)
	}

	return gameID, nil
}

func gameIDList() []string {
	ids := make([]string, 0, len(games.Registry))
	for id := range games.Registry {
		ids = append(ids, id)
	}
	return ids
}

func handleLeaderboard(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := database.Query(`
			SELECT
				u.username,
				COUNT(CASE WHEN (gs.winner_idx = 0 AND gs.player1_id = u.id)
				             OR (gs.winner_idx = 1 AND gs.player2_id = u.id) THEN 1 END) AS wins,
				COUNT(CASE WHEN (gs.winner_idx = 1 AND gs.player1_id = u.id)
				             OR (gs.winner_idx = 0 AND gs.player2_id = u.id) THEN 1 END) AS losses,
				COUNT(CASE WHEN gs.winner_idx = -1 THEN 1 END) AS draws
			FROM users u
			JOIN game_sessions gs ON (gs.player1_id = u.id OR gs.player2_id = u.id)
			GROUP BY u.username
			ORDER BY wins DESC, losses ASC
			LIMIT 20
		`)
		if err != nil {
			log.Printf("leaderboard query: %v", err)
			jsonErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		defer rows.Close()

		type entry struct {
			Username string `json:"username"`
			Wins     int    `json:"wins"`
			Losses   int    `json:"losses"`
			Draws    int    `json:"draws"`
		}
		result := []entry{}
		for rows.Next() {
			var e entry
			if err := rows.Scan(&e.Username, &e.Wins, &e.Losses, &e.Draws); err != nil {
				log.Printf("leaderboard scan: %v", err)
				continue
			}
			result = append(result, e)
		}
		jsonOK(w, result)
	}
}

// ---- Room routes ----

func handleCreateRoom(h *hub.Hub, database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			GameID string         `json:"game_id"`
			Config map[string]any `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.GameID == "" {
			gameID, err := getDailyGame(database)
			if err != nil {
				jsonErr(w, http.StatusInternalServerError, "internal error")
				return
			}
			body.GameID = gameID
		}
		if body.Config == nil {
			body.Config = map[string]any{}
		}

		code, err := h.CreateRoom(body.GameID, body.Config)
		if err != nil {
			jsonErr(w, http.StatusBadRequest, err.Error())
			return
		}

		jsonOK(w, map[string]string{
			"code":    code,
			"game_id": body.GameID,
		})
	}
}

func handleGetRoom(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		room := h.GetRoom(strings.ToUpper(code))
		if room == nil {
			jsonErr(w, http.StatusNotFound, "room not found")
			return
		}
		count := 0
		for _, p := range room.Players {
			if p != nil {
				count++
			}
		}
		jsonOK(w, map[string]any{
			"code":           room.Code,
			"game_id":        room.GameID,
			"players_joined": count,
		})
	}
}

// ---- Admin routes ----

func handleSetDaily(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			GameID string `json:"game_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.GameID == "" {
			jsonErr(w, http.StatusBadRequest, "game_id is required")
			return
		}

		if _, ok := games.Registry[body.GameID]; !ok {
			jsonErr(w, http.StatusBadRequest, "unknown game_id")
			return
		}

		today := time.Now().Format("2006-01-02")
		_, err := database.Exec(
			`INSERT INTO daily_games (date, game_id) VALUES ($1, $2)
			 ON CONFLICT (date) DO UPDATE SET game_id = EXCLUDED.game_id`,
			today, body.GameID,
		)
		if err != nil {
			log.Printf("set daily: %v", err)
			jsonErr(w, http.StatusInternalServerError, "internal error")
			return
		}

		jsonOK(w, map[string]string{"date": today, "game_id": body.GameID})
	}
}

func handleAdminSessions(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := database.Query(`
			SELECT gs.id, gs.game_id, u1.username, u2.username, gs.winner_idx, gs.played_at
			FROM game_sessions gs
			LEFT JOIN users u1 ON gs.player1_id = u1.id
			LEFT JOIN users u2 ON gs.player2_id = u2.id
			ORDER BY gs.played_at DESC
			LIMIT 100
		`)
		if err != nil {
			log.Printf("admin sessions query: %v", err)
			jsonErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		defer rows.Close()

		type sessionRow struct {
			ID        string     `json:"id"`
			GameID    string     `json:"game_id"`
			Player1   string     `json:"player1"`
			Player2   string     `json:"player2"`
			WinnerIdx *int       `json:"winner_idx"`
			PlayedAt  time.Time  `json:"played_at"`
		}
		result := []sessionRow{}
		for rows.Next() {
			var s sessionRow
			if err := rows.Scan(&s.ID, &s.GameID, &s.Player1, &s.Player2, &s.WinnerIdx, &s.PlayedAt); err != nil {
				log.Printf("admin sessions scan: %v", err)
				continue
			}
			result = append(result, s)
		}
		jsonOK(w, result)
	}
}

// ---- WebSocket handler ----

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS is handled at the HTTP layer
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleWS(h *hub.Hub, jwtSecret, _ string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			if cookie, err := r.Cookie("auth_token"); err == nil {
				token = cookie.Value
			}
		}
		if token == "" {
			http.Error(w, `{"error":"missing token"}`, http.StatusUnauthorized)
			return
		}

		// Validate inline (can't use middleware for WS upgrade)
		claims := validateWSToken(jwtSecret, token)
		if claims == nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade: %v", err)
			return
		}

		client := hub.NewClient(claims.UserID, claims.Username, conn)
		defer func() {
			h.RemoveClient(client)
			conn.Close()
		}()

		log.Printf("ws: %s connected", claims.Username)

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("ws read error for %s: %v", claims.Username, err)
				}
				break
			}
			h.HandleMessage(client, msg)
		}

		log.Printf("ws: %s disconnected", claims.Username)
	}
}

func validateWSToken(secret, tokenStr string) *auth.Claims {
	// We reuse the same logic as the middleware but return nil on error
	// This is a thin wrapper to avoid importing JWT directly in main
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	var claims *auth.Claims
	handler := auth.RequireAuth(secret)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		claims = auth.UserFromContext(r.Context())
	}))
	rr := &noopResponseWriter{}
	handler.ServeHTTP(rr, req)
	return claims
}

// noopResponseWriter discards writes (used for token validation side-effect only).
type noopResponseWriter struct {
	header http.Header
	code   int
}

func (n *noopResponseWriter) Header() http.Header {
	if n.header == nil {
		n.header = make(http.Header)
	}
	return n.header
}
func (n *noopResponseWriter) Write(b []byte) (int, error) { return len(b), nil }
func (n *noopResponseWriter) WriteHeader(code int)        { n.code = code }
