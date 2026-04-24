package sim

import (
	"errors"

	"particleaccelerator/internal/bignum"
)

// ComponentCostInfo is the per-kind tuning data for component purchase cost.
// The pre-modifier cost of the next purchase is Base * Growth^Owned. If
// SoftCapAt is non-zero and the raw cost exceeds it, the excess is amplified as
// SoftCapAt * (raw / SoftCapAt)^SoftCapPower. The resulting cost is multiplied
// by global and registered modifiers, then ceilinged to a whole-dollar integer.
type ComponentCostInfo struct {
	Base         bignum.Decimal
	Growth       bignum.Decimal
	SoftCapAt    bignum.Decimal
	SoftCapPower int
}

// ComponentCatalog holds the base cost and per-unit growth factor for every
// purchasable component kind. KindCollector is a catalog-only key — see
// internal/sim/kinds.go.
var ComponentCatalog = map[ComponentKind]ComponentCostInfo{
	KindInjector:    {Base: bignum.MustParse("10"), Growth: bignum.MustParse("1.15")},
	KindAccelerator: {Base: bignum.MustParse("5"), Growth: bignum.MustParse("3.2"), SoftCapAt: bignum.MustParse("5000"), SoftCapPower: 2},
	KindMeshGrid:    {Base: bignum.MustParse("15"), Growth: bignum.MustParse("1.20")},
	KindMagnetiser:  {Base: bignum.MustParse("100"), Growth: bignum.MustParse("5")},
	KindRotator:     {Base: bignum.MustParse("8"), Growth: bignum.MustParse("1.20")},
	KindPipe:        {Base: bignum.MustParse("8"), Growth: bignum.MustParse("1.15")},
	KindCollector:   {Base: bignum.MustParse("50"), Growth: bignum.MustParse("1.25")},
	// Phase 4 components — prices to be refined after playtest.
	KindResonator:  {Base: bignum.MustParse("50"), Growth: bignum.MustParse("4")},
	KindCatalyst:   {Base: bignum.MustParse("1000"), Growth: bignum.MustParse("12")},
	KindDuplicator: {Base: bignum.MustParse("10000"), Growth: bignum.MustParse("125")},
	KindCompressor: {Base: bignum.MustParse("7000"), Growth: bignum.MustParse("15"), SoftCapAt: bignum.MustParse("1575000"), SoftCapPower: 2},
}

// CostModifier is the extension point for prestige / research / event effects
// that scale component purchase cost. A modifier returns the multiplier to
// apply for (state, kind); the product of every registered modifier feeds
// into ComponentCost.
type CostModifier func(s *GameState, kind ComponentKind) bignum.Decimal

var costModifiers []CostModifier

// RegisterCostModifier appends a modifier to the global list. Intended for
// prestige-layer upgrades at load/init time; not safe for concurrent use.
func RegisterCostModifier(m CostModifier) { costModifiers = append(costModifiers, m) }

// ResetCostModifiers clears all registered modifiers. Test-only helper.
func ResetCostModifiers() { costModifiers = nil }

var (
	ErrNoSuchComponent = errors.New("sim: unknown component kind")
)

// ComponentCost returns the $USD cost of the next purchase of kind, ceilinged
// to a whole-dollar integer. Unknown kinds return Zero().
func ComponentCost(s *GameState, kind ComponentKind) bignum.Decimal {
	info, ok := ComponentCatalog[kind]
	if !ok {
		return bignum.Zero()
	}
	owned := 0
	if s.Owned != nil {
		owned = s.Owned[kind]
	}
	mult := bignum.One()
	for _, m := range costModifiers {
		mult = mult.Mul(m(s, kind))
	}
	global := s.Modifiers.Normalized().ComponentCostMul
	raw := applyCostSoftCap(info.Base.Mul(powDecimal(info.Growth, owned)), info)
	raw = raw.Mul(global).Mul(mult)
	return raw.Ceil()
}

func applyCostSoftCap(raw bignum.Decimal, info ComponentCostInfo) bignum.Decimal {
	if info.SoftCapAt.IsZero() || info.SoftCapPower <= 1 || raw.LTE(info.SoftCapAt) {
		return raw
	}
	ratio := raw.Div(info.SoftCapAt)
	return info.SoftCapAt.Mul(powDecimal(ratio, info.SoftCapPower))
}

// CanPurchase reports whether s has enough USD to buy one more of kind.
func CanPurchase(s *GameState, kind ComponentKind) bool {
	if _, ok := ComponentCatalog[kind]; !ok {
		return false
	}
	return s.USD.GTE(ComponentCost(s, kind))
}

// PurchaseComponent deducts ComponentCost(kind) from USD and increments
// Owned[kind]. Returns ErrNoSuchComponent for unknown kinds or
// ErrInsufficientFunds if unaffordable.
func PurchaseComponent(s *GameState, kind ComponentKind) error {
	if _, ok := ComponentCatalog[kind]; !ok {
		return ErrNoSuchComponent
	}
	cost := ComponentCost(s, kind)
	if s.USD.LT(cost) {
		return ErrInsufficientFunds
	}
	s.USD = s.USD.Sub(cost)
	if s.Owned == nil {
		s.Owned = map[ComponentKind]int{}
	}
	s.Owned[kind]++
	return nil
}

// CountPlaced returns the number of cells currently occupied by kind.
// Collectors are counted via Cell.IsCollector rather than Component.Kind.
func CountPlaced(s *GameState, kind ComponentKind) int {
	if s.Grid == nil {
		return 0
	}
	n := 0
	for y := range s.Grid.Cells {
		for x := range s.Grid.Cells[y] {
			c := s.Grid.Cells[y][x]
			if kind == KindCollector {
				if c.IsCollector {
					n++
				}
				continue
			}
			if c.Component != nil && c.Component.Kind() == kind {
				n++
			}
		}
	}
	return n
}

// CountAvailable returns the count of kind in inventory (owned minus placed).
// Never negative.
func CountAvailable(s *GameState, kind ComponentKind) int {
	owned := 0
	if s.Owned != nil {
		owned = s.Owned[kind]
	}
	avail := owned - CountPlaced(s, kind)
	if avail < 0 {
		return 0
	}
	return avail
}

// powDecimal returns base raised to the integer power n. For n == 0 returns
// One. Repeated-multiplication implementation is fine for the cost-scaling
// range; bignum exponents are unbounded so no overflow concern.
func powDecimal(base bignum.Decimal, n int) bignum.Decimal {
	if n <= 0 {
		return bignum.One()
	}
	// Exponentiation by squaring keeps this O(log n) for very large Owned.
	result := bignum.One()
	acc := base
	for n > 0 {
		if n&1 == 1 {
			result = result.Mul(acc)
		}
		n >>= 1
		if n > 0 {
			acc = acc.Mul(acc)
		}
	}
	return result
}
