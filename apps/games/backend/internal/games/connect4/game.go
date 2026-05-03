package connect4

import (
	"errors"
	"fmt"
)

type State = map[string]any
type Action = map[string]any

const (
	cols = 7
	rows = 6
)

// Game implements Connect Four (7 cols x 6 rows).
type Game struct{}

func (g *Game) ID() string     { return "connect4" }
func (g *Game) Name() string   { return "Connect Four" }
func (g *Game) RealTime() bool { return false }

func (g *Game) Init(_ int64) State {
	// board[col][row] — row 0 is the bottom
	board := make([][]int, cols)
	for c := range board {
		board[c] = make([]int, rows)
	}
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

	colVal, ok := action["col"]
	if !ok {
		return state, errors.New("action missing 'col'")
	}
	col := toInt(colVal)
	if col < 0 || col >= cols {
		return state, errors.New("col out of range 0-6")
	}

	board := cloneBoard(state["board"])

	// Find lowest empty row in this column
	row := -1
	for r := 0; r < rows; r++ {
		if board[col][r] == 0 {
			row = r
			break
		}
	}
	if row == -1 {
		return state, errors.New("column is full")
	}

	board[col][row] = playerIdx + 1 // 1=p1, 2=p2

	newState := State{
		"board": board,
		"turn":  1 - turn,
	}

	winner := checkWinner(board, col, row)
	newState["winner"] = winner
	return newState, nil
}

func (g *Game) IsOver(state State) bool {
	return toInt(state["winner"]) != -2
}

func (g *Game) Winner(state State) int {
	return toInt(state["winner"])
}

// checkWinner checks for a win from the last placed piece at (col, row).
func checkWinner(board [][]int, col, row int) int {
	piece := board[col][row]

	dirs := [][2]int{
		{1, 0}, {0, 1}, {1, 1}, {1, -1},
	}
	for _, d := range dirs {
		count := 1
		// Count in positive direction
		for i := 1; i < 4; i++ {
			c2, r2 := col+d[0]*i, row+d[1]*i
			if c2 < 0 || c2 >= cols || r2 < 0 || r2 >= rows || board[c2][r2] != piece {
				break
			}
			count++
		}
		// Count in negative direction
		for i := 1; i < 4; i++ {
			c2, r2 := col-d[0]*i, row-d[1]*i
			if c2 < 0 || c2 >= cols || r2 < 0 || r2 >= rows || board[c2][r2] != piece {
				break
			}
			count++
		}
		if count >= 4 {
			return piece - 1 // 0 or 1
		}
	}

	// Check draw: all columns full
	for c := 0; c < cols; c++ {
		if board[c][rows-1] == 0 {
			return -2 // still ongoing
		}
	}
	return -1 // draw
}

func cloneBoard(v any) [][]int {
	board := make([][]int, cols)
	for c := range board {
		board[c] = make([]int, rows)
	}
	switch b := v.(type) {
	case [][]int:
		for c := range b {
			copy(board[c], b[c])
		}
	case []any:
		for c, col := range b {
			switch colv := col.(type) {
			case []int:
				copy(board[c], colv)
			case []any:
				for r, cell := range colv {
					board[c][r] = toInt(cell)
				}
			}
		}
	}
	return board
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
