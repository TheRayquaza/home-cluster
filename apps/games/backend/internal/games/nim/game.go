package nim

import (
	"errors"
	"fmt"
)

type State = map[string]any
type Action = map[string]any

const startingSticks = 21

// Game implements Nim (misère: last to take loses).
type Game struct{}

func (g *Game) ID() string     { return "nim" }
func (g *Game) Name() string   { return "Nim" }
func (g *Game) RealTime() bool { return false }

func (g *Game) Init(_ int64) State {
	return State{
		"sticks": startingSticks,
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

	takeVal, ok := action["take"]
	if !ok {
		return state, errors.New("action missing 'take'")
	}
	take := toInt(takeVal)

	sticks := toInt(state["sticks"])
	maxTake := 3
	if sticks < maxTake {
		maxTake = sticks
	}
	if take < 1 || take > maxTake {
		return state, fmt.Errorf("take must be between 1 and %d", maxTake)
	}

	sticks -= take
	newState := State{
		"sticks": sticks,
		"turn":   1 - turn,
	}

	if sticks == 0 {
		// Misère: the player who took the last stick loses.
		// That player is the current `turn` (playerIdx).
		// So the winner is the opponent.
		newState["winner"] = 1 - turn
	} else {
		newState["winner"] = -2
	}

	return newState, nil
}

func (g *Game) IsOver(state State) bool {
	return toInt(state["winner"]) != -2
}

func (g *Game) Winner(state State) int {
	return toInt(state["winner"])
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
