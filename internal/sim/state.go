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
	// Owned is the total number of each component kind the player has ever
	// purchased. Monotonic: incremented by PurchaseComponent; never
	// decreased by Erase (removed components return to the available pool,
	// not the shop). Available = Owned - count-placed-on-grid.
	Owned       map[ComponentKind]int `json:"owned,omitempty"`
	MaxLoad     int
	CurrentLoad int
	TickRate    int
	Ticks       uint64
}

func NewGameState() *GameState {
	return &GameState{
		Layer:            LayerGenesis,
		Grid:             NewGrid(),
		Research:         map[Element]int{},
		UnlockedElements: map[Element]bool{ElementHydrogen: true},
		Owned:            starterInventory(),
		MaxLoad:          DefaultMaxLoad,
		TickRate:         DefaultTickRate,
	}
}

// starterInventory is the set of components a brand-new game begins with so
// the player can build the first loop without spending $USD. Tuning these
// numbers is a design lever — see docs/features/component-cost.md.
func starterInventory() map[ComponentKind]int {
	return map[ComponentKind]int{
		KindInjector:    1,
		KindAccelerator: 2,
		KindRotator:     1,
		KindCollector:   1,
	}
}

// HardReset wipes in-memory state back to defaults. The caller is responsible
// for persisting the reset (or deleting the save) so a subsequent Load does
// not restore the previous state.
func (s *GameState) HardReset() {
	*s = *NewGameState()
}
