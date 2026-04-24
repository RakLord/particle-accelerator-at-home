package sim

import (
	"errors"

	"particleaccelerator/internal/bignum"
)

// Tier is the per-kind global progression level. Tier 1 is the baseline; every
// placed and inventoried instance of a tierable kind reads the same tier via
// ApplyContext.Tiers. See docs/adr/0011-component-tier-primitive.md.
type Tier int

// BaseTier is the starting tier for every component kind. GameState.ComponentTiers
// maps absent entries to BaseTier.
const BaseTier Tier = 1

// TierView is the read-only accessor for per-kind tier level, handed to
// components via ApplyContext. Absent entries return BaseTier.
type TierView interface {
	For(kind ComponentKind) Tier
}

// tierView is the unexported read-only wrapper. Absent-key fallback keeps
// component stat-table lookups from needing a nil-guard.
type tierView struct{ m map[ComponentKind]Tier }

func newTierView(m map[ComponentKind]Tier) TierView { return tierView{m: m} }

func (v tierView) For(kind ComponentKind) Tier {
	if v.m == nil {
		return BaseTier
	}
	t, ok := v.m[kind]
	if !ok || t < BaseTier {
		return BaseTier
	}
	return t
}

// ClampTier returns the tier index to use for a per-kind stat-table lookup,
// clamped to [BaseTier, maxTier]. Tier-table slices should reserve index 0
// unused and size (maxTier+1) so ClampTier returns a valid index.
// A nil TierView resolves to BaseTier.
func ClampTier(tiers TierView, kind ComponentKind, maxTier int) int {
	t := BaseTier
	if tiers != nil {
		t = tiers.For(kind)
	}
	if t < BaseTier {
		t = BaseTier
	}
	if int(t) > maxTier {
		return maxTier
	}
	return int(t)
}

// TierUpgradeInfo describes a single tier unlock for a given component kind.
// Entries in TierUpgradeCatalog are ordered by Tier ascending.
type TierUpgradeInfo struct {
	Tier             Tier
	Cost             bignum.Decimal
	RequiresElement  Element
	RequiresResearch int
}

// TierUpgradeCatalog is the shop-side data driving per-kind tier progression.
// Each entry list is ordered by Tier ascending; T(N) requires T(N-1) already
// owned. Kinds that don't tier (Injector, Rotator, Collector) have no entry.
var TierUpgradeCatalog = map[ComponentKind][]TierUpgradeInfo{
	KindAccelerator: {
		{Tier: 2, Cost: bignum.MustParse("500"), RequiresElement: ElementHydrogen, RequiresResearch: 3},
		{Tier: 3, Cost: bignum.MustParse("5000"), RequiresElement: ElementHydrogen, RequiresResearch: 15},
	},
	KindMagnetiser: {
		{Tier: 2, Cost: bignum.MustParse("800"), RequiresElement: ElementHydrogen, RequiresResearch: 5},
		{Tier: 3, Cost: bignum.MustParse("8000"), RequiresElement: ElementHydrogen, RequiresResearch: 20},
	},
	KindMeshGrid: {
		{Tier: 2, Cost: bignum.MustParse("400"), RequiresElement: ElementHydrogen, RequiresResearch: 4},
		{Tier: 3, Cost: bignum.MustParse("4000"), RequiresElement: ElementHydrogen, RequiresResearch: 18},
	},
	KindResonator: {
		{Tier: 2, Cost: bignum.MustParse("1200"), RequiresElement: ElementHydrogen, RequiresResearch: 8},
		{Tier: 3, Cost: bignum.MustParse("12000"), RequiresElement: ElementHydrogen, RequiresResearch: 25},
	},
	KindCatalyst: {
		{Tier: 2, Cost: bignum.MustParse("2000"), RequiresElement: ElementHydrogen, RequiresResearch: 30},
		{Tier: 3, Cost: bignum.MustParse("20000"), RequiresElement: ElementHelium, RequiresResearch: 10},
	},
	KindDuplicator: {
		{Tier: 2, Cost: bignum.MustParse("3000"), RequiresElement: ElementHydrogen, RequiresResearch: 12},
		{Tier: 3, Cost: bignum.MustParse("30000"), RequiresElement: ElementHelium, RequiresResearch: 15},
	},
	KindCompressor: {
		{Tier: 2, Cost: bignum.MustParse("15000"), RequiresElement: ElementHydrogen, RequiresResearch: 10},
		{Tier: 3, Cost: bignum.MustParse("150000"), RequiresElement: ElementHelium, RequiresResearch: 5},
	},
}

// NextTierUpgrade returns the next tier unlock available for kind, given the
// current tier level on s. Returns (_, false) if kind has no catalog entry or
// the player is already at max tier.
func NextTierUpgrade(s *GameState, kind ComponentKind) (TierUpgradeInfo, bool) {
	entries, ok := TierUpgradeCatalog[kind]
	if !ok {
		return TierUpgradeInfo{}, false
	}
	current := currentTier(s, kind)
	for _, e := range entries {
		if e.Tier == current+1 {
			return e, true
		}
	}
	return TierUpgradeInfo{}, false
}

// currentTier reads the persisted tier for kind, defaulting to BaseTier.
func currentTier(s *GameState, kind ComponentKind) Tier {
	if s.ComponentTiers == nil {
		return BaseTier
	}
	if t, ok := s.ComponentTiers[kind]; ok && t >= BaseTier {
		return t
	}
	return BaseTier
}

var (
	ErrNoTierUpgrade      = errors.New("sim: no tier upgrade available for this kind")
	ErrTierResearchTooLow = errors.New("sim: research threshold not met for tier upgrade")
)

// PurchaseTierUpgrade advances ComponentTiers[kind] by one if the next tier's
// cost and research prerequisites are met. Returns ErrNoTierUpgrade when kind
// is at max tier or has no catalog entry; ErrTierResearchTooLow when research
// is insufficient; ErrInsufficientFunds when USD is short. Successful purchase
// deducts cost and flips the tier level.
func PurchaseTierUpgrade(s *GameState, kind ComponentKind) error {
	up, ok := NextTierUpgrade(s, kind)
	if !ok {
		return ErrNoTierUpgrade
	}
	if up.RequiresElement != "" && s.Research[up.RequiresElement] < up.RequiresResearch {
		return ErrTierResearchTooLow
	}
	if s.USD.LT(up.Cost) {
		return ErrInsufficientFunds
	}
	s.USD = s.USD.Sub(up.Cost)
	if s.ComponentTiers == nil {
		s.ComponentTiers = map[ComponentKind]Tier{}
	}
	s.ComponentTiers[kind] = up.Tier
	return nil
}
