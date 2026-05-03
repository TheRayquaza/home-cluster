package wordle

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
)

type State = map[string]any
type Action = map[string]any

const maxGuesses = 6

// Game implements Wordle (both players guess the same secret word).
type Game struct{}

func (g *Game) ID() string     { return "wordle" }
func (g *Game) Name() string   { return "Wordle" }
func (g *Game) RealTime() bool { return false }

func (g *Game) Init(seed int64) State {
	rng := rand.New(rand.NewSource(seed))
	word := Words[rng.Intn(len(Words))]

	return State{
		// secret is stored server-side in state but MUST be stripped before sending to clients
		"secret":    word,
		"guesses":   []any{[]any{}, []any{}},
		"hints":     []any{[]any{}, []any{}},
		"solved":    []bool{false, false},
		"attempts":  []int{0, 0},
		"game_over": false,
		"winner":    -2,
	}
}

func (g *Game) Apply(state State, playerIdx int, action Action) (State, error) {
	if g.IsOver(state) {
		return state, errors.New("game is already over")
	}

	guessVal, ok := action["guess"]
	if !ok {
		return state, errors.New("action missing 'guess'")
	}
	guess, ok := guessVal.(string)
	if !ok {
		return state, errors.New("'guess' must be a string")
	}
	guess = strings.ToUpper(strings.TrimSpace(guess))
	if len(guess) != 5 {
		return state, fmt.Errorf("guess must be exactly 5 letters, got %d", len(guess))
	}

	solved := cloneBools(state["solved"])
	if solved[playerIdx] {
		return state, errors.New("you have already solved the puzzle")
	}

	attempts := cloneInts(state["attempts"])
	if attempts[playerIdx] >= maxGuesses {
		return state, errors.New("no guesses remaining")
	}

	secret, _ := state["secret"].(string)
	hint := computeHint(secret, guess)

	guesses := cloneGuessMatrix(state["guesses"])
	hints := cloneHintMatrix(state["hints"])

	guesses[playerIdx] = append(guesses[playerIdx], guess)
	hints[playerIdx] = append(hints[playerIdx], hint)
	attempts[playerIdx]++

	if guess == secret {
		solved[playerIdx] = true
	}

	newState := State{
		"secret":    secret,
		"guesses":   guesses,
		"hints":     hints,
		"solved":    solved,
		"attempts":  attempts,
		"game_over": false,
		"winner":    -2,
	}

	// Check if game is over
	winner := computeWinner(solved, attempts)
	if winner != -2 {
		newState["winner"] = winner
		newState["game_over"] = true
	}

	return newState, nil
}

// computeHint returns per-letter hints: 2=green, 1=yellow, 0=gray.
func computeHint(secret, guess string) []int {
	hint := make([]int, 5)
	used := make([]bool, 5)

	// First pass: find greens
	for i := 0; i < 5; i++ {
		if guess[i] == secret[i] {
			hint[i] = 2
			used[i] = true
		}
	}

	// Second pass: find yellows
	for i := 0; i < 5; i++ {
		if hint[i] == 2 {
			continue
		}
		for j := 0; j < 5; j++ {
			if !used[j] && guess[i] == secret[j] {
				hint[i] = 1
				used[j] = true
				break
			}
		}
	}

	return hint
}

// computeWinner returns -2=ongoing, -1=draw, 0=p1, 1=p2.
func computeWinner(solved []bool, attempts []int) int {
	p0done := solved[0] || attempts[0] >= maxGuesses
	p1done := solved[1] || attempts[1] >= maxGuesses

	if !p0done && !p1done {
		return -2
	}

	// If one solved, allow other to finish remaining guesses
	if solved[0] && !p1done {
		return -2 // wait for p2 to finish
	}
	if solved[1] && !p0done {
		return -2 // wait for p1 to finish
	}

	// Both done: determine winner
	if solved[0] && solved[1] {
		if attempts[0] < attempts[1] {
			return 0
		} else if attempts[1] < attempts[0] {
			return 1
		}
		return -1 // draw
	}
	if solved[0] {
		return 0
	}
	if solved[1] {
		return 1
	}
	return -1 // neither solved: draw
}

func (g *Game) IsOver(state State) bool {
	return toInt(state["winner"]) != -2
}

func (g *Game) Winner(state State) int {
	return toInt(state["winner"])
}

// StripSecret returns a copy of the state without the secret word (safe to send to clients).
func StripSecret(state State) State {
	out := make(State, len(state))
	for k, v := range state {
		out[k] = v
	}
	delete(out, "secret")
	return out
}

func cloneBools(v any) []bool {
	switch b := v.(type) {
	case []bool:
		out := make([]bool, len(b))
		copy(out, b)
		return out
	case []any:
		out := make([]bool, len(b))
		for i, x := range b {
			if bv, ok := x.(bool); ok {
				out[i] = bv
			}
		}
		return out
	}
	return []bool{false, false}
}

func cloneInts(v any) []int {
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
	return []int{0, 0}
}

func cloneGuessMatrix(v any) [][]string {
	out := [][]string{{}, {}}
	switch s := v.(type) {
	case [][]string:
		for i, row := range s {
			out[i] = make([]string, len(row))
			copy(out[i], row)
		}
	case []any:
		for i, row := range s {
			switch r := row.(type) {
			case []string:
				out[i] = make([]string, len(r))
				copy(out[i], r)
			case []any:
				for _, gg := range r {
					if gs, ok := gg.(string); ok {
						out[i] = append(out[i], gs)
					}
				}
			}
		}
	}
	return out
}

func cloneHintMatrix(v any) [][][]int {
	out := [][][]int{{}, {}}
	switch s := v.(type) {
	case [][][]int:
		for i, row := range s {
			out[i] = make([][]int, len(row))
			for j, h := range row {
				out[i][j] = make([]int, len(h))
				copy(out[i][j], h)
			}
		}
	case []any:
		for i, row := range s {
			switch r := row.(type) {
			case [][]int:
				out[i] = make([][]int, len(r))
				for j, h := range r {
					out[i][j] = make([]int, len(h))
					copy(out[i][j], h)
				}
			case []any:
				for _, hint := range r {
					switch h := hint.(type) {
					case []int:
						out[i] = append(out[i], h)
					case []any:
						hintsRow := make([]int, len(h))
						for k, x := range h {
							hintsRow[k] = toInt(x)
						}
						out[i] = append(out[i], hintsRow)
					}
				}
			}
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
