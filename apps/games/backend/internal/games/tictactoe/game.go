package tictactoe

import (
	"errors"
	"fmt"
)

type State = map[string]any
type Action = map[string]any

// Game implements a standard 3x3 Tic-Tac-Toe game.
type Game struct{}

func (g *Game) ID() string     { return "tictactoe" }
func (g *Game) Name() string   { return "Tic Tac Toe" }
func (g *Game) RealTime() bool { return false }

func (g *Game) Init(_ int64) State {
	board := make([]int, 9)
	return State{
		"board":  board,
		"turn":   0,
		"winner": -2,
	}
}

func (g *Game) Apply(state State, playerIdx int, action Action) (State, error) {
	if g.IsOver(state) {
		return state, errors.New("game is already over")
	}

	turn := toInt(state["turn"])
	if turn != playerIdx {
		return state, fmt.Errorf("not your turn: expected player %d", turn)
	}

	cellVal, ok := action["cell"]
	if !ok {
		return state, errors.New("action missing 'cell'")
	}
	cell := toInt(cellVal)
	if cell < 0 || cell > 8 {
		return state, errors.New("cell out of range 0-8")
	}

	board := cloneIntSlice(state["board"])
	if board[cell] != 0 {
		return state, errors.New("cell already occupied")
	}

	board[cell] = playerIdx + 1 // 1=p1, 2=p2

	newState := State{
		"board": board,
		"turn":  1 - turn,
	}

	winner := checkWinner(board)
	newState["winner"] = winner
	return newState, nil
}

func (g *Game) IsOver(state State) bool {
	return toInt(state["winner"]) != -2
}

func (g *Game) Winner(state State) int {
	return toInt(state["winner"])
}

// checkWinner returns 0=p1, 1=p2, -1=draw, -2=ongoing.
func checkWinner(board []int) int {
	lines := [][3]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // cols
		{0, 4, 8}, {2, 4, 6},             // diags
	}
	for _, l := range lines {
		if board[l[0]] != 0 && board[l[0]] == board[l[1]] && board[l[1]] == board[l[2]] {
			return board[l[0]] - 1 // 0 or 1
		}
	}
	for _, v := range board {
		if v == 0 {
			return -2 // still ongoing
		}
	}
	return -1 // draw: all filled, no winner
}

func cloneIntSlice(v any) []int {
	switch s := v.(type) {
	case []int:
		out := make([]int, len(s))
		copy(out, s)
		return out
	case []any:
		out := make([]int, len(s))
		for i, x := range s {
			out[i] = toInt(x)
		}
		return out
	}
	return make([]int, 9)
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
