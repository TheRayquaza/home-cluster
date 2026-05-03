package pong

import (
	"math"
	"math/rand"
)

type State = map[string]any
type Action = map[string]any

const (
	width       = 800.0
	height      = 400.0
	paddleH     = 80.0
	paddleW     = 12.0
	ballSize    = 10.0
	paddleSpeed = 300.0
	winScore    = 7
)

// Game implements Pong (real-time).
type Game struct{}

func (g *Game) ID() string     { return "pong" }
func (g *Game) Name() string   { return "Pong" }
func (g *Game) RealTime() bool { return true }

func (g *Game) Init(seed int64) State {
	rng := rand.New(rand.NewSource(seed))

	// Random initial ball direction
	angle := (rng.Float64()*60 - 30) * math.Pi / 180
	speed := 200.0
	vx := math.Cos(angle) * speed
	vy := math.Sin(angle) * speed
	// Randomly choose left or right direction
	if rng.Intn(2) == 0 {
		vx = -vx
	}

	return State{
		"ball": map[string]any{
			"x":  width / 2,
			"y":  height / 2,
			"vx": vx,
			"vy": vy,
		},
		"paddles":  []float64{height/2 - paddleH/2, height/2 - paddleH/2},
		"scores":   []int{0, 0},
		"width":    width,
		"height":   height,
		"paddle_h": paddleH,
		"paddle_w": paddleW,
		"winner":   -2,
		"bounces":  0,
	}
}

func (g *Game) Apply(state State, _ int, _ Action) (State, error) {
	// Pong inputs are handled via Tick; Apply is a no-op for real-time games.
	return state, nil
}

func (g *Game) IsOver(state State) bool {
	return toInt(state["winner"]) != -2
}

func (g *Game) Winner(state State) int {
	return toInt(state["winner"])
}

func (g *Game) Tick(state State, inputs [2]*Action, dt float64) State {
	if g.IsOver(state) {
		return state
	}

	paddles := cloneFloat64s(state["paddles"])
	scores := cloneInts(state["scores"])
	ball := cloneBallMap(state["ball"])
	bounces := toInt(state["bounces"])

	bx := toFloat(ball["x"])
	by := toFloat(ball["y"])
	bvx := toFloat(ball["vx"])
	bvy := toFloat(ball["vy"])

	// Move paddles based on inputs
	for i, inp := range inputs {
		if inp == nil {
			continue
		}
		dir, _ := (*inp)["dir"].(string)
		switch dir {
		case "up":
			paddles[i] -= paddleSpeed * dt
		case "down":
			paddles[i] += paddleSpeed * dt
		}
		// Clamp paddle position
		if paddles[i] < 0 {
			paddles[i] = 0
		}
		if paddles[i]+paddleH > height {
			paddles[i] = height - paddleH
		}
	}

	// Move ball
	bx += bvx * dt
	by += bvy * dt

	// Bounce off top/bottom walls
	if by <= 0 {
		by = 0
		bvy = math.Abs(bvy)
	}
	if by+ballSize >= height {
		by = height - ballSize
		bvy = -math.Abs(bvy)
	}

	scored := -1

	// Left paddle collision (player 0, x=0..paddleW)
	if bx <= paddleW && bvx < 0 {
		padY := paddles[0]
		if by+ballSize >= padY && by <= padY+paddleH {
			bx = paddleW
			bvx = math.Abs(bvx)
			// Add slight angle based on where ball hits paddle
			relY := (by + ballSize/2) - (padY + paddleH/2)
			bvy = relY * 5
			bounces++
			boost := 1.0 + float64(bounces)*0.025
			currentSpeed := math.Sqrt(bvx*bvx + bvy*bvy)
			if currentSpeed < 500 {
				bvx *= boost
				bvy *= boost
			}
		} else if bx < 0 {
			// Scored: player 1 wins this point
			scored = 1
		}
	}

	// Right paddle collision (player 1, x=width-paddleW..width)
	if bx+ballSize >= width-paddleW && bvx > 0 {
		padY := paddles[1]
		if by+ballSize >= padY && by <= padY+paddleH {
			bx = width - paddleW - ballSize
			bvx = -math.Abs(bvx)
			relY := (by + ballSize/2) - (padY + paddleH/2)
			bvy = relY * 5
			bounces++
			boost := 1.0 + float64(bounces)*0.025
			currentSpeed := math.Sqrt(bvx*bvx + bvy*bvy)
			if currentSpeed < 500 {
				bvx *= boost
				bvy *= boost
			}
		} else if bx+ballSize > width {
			// Scored: player 0 wins this point
			scored = 0
		}
	}

	if scored >= 0 {
		scores[scored]++
		// Reset ball to center
		rng := rand.New(rand.NewSource(int64(scores[0]*100 + scores[1])))
		angle := (rng.Float64()*60 - 30) * math.Pi / 180
		speed := 200.0
		bx = width / 2
		by = height / 2
		bvx = math.Cos(angle) * speed
		bvy = math.Sin(angle) * speed
		if scored == 0 {
			bvx = -bvx // serve toward scoring side
		}
		bounces = 0
	}

	ball["x"] = bx
	ball["y"] = by
	ball["vx"] = bvx
	ball["vy"] = bvy

	newState := State{
		"ball":     ball,
		"paddles":  paddles,
		"scores":   scores,
		"width":    width,
		"height":   height,
		"paddle_h": paddleH,
		"paddle_w": paddleW,
		"winner":   -2,
		"bounces":  bounces,
	}

	// Check win condition
	for i, s := range scores {
		if s >= winScore {
			newState["winner"] = i
			break
		}
	}

	return newState
}

func cloneFloat64s(v any) []float64 {
	switch s := v.(type) {
	case []float64:
		out := make([]float64, len(s))
		copy(out, s)
		return out
	case []any:
		out := make([]float64, len(s))
		for i, x := range s {
			out[i] = toFloat(x)
		}
		return out
	}
	return []float64{0, 0}
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

func cloneBallMap(v any) map[string]any {
	out := map[string]any{}
	if m, ok := v.(map[string]any); ok {
		for k, val := range m {
			out[k] = val
		}
	}
	return out
}

func toFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	}
	return 0
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
