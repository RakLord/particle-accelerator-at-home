package sim

// GlobalModifiers aggregates player-purchased global upgrades. Phase 1 lands
// the type as an empty struct so it can ride on ApplyContext; fields arrive
// in Phase 2 (ADR 0010). The zero value is the no-upgrades identity.
//
// Components and hot paths read fields directly; the struct is never mutated
// from gameplay code — only the (future) rebuildModifiers(s) helper writes to
// it, derived from GameState.PurchasedUpgrades.
type GlobalModifiers struct {
	// Intentionally empty. Phase 2 adds CollectorValueMul, InjectorRateMul,
	// AcceleratorSpeedBonus, MagnetiserBonusMul, ResearchPerCollectBonus,
	// MaxLoadBonus. See docs/adr/0010-global-modifier-pipeline.md.
}

// Normalized returns a copy with zero-valued Decimal fields promoted to 1 so
// downstream multiplication is safe. Phase 1 is a no-op; Phase 2 fills in the
// normalization logic when Decimal fields exist.
func (m GlobalModifiers) Normalized() GlobalModifiers { return m }
