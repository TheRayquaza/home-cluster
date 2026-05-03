package memory

import (
	"errors"
	"math/rand"
)

type State = map[string]any
type Action = map[string]any

// Game implements Memory Match with configurable grid size.
type Game struct{}

func (g *Game) ID() string     { return "memory" }
func (g *Game) Name() string   { return "Memory Match" }
func (g *Game) RealTime() bool { return false }

func (g *Game) Init(seed int64) State {
	return g.InitWithConfig(seed, nil)
}

// InitWithConfig supports "pairs" config key (default 8, min 4, max 18).
func (g *Game) InitWithConfig(seed int64, config map[string]any) State {
	pairs := 8
	if config != nil {
		if v, ok := config["pairs"]; ok {
			if n := toInt(v); n >= 4 && n <= 18 {
				pairs = n
			}
		}
	}

	total := pairs * 2
	deck := make([]int, total)
	for i := 0; i < pairs; i++ {
		deck[i*2] = i
		deck[i*2+1] = i
	}
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	cols, rows := gridDims(pairs)

	return State{
		"cards":    deck,
		"revealed": make([]bool, total),
		"matched":  make([]bool, total),
		"scores":   []int{0, 0},
		"turn":     0,
		"pending":  []int{},
		"waiting":  false,
		"winner":   -2,
		"pairs":    pairs,
		"cols":     cols,
		"rows":     rows,
	}
}

func gridDims(pairs int) (cols, rows int) {
	switch pairs {
	case 4:
		return 4, 2
	case 6:
		return 4, 3
	case 8:
		return 4, 4
	case 10:
		return 5, 4
	case 12:
		return 6, 4
	case 15:
		return 5, 6
	case 18:
		return 6, 6
	default:
		for c := 6; c >= 2; c-- {
			if (pairs*2)%c == 0 {
				return c, (pairs * 2) / c
			}
		}
		return 4, pairs / 2
	}
}

func (g *Game) Apply(state State, playerIdx int, action Action) (State, error) {
	if g.IsOver(state) {
		return state, errors.New("game is already over")
	}

	turn := toInt(state["turn"])
	if turn != playerIdx {
		return state, errors.New("not your turn")
	}

	// Resolve previous non-match before processing new flip
	if toBool(state["waiting"]) {
		state = resolveWaiting(state)
	}

	cardVal, ok := action["card"]
	if !ok {
		return state, errors.New("action missing 'card'")
	}
	card := toInt(cardVal)

	cards := cloneInts(state["cards"])
	total := len(cards)

	if card < 0 || card >= total {
		return state, errors.New("card index out of range")
	}

	revealed := cloneBools(state["revealed"], total)
	matched := cloneBools(state["matched"], total)
	scores := cloneInts(state["scores"])
	pending := cloneInts(state["pending"])

	if revealed[card] || matched[card] {
		return state, errors.New("card already revealed or matched")
	}

	revealed[card] = true
	pending = append(pending, card)

	newState := State{
		"cards":    cards,
		"revealed": revealed,
		"matched":  matched,
		"scores":   scores,
		"turn":     turn,
		"pending":  pending,
		"waiting":  false,
		"winner":   -2,
		"pairs":    state["pairs"],
		"cols":     state["cols"],
		"rows":     state["rows"],
	}

	if len(pending) == 2 {
		if cards[pending[0]] == cards[pending[1]] {
			matched[pending[0]] = true
			matched[pending[1]] = true
			scores[turn]++
			newState["matched"] = matched
			newState["scores"] = scores
			newState["pending"] = []int{}
		} else {
			newState["waiting"] = true
			newState["turn"] = 1 - turn
		}
	}

	allMatched := true
	for _, m := range matched {
		if !m {
			allMatched = false
			break
		}
	}
	if allMatched {
		winner := -1
		if scores[0] > scores[1] {
			winner = 0
		} else if scores[1] > scores[0] {
			winner = 1
		}
		newState["winner"] = winner
	}

	return newState, nil
}

func resolveWaiting(state State) State {
	cards := cloneInts(state["cards"])
	pending := cloneInts(state["pending"])
	revealed := cloneBools(state["revealed"], len(cards))
	for _, idx := range pending {
		revealed[idx] = false
	}
	out := copyState(state)
	out["revealed"] = revealed
	out["pending"] = []int{}
	out["waiting"] = false
	return out
}

func (g *Game) IsOver(state State) bool {
	w, ok := state["winner"]
	if !ok {
		return false
	}
	return toInt(w) != -2
}

func (g *Game) Winner(state State) int {
	return toInt(state["winner"])
}

func copyState(state State) State {
	out := make(State, len(state))
	for k, v := range state {
		out[k] = v
	}
	return out
}

func cloneBools(v any, size int) []bool {
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
	return make([]bool, size)
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
	return []int{}
}

func toBool(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
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
