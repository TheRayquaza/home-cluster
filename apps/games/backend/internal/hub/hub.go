package hub

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"

	"games/internal/games"
)

// Hub manages all active rooms and WebSocket clients.
type Hub struct {
	rooms map[string]*Room
	mu    sync.RWMutex
	db    *sql.DB
}

// NewHub creates a new Hub backed by the given database.
func NewHub(db *sql.DB) *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
		db:    db,
	}
}

// CreateRoom creates a new room for the given game ID, returning the 4-char room code.
func (h *Hub) CreateRoom(gameID string, config map[string]any) (string, error) {
	g, ok := games.Registry[gameID]
	if !ok {
		return "", fmt.Errorf("unknown game: %s", gameID)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	code := h.generateCode()
	room := newRoom(code, gameID, g, config)
	h.rooms[code] = room

	log.Printf("hub: created room %s for game %s (config=%v)", code, gameID, config)
	return code, nil
}

// GetRoom returns the room for the given code, or nil.
func (h *Hub) GetRoom(code string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[code]
}

// JoinRoom adds a client to an existing room, returning their player index.
func (h *Hub) JoinRoom(code string, client *Client) (int, error) {
	h.mu.Lock()
	room, ok := h.rooms[code]
	h.mu.Unlock()
	if !ok {
		return -1, errors.New("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// Assign to first empty slot
	for i, p := range room.Players {
		if p == nil {
			room.Players[i] = client
			log.Printf("hub: player %s joined room %s as player %d", client.Username, code, i)
			return i, nil
		}
	}

	return -1, errors.New("room is full")
}

// HandleMessage routes a WebSocket message from a client.
func (h *Hub) HandleMessage(client *Client, raw []byte) {
	var msg WSMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		sendErrorDirect(client, "invalid message format")
		return
	}

	switch msg.Type {
	case "join":
		h.handleJoin(client, msg.Payload)
	case "ready":
		h.handleReady(client)
	case "action":
		h.handleAction(client, msg.Payload)
	default:
		sendErrorDirect(client, fmt.Sprintf("unknown message type: %s", msg.Type))
	}
}

func (h *Hub) handleJoin(client *Client, payload json.RawMessage) {
	var p struct {
		Room string `json:"room"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || p.Room == "" {
		sendErrorDirect(client, "join: missing 'room'")
		return
	}

	idx, err := h.JoinRoom(p.Room, client)
	if err != nil {
		sendErrorDirect(client, err.Error())
		return
	}

	room := h.GetRoom(p.Room)
	room.sendJoined(idx)

	room.mu.Lock()
	defer room.mu.Unlock()

	// Game already running (reconnect): resend state to this player
	if room.State != nil {
		room.broadcastToPlayer(idx, nil)
		return
	}

	// Both players now present: auto-start
	if room.Players[0] != nil && room.Players[1] != nil {
		h.startGame(room)
	}
}

func (h *Hub) handleReady(client *Client) {
	// ready is a no-op — game auto-starts on second join.
	// Kept for protocol compatibility; clients may still send it.
}

func (h *Hub) startGame(room *Room) {
	seed := newSeed()
	if room.Game.RealTime() {
		fps := gameFPS(room.GameID)
		go room.startRealTime(seed, h.db, fps)
	} else {
		room.startTurnBased(seed)
	}

	log.Printf("hub: game started in room %s (game=%s)", room.Code, room.GameID)
}

func (h *Hub) handleAction(client *Client, payload json.RawMessage) {
	room := h.findRoomForClient(client)
	if room == nil {
		sendErrorDirect(client, "you are not in a room")
		return
	}

	playerIdx := room.playerIndex(client)
	if playerIdx < 0 {
		sendErrorDirect(client, "not a player in this room")
		return
	}

	if room.State == nil {
		room.sendError(playerIdx, "game has not started yet")
		return
	}

	room.handleAction(h.db, playerIdx, payload)
}

// RemoveClient removes a client from whatever room they're in.
func (h *Hub) RemoveClient(client *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, room := range h.rooms {
		room.mu.Lock()
		for i, p := range room.Players {
			if p == client {
				room.Players[i] = nil
				log.Printf("hub: player %s left room %s", client.Username, room.Code)
			}
		}
		room.mu.Unlock()
	}
}

// findRoomForClient returns the room a client is currently in, or nil.
func (h *Hub) findRoomForClient(client *Client) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, room := range h.rooms {
		for _, p := range room.Players {
			if p == client {
				return room
			}
		}
	}
	return nil
}

// generateCode produces a unique 4-char uppercase room code.
// Caller must hold h.mu (write lock).
func (h *Hub) generateCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	for {
		b := make([]byte, 4)
		for i := range b {
			b[i] = chars[rand.Intn(len(chars))]
		}
		code := string(b)
		if _, exists := h.rooms[code]; !exists {
			return code
		}
	}
}
