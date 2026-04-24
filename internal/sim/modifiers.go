package sim

import "particleaccelerator/internal/bignum"

// GlobalModifiers aggregates player-purchased global upgrades. Decimal fields
// are multiplicative (zero value means "no bonus" after Normalized()). Integer
// fields are additive.
//
// Components and hot paths read fields directly; the struct is never mutated
// from gameplay code — only rebuildModifiers(s) (Phase 6) writes to it,
// derived from GameState.PurchasedUpgrades.
//
// See docs/adr/0010-global-modifier-pipeline.md.
type GlobalModifiers struct {
	CollectorValueMul       bignum.Decimal `json:"collector_value_mul,omitempty"`
	ComponentCostMul        bignum.Decimal `json:"component_cost_mul,omitempty"`
	InjectorRateMul         bignum.Decimal `json:"injector_rate_mul,omitempty"`
	AcceleratorSpeedBonus   int            `json:"accelerator_speed_bonus,omitempty"`
	MagnetiserBonusMul      bignum.Decimal `json:"magnetiser_bonus_mul,omitempty"`
	ResearchPerCollectBonus int            `json:"research_per_collect_bonus,omitempty"`
	MaxLoadBonus            int            `json:"max_load_bonus,omitempty"`
}

// Normalized returns a copy with zero-valued Decimal fields promoted to 1 so
// downstream multiplication is safe. Integer fields are left as-is — their
// zero is the additive identity.
//
// The tick loop calls this once per tick when building ApplyContext; read
// sites can then multiply without checking for the zero-value case.
func (m GlobalModifiers) Normalized() GlobalModifiers {
	if m.CollectorValueMul.IsZero() {
		m.CollectorValueMul = bignum.One()
	}
	if m.ComponentCostMul.IsZero() {
		m.ComponentCostMul = bignum.One()
	}
	if m.InjectorRateMul.IsZero() {
		m.InjectorRateMul = bignum.One()
	}
	if m.MagnetiserBonusMul.IsZero() {
		m.MagnetiserBonusMul = bignum.One()
	}
	return m
}
