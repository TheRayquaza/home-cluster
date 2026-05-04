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
	return g.InitWithConfig(seed, nil)
}

// InitWithConfig implements ConfigurableGame. Accepts "word_length" (int, 3–8, default 5).
func (g *Game) InitWithConfig(seed int64, config map[string]any) State {
	wordLen := 5
	if config != nil {
		if wl, ok := config["word_length"]; ok {
			if n := toInt(wl); n >= 3 && n <= 8 {
				wordLen = n
			}
		}
	}

	// Filter words to the desired length
	var filtered []string
	for _, w := range Words {
		if len(w) == wordLen {
			filtered = append(filtered, w)
		}
	}
	if len(filtered) == 0 {
		filtered = Words // fallback
	}

	rng := rand.New(rand.NewSource(seed))
	word := filtered[rng.Intn(len(filtered))]

	return State{
		// secret is stored server-side in state but MUST be stripped before sending to clients
		"secret":      word,
		"word_length": wordLen,
		"guesses":     []any{[]any{}, []any{}},
		"hints":       []any{[]any{}, []any{}},
		"solved":      []bool{false, false},
		"attempts":    []int{0, 0},
		"game_over":   false,
		"winner":      -2,
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

	secret, _ := state["secret"].(string)
	wordLen := len(secret)
	if len(guess) != wordLen {
		return state, fmt.Errorf("guess must be exactly %d letters, got %d", wordLen, len(guess))
	}

	solved := cloneBools(state["solved"])
	attempts := cloneInts(state["attempts"])

	if solved[playerIdx] {
		return state, errors.New("you have already solved the puzzle")
	}
	if attempts[playerIdx] >= maxGuesses {
		return state, errors.New("no guesses remaining")
	}

	hint := computeHint(secret, guess)

	guesses := cloneGuessMatrix(state["guesses"])
	hints := cloneHintMatrix(state["hints"])

	guesses[playerIdx] = append(guesses[playerIdx], guess)
	hints[playerIdx] = append(hints[playerIdx], hint)
	attempts[playerIdx]++

	if guess == secret {
		solved[playerIdx] = true
	}

	wordLength := toInt(state["word_length"])
	if wordLength == 0 {
		wordLength = wordLen
	}

	newState := State{
		"secret":      secret,
		"word_length": wordLength,
		"guesses":     guesses,
		"hints":       hints,
		"solved":      solved,
		"attempts":    attempts,
		"game_over":   false,
		"winner":      -2,
	}

	// Both players must be done before game ends
	p0done := solved[0] || attempts[0] >= maxGuesses
	p1done := solved[1] || attempts[1] >= maxGuesses
	if p0done && p1done {
		winner := computeWinner(solved, attempts, hints)
		newState["winner"] = winner
		newState["game_over"] = true
		newState["revealed_word"] = secret
	}

	return newState, nil
}

// computeHint returns per-letter hints: 2=green, 1=yellow, 0=gray.
func computeHint(secret, guess string) []int {
	n := len(secret)
	hint := make([]int, n)
	used := make([]bool, n)

	// First pass: find greens
	for i := 0; i < n; i++ {
		if guess[i] == secret[i] {
			hint[i] = 2
			used[i] = true
		}
	}

	// Second pass: find yellows
	for i := 0; i < n; i++ {
		if hint[i] == 2 {
			continue
		}
		for j := 0; j < n; j++ {
			if !used[j] && guess[i] == secret[j] {
				hint[i] = 1
				used[j] = true
				break
			}
		}
	}

	return hint
}

// countGreens counts total green hints (value==2) across all rows in a hint matrix player slice.
func countGreens(playerHints [][]int) int {
	total := 0
	for _, row := range playerHints {
		for _, v := range row {
			if v == 2 {
				total++
			}
		}
	}
	return total
}

// computeWinner returns -1=draw, 0=p0 wins, 1=p1 wins.
// Called only when both players are done.
func computeWinner(solved []bool, attempts []int, hints [][][]int) int {
	if solved[0] && solved[1] {
		// Both solved: fewer attempts wins; tie → draw
		if attempts[0] < attempts[1] {
			return 0
		} else if attempts[1] < attempts[0] {
			return 1
		}
		return -1
	}
	if solved[0] {
		return 0
	}
	if solved[1] {
		return 1
	}
	// Neither solved: most greens wins; tie → draw
	g0 := countGreens(hints[0])
	g1 := countGreens(hints[1])
	if g0 > g1 {
		return 0
	} else if g1 > g0 {
		return 1
	}
	return -1
}

func (g *Game) IsOver(state State) bool {
	return toInt(state["winner"]) != -2
}

func (g *Game) Winner(state State) int {
	return toInt(state["winner"])
}

// StripSecret returns a copy of the state without the secret word (safe to send to clients).
// revealed_word is also stripped unless game_over is true.
func StripSecret(state State) State {
	out := make(State, len(state))
	for k, v := range state {
		out[k] = v
	}
	delete(out, "secret")
	if !toBool(out["game_over"]) {
		delete(out, "revealed_word")
	}
	return out
}

// toBool converts an interface value to bool.
func toBool(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
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
