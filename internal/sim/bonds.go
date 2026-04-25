package sim

import (
	"errors"
	"math"

	"particleaccelerator/internal/bignum"
)

type BondID string

const (
	BondMethane   BondID = "methane"
	BondAcetylene BondID = "acetylene"
	BondEthylene  BondID = "ethylene"
	BondBenzene   BondID = "benzene"
	BondDiamond   BondID = "diamond"
)

type Bond struct {
	ID          BondID
	Name        string
	Formula     string
	Description string
	TokenCost   map[Element]int
	BondPoints  int
	Apply       func(m *GlobalModifiers)
}

var BondCatalogOrder = []BondID{
	BondMethane,
	BondAcetylene,
	BondEthylene,
	BondBenzene,
	BondDiamond,
}

var BondCatalog = map[BondID]Bond{
	BondMethane: {
		ID:          BondMethane,
		Name:        "Methane",
		Formula:     "CH4",
		Description: "The simplest hydrocarbon.",
		TokenCost:   map[Element]int{ElementCarbon: 1, ElementHydrogen: 4},
		BondPoints:  1,
		Apply: func(m *GlobalModifiers) {
			m.CollectorValueMul = multiplyDecimalModifier(m.CollectorValueMul, bignum.MustParse("1.15"))
		},
	},
	BondAcetylene: {
		ID:          BondAcetylene,
		Name:        "Acetylene",
		Formula:     "C2H2",
		Description: "A compact, high-energy hydrocarbon.",
		TokenCost:   map[Element]int{ElementCarbon: 2, ElementHydrogen: 2},
		BondPoints:  1,
		Apply: func(m *GlobalModifiers) {
			m.InjectorRateMul = multiplyDecimalModifier(m.InjectorRateMul, bignum.FromInt(4).DivInt(3))
		},
	},
	BondEthylene: {
		ID:          BondEthylene,
		Name:        "Ethylene",
		Formula:     "C2H4",
		Description: "A double-bonded hydrocarbon used as an accelerator feedstock.",
		TokenCost:   map[Element]int{ElementCarbon: 2, ElementHydrogen: 4},
		BondPoints:  2,
		Apply: func(m *GlobalModifiers) {
			m.AcceleratorSpeedBonus++
		},
	},
	BondBenzene: {
		ID:          BondBenzene,
		Name:        "Benzene",
		Formula:     "C6H6",
		Description: "A stable ring that unlocks idle operation.",
		TokenCost:   map[Element]int{ElementCarbon: 6, ElementHydrogen: 6},
		BondPoints:  3,
		Apply: func(m *GlobalModifiers) {
			m.AutoInjectEnabled = true
		},
	},
	BondDiamond: {
		ID:          BondDiamond,
		Name:        "Diamond",
		Formula:     "C12",
		Description: "A rigid carbon lattice that expands accelerator throughput.",
		TokenCost:   map[Element]int{ElementCarbon: 12},
		BondPoints:  3,
		Apply: func(m *GlobalModifiers) {
			m.MaxLoadBonus += 15
		},
	},
}

var (
	ErrInsufficientReserve = errors.New("sim: insufficient binder store reserve")
	ErrBondUnknown         = errors.New("sim: unknown bond")
	ErrBondAlreadyOwned    = errors.New("sim: bond already synthesised")
	ErrInsufficientTokens  = errors.New("sim: insufficient tokens")
)

const (
	tokenCostBase   = 10
	tokenCostGrowth = 5.3
)

// CrystallisationCost returns the reserve count required to mint the next
// Token for an Element. It uses a geometric formula with nice-number rounding,
// producing the early sequence 10, 50, 300, 1500, 8000 without hard-coding it.
func CrystallisationCost(e Element, owned int) int {
	if _, ok := ElementCatalog[e]; !ok {
		return 0
	}
	if owned <= 0 {
		return tokenCostBase
	}
	cost := float64(tokenCostBase)
	for range owned {
		cost = float64(roundNiceTokenCost(cost * tokenCostGrowth))
	}
	return int(cost)
}

func roundNiceTokenCost(raw float64) int {
	if raw <= 0 || math.IsNaN(raw) {
		return 0
	}
	if math.IsInf(raw, 1) {
		return int(^uint(0) >> 1)
	}
	exponent := math.Floor(math.Log10(raw))
	scale := math.Pow(10, exponent)
	mantissa := raw / scale
	nice := [...]float64{1, 1.5, 2, 3, 4, 5, 8, 10}
	best := nice[0]
	bestDiff := math.Abs(mantissa - best)
	for _, candidate := range nice[1:] {
		diff := math.Abs(mantissa - candidate)
		if diff < bestDiff {
			best = candidate
			bestDiff = diff
		}
	}
	rounded := int(math.Round(best * scale))
	if rounded < 1 {
		return 1
	}
	return rounded
}

func CrystalliseToken(s *GameState, e Element) error {
	if s == nil {
		return ErrElementUnknown
	}
	cost := CrystallisationCost(e, s.TokenInventory[e])
	if cost <= 0 {
		return ErrElementUnknown
	}
	if s.BinderReserves == nil {
		s.BinderReserves = map[Element]int{}
	}
	if s.TokenInventory == nil {
		s.TokenInventory = map[Element]int{}
	}
	if s.BinderReserves[e] < cost {
		return ErrInsufficientReserve
	}
	s.BinderReserves[e] -= cost
	s.TokenInventory[e]++
	return nil
}

func SynthesiseBond(s *GameState, id BondID) error {
	if s == nil {
		return ErrBondUnknown
	}
	bond, ok := BondCatalog[id]
	if !ok {
		return ErrBondUnknown
	}
	if s.BondsState == nil {
		s.BondsState = map[BondID]bool{}
	}
	if s.TokenInventory == nil {
		s.TokenInventory = map[Element]int{}
	}
	if s.BondsState[id] {
		return ErrBondAlreadyOwned
	}
	for e, n := range bond.TokenCost {
		if s.TokenInventory[e] < n {
			return ErrInsufficientTokens
		}
	}
	for e, n := range bond.TokenCost {
		s.TokenInventory[e] -= n
	}
	s.BondsState[id] = true
	s.BondPoints += bond.BondPoints
	rebuildModifiers(s)
	return nil
}
