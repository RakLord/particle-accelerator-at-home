package sim

import (
	"errors"

	"particleaccelerator/internal/bignum"
)

type Element string

const (
	ElementHydrogen Element = "hydrogen"
	ElementHelium   Element = "helium"
)

// ElementInfo is the per-Element metadata used for value, UI, and unlock gating.
// Add new Elements by appending here and to CatalogOrder.
type ElementInfo struct {
	AtomicNumber      int
	Name              string
	Symbol            string
	Period            int
	Group             int
	Multiplier        bignum.Decimal
	UnlocksFrom       Element
	ResearchThreshold int
	UnlockCost        bignum.Decimal
}

var ElementCatalog = map[Element]ElementInfo{
	ElementHydrogen: {
		AtomicNumber: 1,
		Name:         "Hydrogen",
		Symbol:       "H",
		Period:       1,
		Group:        1,
		Multiplier:   bignum.MustParse("1"),
	},
	ElementHelium: {
		AtomicNumber:      2,
		Name:              "Helium",
		Symbol:            "He",
		Period:            1,
		Group:             18,
		Multiplier:        bignum.MustParse("2.5"),
		UnlocksFrom:       ElementHydrogen,
		ResearchThreshold: 10,
		UnlockCost:        bignum.MustParse("500"),
	},
}

// CatalogOrder is the deterministic display order for the Periodic Table.
var CatalogOrder = []Element{ElementHydrogen, ElementHelium}

// Value formula constants. See docs/features/value-formula.md.
var (
	speedValueK = bignum.MustParse("1")
	magValueK   = bignum.MustParse("0.5")
)

// collectValue is the $USD awarded when a Subject is collected. mods must be
// Normalized so Decimal fields multiply safely.
// See docs/features/value-formula.md and docs/adr/0010-global-modifier-pipeline.md.
func collectValue(s Subject, mods GlobalModifiers) bignum.Decimal {
	info := ElementCatalog[s.Element]
	base := s.Mass.MulInt(s.Speed).Mul(speedValueK).Add(s.Magnetism.Mul(magValueK))
	return base.Mul(info.Multiplier).Mul(mods.CollectorValueMul)
}

var (
	ErrElementUnknown      = errors.New("sim: unknown element")
	ErrElementAlreadyOwned = errors.New("sim: element already owned")
	ErrResearchTooLow      = errors.New("sim: research threshold not met")
	ErrInsufficientFunds   = errors.New("sim: insufficient USD")
)

// IsElementUnlocked reports whether the player has purchased the Element.
func IsElementUnlocked(s *GameState, e Element) bool {
	return s.UnlockedElements[e]
}

// IsElementPurchasable reports whether the Element can be bought right now:
// not already owned, and the prerequisite research threshold is met.
func IsElementPurchasable(s *GameState, e Element) bool {
	if s.UnlockedElements[e] {
		return false
	}
	info, ok := ElementCatalog[e]
	if !ok {
		return false
	}
	if info.UnlocksFrom != "" && s.Research[info.UnlocksFrom] < info.ResearchThreshold {
		return false
	}
	return true
}

// PurchaseElement deducts UnlockCost and flips the UnlockedElements flag.
// Returns a sentinel error on any of: unknown element, already owned, research
// too low, or insufficient USD.
func PurchaseElement(s *GameState, e Element) error {
	info, ok := ElementCatalog[e]
	if !ok {
		return ErrElementUnknown
	}
	if s.UnlockedElements[e] {
		return ErrElementAlreadyOwned
	}
	if info.UnlocksFrom != "" && s.Research[info.UnlocksFrom] < info.ResearchThreshold {
		return ErrResearchTooLow
	}
	if s.USD.LT(info.UnlockCost) {
		return ErrInsufficientFunds
	}
	s.USD = s.USD.Sub(info.UnlockCost)
	if s.UnlockedElements == nil {
		s.UnlockedElements = map[Element]bool{}
	}
	s.UnlockedElements[e] = true
	return nil
}
