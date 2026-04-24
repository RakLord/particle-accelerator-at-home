package sim

import (
	"errors"
	"testing"

	"particleaccelerator/internal/bignum"
)

func TestTierViewAbsentKindReturnsBase(t *testing.T) {
	v := newTierView(nil)
	if got := v.For(KindAccelerator); got != BaseTier {
		t.Fatalf("nil map: got %d want %d", got, BaseTier)
	}
	v = newTierView(map[ComponentKind]Tier{})
	if got := v.For(KindAccelerator); got != BaseTier {
		t.Fatalf("empty map: got %d want %d", got, BaseTier)
	}
}

func TestTierViewBelowBaseIsClampedToBase(t *testing.T) {
	v := newTierView(map[ComponentKind]Tier{KindAccelerator: Tier(0)})
	if got := v.For(KindAccelerator); got != BaseTier {
		t.Fatalf("Tier 0 clamp: got %d want %d", got, BaseTier)
	}
}

func TestClampTierBounds(t *testing.T) {
	v := newTierView(map[ComponentKind]Tier{KindAccelerator: Tier(9)})
	if got := ClampTier(v, KindAccelerator, 3); got != 3 {
		t.Fatalf("above-max clamp: got %d want 3", got)
	}
	if got := ClampTier(nil, KindAccelerator, 5); got != int(BaseTier) {
		t.Fatalf("nil view: got %d want %d", got, BaseTier)
	}
}

func TestNextTierUpgradeNone(t *testing.T) {
	s := NewGameState()
	// Non-tierable kind has no catalog entry.
	if _, ok := NextTierUpgrade(s, KindInjector); ok {
		t.Fatal("Injector should have no tier upgrade available")
	}
}

func TestNextTierUpgradeProgression(t *testing.T) {
	s := NewGameState()
	// From base tier, next is T2.
	up, ok := NextTierUpgrade(s, KindAccelerator)
	if !ok || up.Tier != Tier(2) {
		t.Fatalf("fresh state: got %+v ok=%v want T2", up, ok)
	}
	// After T2 purchased, next is T3.
	s.ComponentTiers = map[ComponentKind]Tier{KindAccelerator: Tier(2)}
	up, ok = NextTierUpgrade(s, KindAccelerator)
	if !ok || up.Tier != Tier(3) {
		t.Fatalf("at T2: got %+v ok=%v want T3", up, ok)
	}
	// At max tier, no further upgrade.
	s.ComponentTiers[KindAccelerator] = Tier(3)
	if _, ok := NextTierUpgrade(s, KindAccelerator); ok {
		t.Fatal("at max tier: expected no further upgrade")
	}
}

func TestPurchaseTierUpgradeErrors(t *testing.T) {
	s := NewGameState()
	s.USD = bignum.FromInt(100) // too poor
	if err := PurchaseTierUpgrade(s, KindInjector); !errors.Is(err, ErrNoTierUpgrade) {
		t.Fatalf("no catalog entry: got %v want ErrNoTierUpgrade", err)
	}
	// Accelerator T2 requires Hydrogen research ≥ 3.
	if err := PurchaseTierUpgrade(s, KindAccelerator); !errors.Is(err, ErrTierResearchTooLow) {
		t.Fatalf("research gate: got %v want ErrTierResearchTooLow", err)
	}
	s.Research[ElementHydrogen] = 10
	if err := PurchaseTierUpgrade(s, KindAccelerator); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("funds gate: got %v want ErrInsufficientFunds", err)
	}
}

func TestPurchaseTierUpgradeHappyPath(t *testing.T) {
	s := NewGameState()
	s.Research[ElementHydrogen] = 10
	s.USD = bignum.FromInt(1000)
	if err := PurchaseTierUpgrade(s, KindAccelerator); err != nil {
		t.Fatalf("T2 purchase: %v", err)
	}
	if got := s.ComponentTiers[KindAccelerator]; got != Tier(2) {
		t.Fatalf("after T2 purchase: got tier %d want 2", got)
	}
	// Cost of 500 was deducted.
	if !s.USD.Eq(bignum.FromInt(500)) {
		t.Fatalf("USD after T2: got %v want 500", s.USD)
	}
}

func TestPurchaseTierUpgradeAtMaxReturnsErrNoTierUpgrade(t *testing.T) {
	s := NewGameState()
	s.ComponentTiers = map[ComponentKind]Tier{KindAccelerator: Tier(3)}
	s.Research[ElementHydrogen] = 999
	s.USD = bignum.FromInt(1_000_000)
	if err := PurchaseTierUpgrade(s, KindAccelerator); !errors.Is(err, ErrNoTierUpgrade) {
		t.Fatalf("at max tier: got %v want ErrNoTierUpgrade", err)
	}
}
