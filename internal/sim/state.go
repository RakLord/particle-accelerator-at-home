package sim

import "particleaccelerator/internal/bignum"

const (
	DefaultMaxLoad = 16
	// DefaultTickRate: interpolation is live (see docs/features/smooth-motion.md)
	// so this is no longer a rendering constraint — raising it is a gameplay
	// decision about how fast the grid feels.
	DefaultTickRate = 10
)

type GameState struct {
	Layer            Layer
	Grid             *Grid
	USD              bignum.Decimal
	Research         map[Element]int
	UnlockedElements map[Element]bool
	MaxLoad          int
	CurrentLoad      int
	TickRate         int
	Ticks            uint64
}

func NewGameState() *GameState {
	return &GameState{
		Layer:            LayerGenesis,
		Grid:             NewGrid(),
		Research:         map[Element]int{},
		UnlockedElements: map[Element]bool{ElementHydrogen: true},
		MaxLoad:          DefaultMaxLoad,
		TickRate:         DefaultTickRate,
	}
}

// HardReset wipes in-memory state back to defaults. The caller is responsible
// for persisting the reset (or deleting the save) so a subsequent Load does
// not restore the previous state.
func (s *GameState) HardReset() {
	*s = *NewGameState()
}
