package sim

import (
	"testing"

	"particleaccelerator/internal/bignum"
)

func TestGlobalModifiersNormalizedPromotesZeroDecimals(t *testing.T) {
	n := GlobalModifiers{}.Normalized()
	if !n.CollectorValueMul.Eq(bignum.One()) {
		t.Fatalf("CollectorValueMul: got %v want 1", n.CollectorValueMul)
	}
	if !n.InjectorRateMul.Eq(bignum.One()) {
		t.Fatalf("InjectorRateMul: got %v want 1", n.InjectorRateMul)
	}
	if !n.MagnetiserBonusMul.Eq(bignum.One()) {
		t.Fatalf("MagnetiserBonusMul: got %v want 1", n.MagnetiserBonusMul)
	}
	if !n.BinderStoreCapacityMul.Eq(bignum.One()) {
		t.Fatalf("BinderStoreCapacityMul: got %v want 1", n.BinderStoreCapacityMul)
	}
	// Integer fields stay zero — zero is the additive identity.
	if n.AcceleratorSpeedBonus != 0 || n.ResearchPerCollectBonus != 0 || n.MaxLoadBonus != 0 {
		t.Fatalf("integer fields should remain zero: %+v", n)
	}
}

func TestGlobalModifiersNormalizedPreservesNonZeroValues(t *testing.T) {
	in := GlobalModifiers{
		CollectorValueMul:      bignum.MustParse("1.25"),
		InjectorRateMul:        bignum.MustParse("2"),
		MagnetiserBonusMul:     bignum.MustParse("3"),
		BinderStoreCapacityMul: bignum.MustParse("4"),
	}
	n := in.Normalized()
	if !n.CollectorValueMul.Eq(in.CollectorValueMul) ||
		!n.InjectorRateMul.Eq(in.InjectorRateMul) ||
		!n.MagnetiserBonusMul.Eq(in.MagnetiserBonusMul) ||
		!n.BinderStoreCapacityMul.Eq(in.BinderStoreCapacityMul) {
		t.Fatalf("non-zero Decimal fields mutated by Normalized: got %+v", n)
	}
}

func TestEffectiveMaxLoadAddsBonus(t *testing.T) {
	s := NewGameState()
	base := s.MaxLoad
	if got := s.EffectiveMaxLoad(); got != base {
		t.Fatalf("no-bonus EffectiveMaxLoad: got %d want %d", got, base)
	}
	s.Modifiers.MaxLoadBonus = 7
	if got := s.EffectiveMaxLoad(); got != base+7 {
		t.Fatalf("with bonus: got %d want %d", got, base+7)
	}
}
