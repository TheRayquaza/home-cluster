package lightcycles

type State = map[string]any
type Action = map[string]any

const (
	gridWidth  = 100
	gridHeight = 60
)

// Game implements Light Cycles (real-time, 10fps tick).
type Game struct{}

func (g *Game) ID() string     { return "lightcycles" }
func (g *Game) Name() string   { return "Light Cycles" }
func (g *Game) RealTime() bool { return true }

func (g *Game) Init(_ int64) State {
	return State{
		"width":  gridWidth,
		"height": gridHeight,
		"players": []any{
			map[string]any{"x": 15, "y": 30, "dir": "right"},
			map[string]any{"x": 85, "y": 30, "dir": "left"},
		},
		"trails": []any{
			[]any{[]any{15, 30}},
			[]any{[]any{85, 30}},
		},
		"alive":     []bool{true, true},
		"winner":    -2,
		"phase":     "countdown",
		"countdown": 3,
		"tickCount": 0,
	}
}

func (g *Game) Apply(state State, _ int, _ Action) (State, error) {
	// Light Cycles inputs are handled via Tick; Apply is a no-op.
	return state, nil
}

func (g *Game) IsOver(state State) bool {
	return toInt(state["winner"]) != -2
}

func (g *Game) Winner(state State) int {
	return toInt(state["winner"])
}

func (g *Game) Tick(state State, inputs [2]*Action, _ float64) State {
	if g.IsOver(state) {
		return state
	}

	players := clonePlayers(state["players"])
	trails := cloneTrails(state["trails"])
	alive := cloneAlive(state["alive"])
	phase, _ := state["phase"].(string)
	tickCount := toInt(state["tickCount"]) + 1
	countdown := toInt(state["countdown"])

	if phase == "countdown" {
		// Every 10 ticks decrement countdown (10fps → 1s per 10 ticks)
		if tickCount%10 == 0 {
			countdown--
		}
		if countdown <= 0 {
			phase = "playing"
		}
		return State{
			"width":     gridWidth,
			"height":    gridHeight,
			"players":   players,
			"trails":    trails,
			"alive":     alive,
			"winner":    -2,
			"phase":     phase,
			"countdown": countdown,
			"tickCount": tickCount,
		}
	}

	// Apply direction inputs (prevent reversals)
	for i, inp := range inputs {
		if inp == nil || !alive[i] {
			continue
		}
		newDir, _ := (*inp)["dir"].(string)
		if newDir == "" {
			continue
		}
		currentDir, _ := players[i]["dir"].(string)
		if !isReversal(currentDir, newDir) {
			players[i]["dir"] = newDir
		}
	}

	// Advance each alive player
	newPositions := make([][2]int, 2)
	for i := range players {
		if !alive[i] {
			continue
		}
		x := toInt(players[i]["x"])
		y := toInt(players[i]["y"])
		dir, _ := players[i]["dir"].(string)
		nx, ny := x, y
		switch dir {
		case "up":
			ny--
		case "down":
			ny++
		case "left":
			nx--
		case "right":
			nx++
		}
		newPositions[i] = [2]int{nx, ny}
	}

	// Build set of all trail cells (before movement)
	trailSet := map[[2]int]bool{}
	for _, trail := range trails {
		for _, cell := range trail {
			c := toPoint(cell)
			trailSet[c] = true
		}
	}

	// Check collisions for each alive player
	for i := range players {
		if !alive[i] {
			continue
		}
		nx, ny := newPositions[i][0], newPositions[i][1]
		pos := [2]int{nx, ny}

		// Wall collision
		if nx < 0 || nx >= gridWidth || ny < 0 || ny >= gridHeight {
			alive[i] = false
			continue
		}
		// Trail collision (any trail, including own)
		if trailSet[pos] {
			alive[i] = false
			continue
		}
	}

	// Check head-on collision (both move to same cell)
	if alive[0] && alive[1] && newPositions[0] == newPositions[1] {
		alive[0] = false
		alive[1] = false
	}

	// Update positions and trails for surviving players
	for i := range players {
		if !alive[i] {
			continue
		}
		nx, ny := newPositions[i][0], newPositions[i][1]
		players[i]["x"] = nx
		players[i]["y"] = ny
		trails[i] = append(trails[i], []any{nx, ny})
	}

	// Determine winner
	winner := -2
	if !alive[0] && !alive[1] {
		winner = -1 // draw
	} else if !alive[0] {
		winner = 1
	} else if !alive[1] {
		winner = 0
	}

	return State{
		"width":     gridWidth,
		"height":    gridHeight,
		"players":   players,
		"trails":    trails,
		"alive":     alive,
		"winner":    winner,
		"phase":     "playing",
		"countdown": 0,
		"tickCount": tickCount,
	}
}

func isReversal(current, next string) bool {
	return (current == "up" && next == "down") ||
		(current == "down" && next == "up") ||
		(current == "left" && next == "right") ||
		(current == "right" && next == "left")
}

func clonePlayers(v any) []map[string]any {
	out := make([]map[string]any, 2)
	for i := range out {
		out[i] = map[string]any{}
	}
	switch s := v.(type) {
	case []map[string]any:
		for i, p := range s {
			for k, val := range p {
				out[i][k] = val
			}
		}
	case []any:
		for i, p := range s {
			if m, ok := p.(map[string]any); ok {
				for k, val := range m {
					out[i][k] = val
				}
			}
		}
	}
	return out
}

func cloneTrails(v any) [][]any {
	out := [][]any{{}, {}}
	switch s := v.(type) {
	case [][]any:
		for i, trail := range s {
			out[i] = make([]any, len(trail))
			copy(out[i], trail)
		}
	case []any:
		for i, trail := range s {
			switch t := trail.(type) {
			case []any:
				out[i] = make([]any, len(t))
				copy(out[i], t)
			}
		}
	}
	return out
}

func cloneAlive(v any) []bool {
	out := []bool{true, true}
	switch s := v.(type) {
	case []bool:
		out[0] = s[0]
		out[1] = s[1]
	case []any:
		for i, x := range s {
			if b, ok := x.(bool); ok {
				out[i] = b
			}
		}
	}
	return out
}

func toPoint(v any) [2]int {
	switch s := v.(type) {
	case []any:
		if len(s) >= 2 {
			return [2]int{toInt(s[0]), toInt(s[1])}
		}
	case []int:
		if len(s) >= 2 {
			return [2]int{s[0], s[1]}
		}
	}
	return [2]int{0, 0}
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
