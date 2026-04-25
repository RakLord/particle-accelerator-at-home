package sim

import (
	"errors"

	"particleaccelerator/internal/bignum"
)

type LabUpgradeID string

const (
	LabCovalentMemory  LabUpgradeID = "covalent_memory"
	LabStableIsotope   LabUpgradeID = "stable_isotope"
	LabChainReaction   LabUpgradeID = "chain_reaction"
	LabCarbonCore      LabUpgradeID = "carbon_core"
	LabDensePacking    LabUpgradeID = "dense_packing"
	LabAutoInjectSpeed LabUpgradeID = "auto_inject_speed"
)

type LabApplyPhase int

const (
	LabApplyModifiers LabApplyPhase = iota
	LabApplyResetSeed
)

type LabUpgrade struct {
	ID           LabUpgradeID
	Name         string
	Description  string
	MaxPurchases int
	CostByLevel  []int
	AppliesIn    LabApplyPhase
	Apply        func(m *GlobalModifiers, s *GameState, level int)
}

var LabCatalogOrder = []LabUpgradeID{
	LabCovalentMemory,
	LabStableIsotope,
	LabChainReaction,
	LabCarbonCore,
	LabDensePacking,
	LabAutoInjectSpeed,
}

var LabCatalog = map[LabUpgradeID]LabUpgrade{
	LabCovalentMemory: {
		ID:           LabCovalentMemory,
		Name:         "Covalent Memory",
		Description:  "Start each run with Helium already unlocked.",
		MaxPurchases: 1,
		CostByLevel:  []int{1},
		AppliesIn:    LabApplyResetSeed,
		Apply: func(_ *GlobalModifiers, s *GameState, _ int) {
			if s.UnlockedElements == nil {
				s.UnlockedElements = map[Element]bool{}
			}
			s.UnlockedElements[ElementHydrogen] = true
			s.UnlockedElements[ElementHelium] = true
		},
	},
	LabStableIsotope: {
		ID:           LabStableIsotope,
		Name:         "Stable Isotope",
		Description:  "Per-Element research carries over at 30% on prestige.",
		MaxPurchases: 1,
		CostByLevel:  []int{2},
		AppliesIn:    LabApplyResetSeed,
		Apply:        func(_ *GlobalModifiers, _ *GameState, _ int) {},
	},
	LabChainReaction: {
		ID:           LabChainReaction,
		Name:         "Chain Reaction",
		Description:  "Injectors start at 2x base emission rate.",
		MaxPurchases: 1,
		CostByLevel:  []int{2},
		AppliesIn:    LabApplyModifiers,
		Apply: func(m *GlobalModifiers, _ *GameState, _ int) {
			m.InjectorRateMul = multiplyDecimalModifier(m.InjectorRateMul, bignum.FromInt(2))
		},
	},
	LabCarbonCore: {
		ID:           LabCarbonCore,
		Name:         "Carbon Core",
		Description:  "Start each run with Hydrogen through Carbon unlocked.",
		MaxPurchases: 1,
		CostByLevel:  []int{1},
		AppliesIn:    LabApplyResetSeed,
		Apply: func(_ *GlobalModifiers, s *GameState, _ int) {
			unlockElementsThrough(s, ElementCarbon)
		},
	},
	LabDensePacking: {
		ID:           LabDensePacking,
		Name:         "Dense Packing",
		Description:  "Doubles Binder Store capacity per level.",
		MaxPurchases: 5,
		CostByLevel:  []int{1, 2, 3, 4, 5},
		AppliesIn:    LabApplyModifiers,
		Apply: func(m *GlobalModifiers, _ *GameState, level int) {
			if level > 5 {
				level = 5
			}
			m.BinderStoreCapacityMul = multiplyDecimalModifier(m.BinderStoreCapacityMul, bignum.FromInt(1<<level))
		},
	},
	LabAutoInjectSpeed: {
		ID:           LabAutoInjectSpeed,
		Name:         "Auto-Inject Speed",
		Description:  "Reduces Auto-Inject cadence from 10s to 4s over four levels.",
		MaxPurchases: 4,
		CostByLevel:  []int{1, 2, 3, 4},
		AppliesIn:    LabApplyModifiers,
		Apply:        func(_ *GlobalModifiers, _ *GameState, _ int) {},
	},
}

var (
	ErrLabUpgradeUnknown       = errors.New("sim: unknown laboratory upgrade")
	ErrLabUpgradeMaxed         = errors.New("sim: laboratory upgrade maxed")
	ErrLabUpgradeMisconfigured = errors.New("sim: laboratory upgrade misconfigured")
	ErrInsufficientBondPoints  = errors.New("sim: insufficient bond points")
)

func PurchaseLabUpgrade(s *GameState, id LabUpgradeID) error {
	if s == nil {
		return ErrLabUpgradeUnknown
	}
	upgrade, ok := LabCatalog[id]
	if !ok {
		return ErrLabUpgradeUnknown
	}
	if s.LaboratoryUpgrades == nil {
		s.LaboratoryUpgrades = map[LabUpgradeID]int{}
	}
	cur := s.LaboratoryUpgrades[id]
	if cur >= upgrade.MaxPurchases {
		return ErrLabUpgradeMaxed
	}
	if cur < 0 || cur >= len(upgrade.CostByLevel) {
		return ErrLabUpgradeMisconfigured
	}
	cost := upgrade.CostByLevel[cur]
	if s.BondPoints < cost {
		return ErrInsufficientBondPoints
	}
	s.BondPoints -= cost
	s.LaboratoryUpgrades[id] = cur + 1
	rebuildModifiers(s)
	return nil
}

func unlockElementsThrough(s *GameState, target Element) {
	if s.UnlockedElements == nil {
		s.UnlockedElements = map[Element]bool{}
	}
	for _, e := range CatalogOrder {
		s.UnlockedElements[e] = true
		if e == target {
			return
		}
	}
}

func computeAutoInjectCadence(s *GameState) int {
	if s == nil || !s.BondsState[BondBenzene] {
		return 0
	}
	secondsByLevel := [...]int{10, 8, 6, 5, 4}
	level := s.LaboratoryUpgrades[LabAutoInjectSpeed]
	if level < 0 {
		level = 0
	}
	if level >= len(secondsByLevel) {
		level = len(secondsByLevel) - 1
	}
	tickRate := s.TickRate
	if tickRate <= 0 {
		tickRate = DefaultTickRate
	}
	return secondsByLevel[level] * tickRate
}
