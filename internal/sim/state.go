package sim

import "particleaccelerator/internal/bignum"

const (
	DefaultMaxLoad = 1
	// DefaultInjectionCooldownSeconds is the base manual injection cooldown.
	// Injector-rate upgrades shorten the effective tick count derived from this.
	DefaultInjectionCooldownSeconds = 5
	// MaxCollectionLogEntries is the number of recent collected Subjects kept
	// for the in-game collection log.
	MaxCollectionLogEntries = 10
	// MaxNotificationLogEntries is the number of logged helper notifications kept
	// for the in-game notification history.
	MaxNotificationLogEntries = 50
	// DefaultTickRate: interpolation is live (see docs/features/0005-smooth-motion.md)
	// so this is no longer a rendering constraint — raising it is a gameplay
	// decision about how fast the grid feels.
	DefaultTickRate = 10
)

type GameState struct {
	Layer           Layer
	Grid            *Grid
	USD             bignum.Decimal
	Research        map[Element]int
	BestStats       map[Element]ElementBestStats
	CollectionLog   []CollectionLogEntry `json:"collection_log,omitempty"`
	NotificationLog []NotificationEntry  `json:"notification_log,omitempty"`
	// ShownHelperMilestones records one-shot helper IDs that have already been
	// shown in this save. HardReset clears it by restoring NewGameState defaults.
	ShownHelperMilestones map[string]bool `json:"shown_helper_milestones,omitempty"`
	UnlockedElements      map[Element]bool
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

	// Modifiers aggregates active global effects. It is derived from durable
	// unlock state by rebuildModifiers; zero-valued Decimal fields read as
	// identity after Normalized().
	Modifiers GlobalModifiers `json:"modifiers,omitempty"`

	// ComponentTiers is the global tier level per component kind. Absent
	// entries default to sim.BaseTier. Tier upgrades purchased via
	// PurchaseTierUpgrade advance this map by one step.
	// See docs/adr/0011-component-tier-primitive.md.
	ComponentTiers map[ComponentKind]Tier `json:"component_tiers,omitempty"`

	// Prestige-layer state. These fields are additive save data; empty values
	// mean the save has not entered the Carbon reset loop yet.
	BinderReserves        map[Element]int      `json:"binder_reserves,omitempty"`
	TokenInventory        map[Element]int      `json:"token_inventory,omitempty"`
	BondsState            map[BondID]bool      `json:"bonds_state,omitempty"`
	BondPoints            int                  `json:"bond_points,omitempty"`
	LaboratoryUpgrades    map[LabUpgradeID]int `json:"laboratory_upgrades,omitempty"`
	AutoInjectActive      bool                 `json:"auto_inject_active,omitempty"`
	AutoInjectTickCounter int                  `json:"-"`
	RunCount              int                  `json:"run_count,omitempty"`
}

type ElementBestStats struct {
	MaxSpeed          Speed          `json:"max_speed,omitempty"`
	MaxMass           bignum.Decimal `json:"max_mass,omitempty"`
	MaxCollectedValue bignum.Decimal `json:"max_collected_value,omitempty"`
}

type CollectionLogEntry struct {
	Element       Element        `json:"element"`
	Mass          bignum.Decimal `json:"mass"`
	Speed         Speed          `json:"speed"`
	Magnetism     bignum.Decimal `json:"magnetism"`
	ResearchLevel int            `json:"research_level"`
	Value         bignum.Decimal `json:"value"`
	Tick          uint64         `json:"tick"`
}

type NotificationEntry struct {
	Header string `json:"header"`
	Body   string `json:"body"`
	// TimeHHMM is render-supplied local time at notification creation. Keeping it
	// as display text avoids time-zone interpretation in older saves.
	TimeHHMM string `json:"time_hhmm,omitempty"`
	Tick     uint64 `json:"tick,omitempty"`
}

func NewGameState() *GameState {
	return &GameState{
		Layer:                 LayerGenesis,
		Grid:                  NewGrid(),
		Research:              map[Element]int{},
		BestStats:             map[Element]ElementBestStats{},
		ShownHelperMilestones: map[string]bool{},
		UnlockedElements:      map[Element]bool{ElementHydrogen: true},
		InjectionElement:      ElementHydrogen,
		Owned:                 starterInventory(),
		MaxLoad:               DefaultMaxLoad,
		TickRate:              DefaultTickRate,
		BinderReserves:        map[Element]int{},
		TokenInventory:        map[Element]int{},
		BondsState:            map[BondID]bool{},
		LaboratoryUpgrades:    map[LabUpgradeID]int{},
	}
}

// starterInventory is the set of components a brand-new game begins with.
// The player starts with only an Injector and a Collector; everything else
// (pipes, accelerators, elbows, etc.) must be purchased. Tuning these
// numbers is a design lever — see docs/features/0007-component-cost.md.
func starterInventory() map[ComponentKind]int {
	return map[ComponentKind]int{
		KindInjector:  1,
		KindCollector: 1,
	}
}

// HardReset wipes in-memory state back to defaults. The caller is responsible
// for persisting the reset (or deleting the save) so a subsequent Load does
// not restore the previous state.
func (s *GameState) HardReset() {
	*s = *NewGameState()
}

// ResetGenesis starts a fresh Genesis run while preserving durable prestige
// progression. It is intentionally separate from HardReset, which wipes every
// prestige field too.
func ResetGenesis(s *GameState) {
	if s == nil {
		return
	}
	var researchSnapshot map[Element]int
	if s.LaboratoryUpgrades[LabStableIsotope] > 0 {
		researchSnapshot = copyElementIntMap(s.Research)
	}

	s.USD = bignum.Zero()
	s.Grid = NewGrid()
	s.Research = map[Element]int{}
	s.CollectionLog = nil
	s.NotificationLog = nil
	s.ShownHelperMilestones = map[string]bool{}
	s.UnlockedElements = map[Element]bool{ElementHydrogen: true}
	s.InjectionElement = ElementHydrogen
	s.Owned = starterInventory()
	s.MaxLoad = DefaultMaxLoad
	s.CurrentLoad = 0
	s.InjectionCooldownRemaining = 0
	s.TickRate = DefaultTickRate
	s.Ticks = 0
	s.ComponentTiers = nil
	s.BinderReserves = map[Element]int{}
	s.TokenInventory = map[Element]int{}
	s.AutoInjectTickCounter = 0
	s.RunCount++
	if s.Layer == "" {
		s.Layer = LayerGenesis
	}
	if s.BondsState == nil {
		s.BondsState = map[BondID]bool{}
	}
	if s.LaboratoryUpgrades == nil {
		s.LaboratoryUpgrades = map[LabUpgradeID]int{}
	}
	if s.BestStats == nil {
		s.BestStats = map[Element]ElementBestStats{}
	}

	for id, level := range s.LaboratoryUpgrades {
		if level <= 0 {
			continue
		}
		upgrade, ok := LabCatalog[id]
		if !ok || upgrade.AppliesIn != LabApplyResetSeed || upgrade.Apply == nil {
			continue
		}
		upgrade.Apply(nil, s, level)
	}
	if researchSnapshot != nil {
		for e, v := range researchSnapshot {
			if v <= 0 {
				continue
			}
			s.Research[e] = v * 30 / 100
		}
	}
	rebuildModifiers(s)
}

func copyElementIntMap(in map[Element]int) map[Element]int {
	if len(in) == 0 {
		return map[Element]int{}
	}
	out := make(map[Element]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
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
