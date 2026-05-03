package games

// State is a serializable game state sent to clients.
type State = map[string]any

// Action is a player action received from a client.
type Action = map[string]any

// Game describes a two-player game that can be turn-based or real-time.
type Game interface {
	// ID returns the stable identifier used in the registry and database.
	ID() string
	// Name returns the human-readable display name.
	Name() string
	// RealTime returns true for games that use the Tick loop (pong, lightcycles).
	RealTime() bool
	// Init creates the initial state. seed is used for deterministic randomness.
	Init(seed int64) State
	// Apply validates and applies a player action to the state, returning the next state.
	Apply(state State, playerIdx int, action Action) (State, error)
	// IsOver returns true when the game has ended.
	IsOver(state State) bool
	// Winner returns the result: 0=p1 wins, 1=p2 wins, -1=draw, -2=ongoing.
	Winner(state State) int
}

// TickableGame extends Game for real-time games that advance via a server-side tick.
type TickableGame interface {
	Game
	// Tick advances the game state by dt seconds given both players' current inputs.
	Tick(state State, inputs [2]*Action, dt float64) State
}

// ConfigurableGame extends Game for games that accept pre-game config (e.g. grid size).
type ConfigurableGame interface {
	Game
	InitWithConfig(seed int64, config map[string]any) State
}
