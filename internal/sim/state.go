package sim

import "particleaccelerator/internal/bignum"

const (
	DefaultMaxLoad = 16
	// DefaultInjectionCooldownSeconds is the base manual injection cooldown.
	// Injector-rate upgrades shorten the effective tick count derived from this.
	DefaultInjectionCooldownSeconds = 5
	// MaxCollectionLogEntries is the number of recent collected Subjects kept
	// for the in-game collection log.
	MaxCollectionLogEntries = 10
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
	BestStats        map[Element]ElementBestStats
	CollectionLog    []CollectionLogEntry `json:"collection_log,omitempty"`
	UnlockedElements map[Element]bool
	// InjectionElement is the globally selected Element emitted by every
	// Injector. It is chosen from unlocked Elements in the Codex.
	InjectionElement Element
	// Owned is the total number of each component kind the player has ever
	// purchased. Monotonic: incremented by PurchaseComponent; never
	// decreased by Erase (removed components return to the available pool,
	// not the shop). Available = Owned - count-placed-on-grid.
	Owned       map[ComponentKind]int `json:"owned,omitempty"`
	MaxLoad     int
	CurrentLoad int
	// InjectionCooldownRemaining is the global manual-injection cooldown in
	// logical ticks. Zero means the Inject button may fire if Load allows.
	InjectionCooldownRemaining int `json:"injection_cooldown_remaining,omitempty"`
	TickRate                   int
	Ticks                      uint64

	// Modifiers aggregates active global upgrades. Derived from (future)
	// PurchasedUpgrades via rebuildModifiers; zero value is identity. Phase 1
	// carries the field through ApplyContext; Phase 2 (ADR 0010) fills the
	// struct fields and wires derivation.
	Modifiers GlobalModifiers `json:"modifiers,omitempty"`

	// ComponentTiers is the global tier level per component kind. Absent
	// entries default to sim.BaseTier. Tier upgrades purchased via
	// PurchaseTierUpgrade advance this map by one step.
	// See docs/adr/0011-component-tier-primitive.md.
	ComponentTiers map[ComponentKind]Tier `json:"component_tiers,omitempty"`
}

type ElementBestStats struct {
	MaxSpeed          int            `json:"max_speed,omitempty"`
	MaxMass           bignum.Decimal `json:"max_mass,omitempty"`
	MaxCollectedValue bignum.Decimal `json:"max_collected_value,omitempty"`
}

type CollectionLogEntry struct {
	Element       Element        `json:"element"`
	Mass          bignum.Decimal `json:"mass"`
	Speed         int            `json:"speed"`
	Magnetism     bignum.Decimal `json:"magnetism"`
	ResearchLevel int            `json:"research_level"`
	Value         bignum.Decimal `json:"value"`
	Tick          uint64         `json:"tick"`
}

func NewGameState() *GameState {
	return &GameState{
		Layer:            LayerGenesis,
		Grid:             NewGrid(),
		Research:         map[Element]int{},
		BestStats:        map[Element]ElementBestStats{},
		UnlockedElements: map[Element]bool{ElementHydrogen: true},
		InjectionElement: ElementHydrogen,
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

func (s *GameState) effectiveInjectionElement() Element {
	if _, ok := ElementCatalog[s.InjectionElement]; ok && IsElementUnlocked(s, s.InjectionElement) {
		return s.InjectionElement
	}
	return ElementHydrogen
}

func (s *GameState) normalizeInjectionElement() {
	s.InjectionElement = s.effectiveInjectionElement()
}

// EffectiveMaxLoad returns the grid-load cap after applying global upgrades.
// Base MaxLoad is the un-upgraded value; upgrade sources (globals, prestige,
// events) contribute flat bonuses on top. All call sites that enforce the
// cap should read this, not MaxLoad directly.
func (s *GameState) EffectiveMaxLoad() int {
	return s.MaxLoad + s.Modifiers.MaxLoadBonus
}

// researchView is the unexported read-only wrapper passed to components via
// ApplyContext.Research. Absent entries return 0.
type researchView struct{ m map[Element]int }

func newResearchView(m map[Element]int) ResearchView { return researchView{m: m} }

func (v researchView) Level(e Element) int {
	if v.m == nil {
		return 0
	}
	return v.m[e]
}
