package sim

const (
	DefaultMaxLoad = 16
	// DefaultTickRate is deliberately below the 60 Hz target in docs/overview.md
	// while rendering is tick-granular. At 60 Hz with Speed=1, Subjects cross
	// the grid in ~80 ms and teleport visually. Raise this back to 60 once
	// render-side interpolation lands.
	DefaultTickRate = 10
)

type GameState struct {
	Grid        *Grid
	USD         float64
	Research    map[Element]int
	MaxLoad     int
	CurrentLoad int
	TickRate    int
	Ticks       uint64
}

func NewGameState() *GameState {
	return &GameState{
		Grid:     NewGrid(),
		Research: map[Element]int{},
		MaxLoad:  DefaultMaxLoad,
		TickRate: DefaultTickRate,
	}
}

// HardReset wipes in-memory state back to defaults. The caller is responsible
// for persisting the reset (or deleting the save) so a subsequent Load does
// not restore the previous state.
func (s *GameState) HardReset() {
	*s = *NewGameState()
}
