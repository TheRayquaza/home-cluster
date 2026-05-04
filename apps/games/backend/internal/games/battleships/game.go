package battleships

import (
	"errors"
	"math/rand"
	"time"
)

type State = map[string]any
type Action = map[string]any

// shipSizes is the fleet configuration: carrier, battleship, cruiser, submarine, destroyer.
var shipSizes = []int{5, 4, 3, 3, 2}

const totalShipCells = 17 // sum of shipSizes

// Game implements Battleships (turn-based, placement then combat).
type Game struct{}

func (g *Game) ID() string     { return "battleships" }
func (g *Game) Name() string   { return "Battleships" }
func (g *Game) RealTime() bool { return false }

func (g *Game) Init(seed int64) State {
	rng := rand.New(rand.NewSource(seed))
	board0 := randomBoard(rng)
	board1 := randomBoard(rng)

	return State{
		"phase":      "placement",
		"boards":     []any{board0[:], board1[:]},
		"shots":      []any{make([]int, 100), make([]int, 100)},
		"ready":      []bool{false, false},
		"ships_left": []int{totalShipCells, totalShipCells},
		"turn":       0,
		"winner":     -2,
		"ship_sizes": shipSizes,
	}
}

// randomBoard places ships of sizes [5,4,3,3,2] randomly without overlap.
func randomBoard(rng *rand.Rand) [100]int {
	var board [100]int
	for _, size := range shipSizes {
		for {
			horizontal := rng.Intn(2) == 0
			var x, y int
			if horizontal {
				x = rng.Intn(10 - size + 1)
				y = rng.Intn(10)
			} else {
				x = rng.Intn(10)
				y = rng.Intn(10 - size + 1)
			}
			// Check no overlap
			overlap := false
			for i := 0; i < size; i++ {
				var idx int
				if horizontal {
					idx = y*10 + x + i
				} else {
					idx = (y+i)*10 + x
				}
				if board[idx] != 0 {
					overlap = true
					break
				}
			}
			if overlap {
				continue
			}
			// Place ship
			for i := 0; i < size; i++ {
				var idx int
				if horizontal {
					idx = y*10 + x + i
				} else {
					idx = (y+i)*10 + x
				}
				board[idx] = 1
			}
			break
		}
	}
	return board
}

func (g *Game) Apply(state State, playerIdx int, action Action) (State, error) {
	if g.IsOver(state) {
		return state, errors.New("game is already over")
	}

	phase, _ := state["phase"].(string)
	boards := cloneBoards(state["boards"])
	shots := cloneShots(state["shots"])
	ready := cloneReady(state["ready"])
	shipsLeft := cloneInts2(state["ships_left"])
	turn := toInt(state["turn"])
	winner := toInt(state["winner"])

	actionType, _ := action["action"].(string)

	switch phase {
	case "placement":
		switch actionType {
		case "ready":
			ready[playerIdx] = true
			if ready[0] && ready[1] {
				phase = "combat"
			}
		case "randomize":
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			newBoard := randomBoard(rng)
			boards[playerIdx] = newBoard[:]
			ready[playerIdx] = false
		default:
			return state, errors.New("unknown action during placement; use 'ready' or 'randomize'")
		}

	case "combat":
		if turn != playerIdx {
			return state, errors.New("not your turn")
		}

		x := toInt(action["x"])
		y := toInt(action["y"])
		if x < 0 || x >= 10 || y < 0 || y >= 10 {
			return state, errors.New("coordinates out of range (0-9)")
		}

		idx := y*10 + x
		opponentIdx := 1 - playerIdx

		if shots[playerIdx][idx] != 0 {
			return state, errors.New("cell already shot")
		}

		if boards[opponentIdx][idx] == 1 {
			// Hit
			boards[opponentIdx][idx] = 2
			shots[playerIdx][idx] = 2
			shipsLeft[opponentIdx]--
			// Hit: same player shoots again (turn unchanged)
			if shipsLeft[opponentIdx] == 0 {
				winner = playerIdx
			}
		} else {
			// Miss
			boards[opponentIdx][idx] = 3
			shots[playerIdx][idx] = 1
			// Miss: switch turn
			turn = 1 - turn
		}

	default:
		return state, errors.New("unknown phase")
	}

	return State{
		"phase":      phase,
		"boards":     []any{boards[0], boards[1]},
		"shots":      []any{shots[0], shots[1]},
		"ready":      ready,
		"ships_left": shipsLeft,
		"turn":       turn,
		"winner":     winner,
		"ship_sizes": shipSizes,
	}, nil
}

func (g *Game) IsOver(state State) bool {
	return toInt(state["winner"]) != -2
}

func (g *Game) Winner(state State) int {
	return toInt(state["winner"])
}

// StripBoards removes the opponent's board data so clients can't cheat.
// Each player should only see their own board and the shots grid.
func StripBoards(state State, playerIdx int) State {
	out := make(State, len(state))
	for k, v := range state {
		out[k] = v
	}

	boards := cloneBoards(state["boards"])
	opponentIdx := 1 - playerIdx

	// Replace opponent's board: only reveal cells that were already hit or missed
	opponentBoard := make([]int, 100)
	for i, v := range boards[opponentIdx] {
		if v == 2 || v == 3 {
			opponentBoard[i] = v
		}
		// else leave as 0 (unknown)
	}

	if playerIdx == 0 {
		out["boards"] = []any{boards[0], opponentBoard}
	} else {
		out["boards"] = []any{opponentBoard, boards[1]}
	}
	return out
}

func cloneBoards(v any) [2][]int {
	out := [2][]int{make([]int, 100), make([]int, 100)}
	switch s := v.(type) {
	case []any:
		for i, board := range s {
			if i >= 2 {
				break
			}
			switch b := board.(type) {
			case []int:
				copy(out[i], b)
			case []any:
				for j, cell := range b {
					if j < 100 {
						out[i][j] = toInt(cell)
					}
				}
			}
		}
	}
	return out
}

func cloneShots(v any) [2][]int {
	return cloneBoards(v)
}

func cloneReady(v any) []bool {
	out := []bool{false, false}
	switch s := v.(type) {
	case []bool:
		if len(s) >= 2 {
			out[0] = s[0]
			out[1] = s[1]
		}
	case []any:
		for i, x := range s {
			if i >= 2 {
				break
			}
			if b, ok := x.(bool); ok {
				out[i] = b
			}
		}
	}
	return out
}

func cloneInts2(v any) []int {
	out := []int{totalShipCells, totalShipCells}
	switch s := v.(type) {
	case []int:
		if len(s) >= 2 {
			out[0] = s[0]
			out[1] = s[1]
		}
	case []any:
		for i, x := range s {
			if i >= 2 {
				break
			}
			out[i] = toInt(x)
		}
	}
	return out
}

func toInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case float32:
		return int(x)
	}
	return 0
}
