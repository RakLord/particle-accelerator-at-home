package sim

import (
	"errors"

	"particleaccelerator/internal/bignum"
)

type Element string

const (
	ElementHydrogen   Element = "hydrogen"
	ElementHelium     Element = "helium"
	ElementLithium    Element = "lithium"
	ElementBeryllium  Element = "beryllium"
	ElementBoron      Element = "boron"
	ElementCarbon     Element = "carbon"
	ElementNitrogen   Element = "nitrogen"
	ElementOxygen     Element = "oxygen"
	ElementFluorine   Element = "fluorine"
	ElementNeon       Element = "neon"
	ElementSodium     Element = "sodium"
	ElementMagnesium  Element = "magnesium"
	ElementAluminium  Element = "aluminium"
	ElementSilicon    Element = "silicon"
	ElementPhosphorus Element = "phosphorus"
	ElementSulfur     Element = "sulfur"
	ElementChlorine   Element = "chlorine"
	ElementArgon      Element = "argon"
	ElementPotassium  Element = "potassium"
	ElementCalcium    Element = "calcium"
)

// ElementInfo is the per-Element metadata used for value, UI, and unlock gating.
// Add new Elements by appending here and to CatalogOrder.
type ElementInfo struct {
	AtomicNumber      int
	Name              string
	Symbol            string
	Period            int
	Group             int
	BaseMass          bignum.Decimal
	BaseSpeed         Speed
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
		BaseMass:     bignum.MustParse("1.008"),
		BaseSpeed:    SpeedFromInt(2),
		Multiplier:   bignum.MustParse("1"),
	},
	ElementHelium: {
		AtomicNumber:      2,
		Name:              "Helium",
		Symbol:            "He",
		Period:            1,
		Group:             18,
		BaseMass:          bignum.MustParse("4.003"),
		BaseSpeed:         SpeedFromInt(2),
		Multiplier:        bignum.MustParse("1.5"),
		UnlocksFrom:       ElementHydrogen,
		ResearchThreshold: 10,
		UnlockCost:        bignum.MustParse("500"),
	},
	ElementLithium: {
		AtomicNumber:      3,
		Name:              "Lithium",
		Symbol:            "Li",
		Period:            2,
		Group:             1,
		BaseMass:          bignum.MustParse("6.94"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("1.8"),
		UnlocksFrom:       ElementHelium,
		ResearchThreshold: 12,
		UnlockCost:        bignum.MustParse("2000"),
	},
	ElementBeryllium: {
		AtomicNumber:      4,
		Name:              "Beryllium",
		Symbol:            "Be",
		Period:            2,
		Group:             2,
		BaseMass:          bignum.MustParse("9.012"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("2.1"),
		UnlocksFrom:       ElementLithium,
		ResearchThreshold: 14,
		UnlockCost:        bignum.MustParse("8000"),
	},
	ElementBoron: {
		AtomicNumber:      5,
		Name:              "Boron",
		Symbol:            "B",
		Period:            2,
		Group:             13,
		BaseMass:          bignum.MustParse("10.81"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("2.4"),
		UnlocksFrom:       ElementBeryllium,
		ResearchThreshold: 16,
		UnlockCost:        bignum.MustParse("30000"),
	},
	ElementCarbon: {
		AtomicNumber:      6,
		Name:              "Carbon",
		Symbol:            "C",
		Period:            2,
		Group:             14,
		BaseMass:          bignum.MustParse("12.011"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("2.8"),
		UnlocksFrom:       ElementBoron,
		ResearchThreshold: 18,
		UnlockCost:        bignum.MustParse("100000"),
	},
	ElementNitrogen: {
		AtomicNumber:      7,
		Name:              "Nitrogen",
		Symbol:            "N",
		Period:            2,
		Group:             15,
		BaseMass:          bignum.MustParse("14.007"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("3.2"),
		UnlocksFrom:       ElementCarbon,
		ResearchThreshold: 20,
		UnlockCost:        bignum.MustParse("300000"),
	},
	ElementOxygen: {
		AtomicNumber:      8,
		Name:              "Oxygen",
		Symbol:            "O",
		Period:            2,
		Group:             16,
		BaseMass:          bignum.MustParse("15.999"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("3.6"),
		UnlocksFrom:       ElementNitrogen,
		ResearchThreshold: 22,
		UnlockCost:        bignum.MustParse("900000"),
	},
	ElementFluorine: {
		AtomicNumber:      9,
		Name:              "Fluorine",
		Symbol:            "F",
		Period:            2,
		Group:             17,
		BaseMass:          bignum.MustParse("18.998"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("3.8"),
		UnlocksFrom:       ElementOxygen,
		ResearchThreshold: 24,
		UnlockCost:        bignum.MustParse("2500000"),
	},
	ElementNeon: {
		AtomicNumber:      10,
		Name:              "Neon",
		Symbol:            "Ne",
		Period:            2,
		Group:             18,
		BaseMass:          bignum.MustParse("20.180"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("4"),
		UnlocksFrom:       ElementFluorine,
		ResearchThreshold: 26,
		UnlockCost:        bignum.MustParse("7500000"),
	},
	ElementSodium: {
		AtomicNumber:      11,
		Name:              "Sodium",
		Symbol:            "Na",
		Period:            3,
		Group:             1,
		BaseMass:          bignum.MustParse("22.990"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("4.5"),
		UnlocksFrom:       ElementNeon,
		ResearchThreshold: 28,
		UnlockCost:        bignum.MustParse("20000000"),
	},
	ElementMagnesium: {
		AtomicNumber:      12,
		Name:              "Magnesium",
		Symbol:            "Mg",
		Period:            3,
		Group:             2,
		BaseMass:          bignum.MustParse("24.305"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("5"),
		UnlocksFrom:       ElementSodium,
		ResearchThreshold: 30,
		UnlockCost:        bignum.MustParse("55000000"),
	},
	ElementAluminium: {
		AtomicNumber:      13,
		Name:              "Aluminium",
		Symbol:            "Al",
		Period:            3,
		Group:             13,
		BaseMass:          bignum.MustParse("26.982"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("5.7"),
		UnlocksFrom:       ElementMagnesium,
		ResearchThreshold: 32,
		UnlockCost:        bignum.MustParse("150000000"),
	},
	ElementSilicon: {
		AtomicNumber:      14,
		Name:              "Silicon",
		Symbol:            "Si",
		Period:            3,
		Group:             14,
		BaseMass:          bignum.MustParse("28.085"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("6.5"),
		UnlocksFrom:       ElementAluminium,
		ResearchThreshold: 34,
		UnlockCost:        bignum.MustParse("400000000"),
	},
	ElementPhosphorus: {
		AtomicNumber:      15,
		Name:              "Phosphorus",
		Symbol:            "P",
		Period:            3,
		Group:             15,
		BaseMass:          bignum.MustParse("30.974"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("7.2"),
		UnlocksFrom:       ElementSilicon,
		ResearchThreshold: 36,
		UnlockCost:        bignum.MustParse("1000000000"),
	},
	ElementSulfur: {
		AtomicNumber:      16,
		Name:              "Sulfur",
		Symbol:            "S",
		Period:            3,
		Group:             16,
		BaseMass:          bignum.MustParse("32.06"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("8"),
		UnlocksFrom:       ElementPhosphorus,
		ResearchThreshold: 38,
		UnlockCost:        bignum.MustParse("2500000000"),
	},
	ElementChlorine: {
		AtomicNumber:      17,
		Name:              "Chlorine",
		Symbol:            "Cl",
		Period:            3,
		Group:             17,
		BaseMass:          bignum.MustParse("35.45"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("8.6"),
		UnlocksFrom:       ElementSulfur,
		ResearchThreshold: 40,
		UnlockCost:        bignum.MustParse("6000000000"),
	},
	ElementArgon: {
		AtomicNumber:      18,
		Name:              "Argon",
		Symbol:            "Ar",
		Period:            3,
		Group:             18,
		BaseMass:          bignum.MustParse("39.948"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("9"),
		UnlocksFrom:       ElementChlorine,
		ResearchThreshold: 42,
		UnlockCost:        bignum.MustParse("15000000000"),
	},
	ElementPotassium: {
		AtomicNumber:      19,
		Name:              "Potassium",
		Symbol:            "K",
		Period:            4,
		Group:             1,
		BaseMass:          bignum.MustParse("39.098"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("10"),
		UnlocksFrom:       ElementArgon,
		ResearchThreshold: 45,
		UnlockCost:        bignum.MustParse("40000000000"),
	},
	ElementCalcium: {
		AtomicNumber:      20,
		Name:              "Calcium",
		Symbol:            "Ca",
		Period:            4,
		Group:             2,
		BaseMass:          bignum.MustParse("40.078"),
		BaseSpeed:         SpeedFromInt(1),
		Multiplier:        bignum.MustParse("12"),
		UnlocksFrom:       ElementPotassium,
		ResearchThreshold: 50,
		UnlockCost:        bignum.MustParse("100000000000"),
	},
}

// CatalogOrder is the deterministic display order for the Periodic Table.
var CatalogOrder = []Element{
	ElementHydrogen,
	ElementHelium,
	ElementLithium,
	ElementBeryllium,
	ElementBoron,
	ElementCarbon,
	ElementNitrogen,
	ElementOxygen,
	ElementFluorine,
	ElementNeon,
	ElementSodium,
	ElementMagnesium,
	ElementAluminium,
	ElementSilicon,
	ElementPhosphorus,
	ElementSulfur,
	ElementChlorine,
	ElementArgon,
	ElementPotassium,
	ElementCalcium,
}

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
	base := s.Mass.Mul(s.Speed.Decimal()).Mul(speedValueK).Add(s.Magnetism.Mul(magValueK))
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
