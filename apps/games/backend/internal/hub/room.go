package hub

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"games/internal/games"
	"games/internal/games/wordle"

	"github.com/gorilla/websocket"
)

// Client represents a connected WebSocket player.
type Client struct {
	UserID   string
	Username string
	conn     *websocket.Conn
	mu       sync.Mutex
}

func NewClient(userID, username string, conn *websocket.Conn) *Client {
	return &Client{
		UserID:   userID,
		Username: username,
		conn:     conn,
	}
}

// Send sends a message to the client, safe for concurrent use.
func (c *Client) Send(msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// Room manages a two-player game session.
type Room struct {
	Code    string
	GameID  string
	Game    games.Game
	Config  map[string]any
	State   games.State
	Players [2]*Client
	Inputs  [2]*games.Action
	mu      sync.Mutex
	done    chan struct{}
}

// Message types for WS communication.
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func newRoom(code, gameID string, g games.Game, config map[string]any) *Room {
	return &Room{
		Code:   code,
		GameID: gameID,
		Game:   g,
		Config: config,
		done:   make(chan struct{}),
	}
}

// broadcast sends the current game state to both players.
func (r *Room) broadcast(db *sql.DB) {
	state := r.State

	// Strip secret for Wordle
	if r.GameID == "wordle" {
		state = wordle.StripSecret(state)
	}

	msg := map[string]any{
		"type":    "state",
		"payload": state,
	}

	for _, p := range r.Players {
		if p != nil {
			if err := p.Send(msg); err != nil {
				log.Printf("broadcast send error: %v", err)
			}
		}
	}

	if r.Game.IsOver(r.State) {
		r.sendGameOver(db)
	}
}

// broadcastToPlayer sends current state to a single player slot (for reconnects).
func (r *Room) broadcastToPlayer(idx int, _ *sql.DB) {
	p := r.Players[idx]
	if p == nil {
		return
	}
	state := r.State
	if r.GameID == "wordle" {
		state = wordle.StripSecret(state)
	}
	msg := map[string]any{"type": "state", "payload": state}
	if err := p.Send(msg); err != nil {
		log.Printf("broadcastToPlayer send error: %v", err)
	}
}

// sendGameOver notifies both players and records the session.
func (r *Room) sendGameOver(db *sql.DB) {
	winner := r.Game.Winner(r.State)
	msg := map[string]any{
		"type": "game_over",
		"payload": map[string]any{
			"winner": winner,
		},
	}
	for _, p := range r.Players {
		if p != nil {
			if err := p.Send(msg); err != nil {
				log.Printf("game_over send error: %v", err)
			}
		}
	}

	// Record session in DB
	if db != nil && r.Players[0] != nil && r.Players[1] != nil {
		go recordSession(db, r.GameID, r.Players[0].UserID, r.Players[1].UserID, winner)
	}

	// Signal done
	select {
	case <-r.done:
	default:
		close(r.done)
	}
}

func recordSession(db *sql.DB, gameID, p1ID, p2ID string, winner int) {
	var winnerIdx *int
	if winner >= 0 {
		w := winner
		winnerIdx = &w
	}
	_, err := db.Exec(
		`INSERT INTO game_sessions (game_id, player1_id, player2_id, winner_idx) VALUES ($1, $2, $3, $4)`,
		gameID, p1ID, p2ID, winnerIdx,
	)
	if err != nil {
		log.Printf("record session error: %v", err)
	}
}

func (r *Room) initState(seed int64) games.State {
	if cg, ok := r.Game.(games.ConfigurableGame); ok {
		return cg.InitWithConfig(seed, r.Config)
	}
	return r.Game.Init(seed)
}

// startTurnBased initializes the turn-based game state.
func (r *Room) startTurnBased(seed int64) {
	r.State = r.initState(seed)
	r.broadcast(nil)
}

// startRealTime initializes and runs the real-time tick loop.
// Must be called in its own goroutine.
func (r *Room) startRealTime(seed int64, db *sql.DB, fps float64) {
	r.State = r.initState(seed)
	r.broadcast(db)

	tg, ok := r.Game.(games.TickableGame)
	if !ok {
		log.Printf("room %s: game %s is not tickable", r.Code, r.GameID)
		return
	}

	ticker := time.NewTicker(time.Duration(float64(time.Second) / fps))
	defer ticker.Stop()

	dt := 1.0 / fps
	for {
		select {
		case <-r.done:
			return
		case <-ticker.C:
			r.mu.Lock()
			r.State = tg.Tick(r.State, r.Inputs, dt)
			r.mu.Unlock()
			r.broadcast(db)
			if r.Game.IsOver(r.State) {
				return
			}
		}
	}
}

// handleAction processes a player action message.
func (r *Room) handleAction(db *sql.DB, playerIdx int, payload json.RawMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Game.IsOver(r.State) {
		return
	}

	var action games.Action
	if err := json.Unmarshal(payload, &action); err != nil {
		log.Printf("room %s: invalid action JSON: %v", r.Code, err)
		return
	}

	if r.Game.RealTime() {
		// Real-time: store input for tick loop
		r.Inputs[playerIdx] = &action
	} else {
		// Turn-based: apply action
		newState, err := r.Game.Apply(r.State, playerIdx, action)
		if err != nil {
			// Send error back to the acting player
			if r.Players[playerIdx] != nil {
				errMsg := map[string]any{
					"type":    "error",
					"payload": map[string]any{"message": err.Error()},
				}
				if sendErr := r.Players[playerIdx].Send(errMsg); sendErr != nil {
					log.Printf("error send failed: %v", sendErr)
				}
			}
			return
		}
		r.State = newState
		r.broadcast(db)
	}
}

// sendJoined notifies a player of their slot and the game.
func (r *Room) sendJoined(playerIdx int) {
	if r.Players[playerIdx] == nil {
		return
	}
	msg := map[string]any{
		"type": "joined",
		"payload": map[string]any{
			"playerIdx": playerIdx,
			"game":      r.GameID,
			"code":      r.Code,
		},
	}
	if err := r.Players[playerIdx].Send(msg); err != nil {
		log.Printf("sendJoined error: %v", err)
	}
}

// sendError sends an error message to a specific player.
func (r *Room) sendError(playerIdx int, message string) {
	if r.Players[playerIdx] == nil {
		return
	}
	msg := map[string]any{
		"type":    "error",
		"payload": map[string]any{"message": message},
	}
	if err := r.Players[playerIdx].Send(msg); err != nil {
		log.Printf("sendError error: %v", err)
	}
}

// sendErrorDirect sends an error to a client not yet assigned to a slot.
func sendErrorDirect(c *Client, message string) {
	msg := map[string]any{
		"type":    "error",
		"payload": map[string]any{"message": message},
	}
	if err := c.Send(msg); err != nil {
		log.Printf("sendErrorDirect: %v", err)
	}
}

// playerIndex returns the slot index of a client in this room (-1 if not found).
func (r *Room) playerIndex(c *Client) int {
	for i, p := range r.Players {
		if p == c {
			return i
		}
	}
	return -1
}

// fps returns the tick rate for the given game.
func gameFPS(gameID string) float64 {
	switch gameID {
	case "lightcycles":
		return 10
	case "pong":
		return 60
	}
	return 60
}

// seed creates a seed from the current time.
func newSeed() int64 {
	return time.Now().UnixNano()
}

// String is a helper for debugging.
func (r *Room) String() string {
	return fmt.Sprintf("Room{code=%s, game=%s}", r.Code, r.GameID)
}
