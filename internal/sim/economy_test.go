package sim

import (
	"errors"
	"testing"

	"particleaccelerator/internal/bignum"
)

func TestCollectValueHydrogenBaseline(t *testing.T) {
	sub := Subject{Element: ElementHydrogen, Mass: bignum.FromInt(2), Speed: 3}
	got := collectValue(sub, GlobalModifiers{}.Normalized())
	want := bignum.FromInt(6) // speedK=1, Hydrogen multiplier 1.0, Mg=0
	if !got.Eq(want) {
		t.Fatalf("Hydrogen baseline: got %v want %v", got, want)
	}
}

func TestCollectValueHeliumMultiplier(t *testing.T) {
	sub := Subject{Element: ElementHelium, Mass: bignum.One(), Speed: 1}
	got := collectValue(sub, GlobalModifiers{}.Normalized())
	if !got.Eq(bignum.MustParse("2.5")) {
		t.Fatalf("Helium baseline: got %v want 2.5", got)
	}
}

func TestCollectValueMagnetismCoefficient(t *testing.T) {
	withMag := Subject{Element: ElementHydrogen, Mass: bignum.One(), Speed: 1, Magnetism: bignum.FromInt(4)}
	without := Subject{Element: ElementHydrogen, Mass: bignum.One(), Speed: 1}
	mods := GlobalModifiers{}.Normalized()
	got := collectValue(withMag, mods).Sub(collectValue(without, mods))
	want := bignum.FromInt(2)
	if !got.Eq(want) {
		t.Fatalf("magnetism delta: got %v want %v", got, want)
	}
}

func TestCollectValueBaseFormula(t *testing.T) {
	// value = (Mass×Speed×speedK + Magnetism×magK) × ElementMultiplier × CollectorValueMul.
	sub := Subject{Element: ElementHydrogen, Mass: bignum.FromInt(3), Speed: 4, Magnetism: bignum.FromInt(2)}
	got := collectValue(sub, GlobalModifiers{}.Normalized())
	// Hand-computed: 3×4×1 + 2×0.5 = 13; Hydrogen multiplier 1.0; value = 13.
	want := bignum.FromInt(13)
	if !got.Eq(want) {
		t.Fatalf("base formula: got %v want %v", got, want)
	}
}

func TestCollectValueCollectorMultiplier(t *testing.T) {
	sub := Subject{Element: ElementHydrogen, Mass: bignum.FromInt(2), Speed: 3}
	base := collectValue(sub, GlobalModifiers{}.Normalized())
	boosted := collectValue(sub, GlobalModifiers{CollectorValueMul: bignum.MustParse("1.5")}.Normalized())
	want := base.Mul(bignum.MustParse("1.5"))
	if !boosted.Eq(want) {
		t.Fatalf("CollectorValueMul 1.5×: got %v want %v", boosted, want)
	}
}

func TestIsElementPurchasable(t *testing.T) {
	s := NewGameState()
	// Hydrogen is already owned.
	if IsElementPurchasable(s, ElementHydrogen) {
		t.Fatalf("Hydrogen already owned should not be purchasable")
	}
	// Helium below threshold.
	if IsElementPurchasable(s, ElementHelium) {
		t.Fatalf("Helium below research threshold should not be purchasable")
	}
	// At threshold.
	s.Research[ElementHydrogen] = 10
	if !IsElementPurchasable(s, ElementHelium) {
		t.Fatalf("Helium at research threshold should be purchasable")
	}
	// Above threshold still purchasable.
	s.Research[ElementHydrogen] = 999
	if !IsElementPurchasable(s, ElementHelium) {
		t.Fatalf("Helium above threshold should be purchasable")
	}
	// After purchase, not purchasable.
	s.UnlockedElements[ElementHelium] = true
	if IsElementPurchasable(s, ElementHelium) {
		t.Fatalf("owned Helium should not be purchasable")
	}
}

func TestPurchaseElement(t *testing.T) {
	s := NewGameState()

	// Research too low.
	s.USD = bignum.FromInt(1000)
	if err := PurchaseElement(s, ElementHelium); !errors.Is(err, ErrResearchTooLow) {
		t.Fatalf("expected ErrResearchTooLow, got %v", err)
	}

	// Meets research, insufficient USD.
	s.Research[ElementHydrogen] = 10
	s.USD = bignum.FromInt(100)
	if err := PurchaseElement(s, ElementHelium); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}
	if s.UnlockedElements[ElementHelium] {
		t.Fatalf("Helium should not be unlocked on insufficient funds")
	}

	// Succeeds.
	s.USD = bignum.FromInt(750)
	if err := PurchaseElement(s, ElementHelium); err != nil {
		t.Fatalf("expected purchase success, got %v", err)
	}
	if !s.USD.Eq(bignum.FromInt(250)) {
		t.Fatalf("USD not deducted: got %v want 250", s.USD)
	}
	if !s.UnlockedElements[ElementHelium] {
		t.Fatalf("Helium not flagged as unlocked")
	}

	// Already owned.
	if err := PurchaseElement(s, ElementHelium); !errors.Is(err, ErrElementAlreadyOwned) {
		t.Fatalf("expected ErrElementAlreadyOwned, got %v", err)
	}

	// Unknown element.
	if err := PurchaseElement(s, Element("neon")); !errors.Is(err, ErrElementUnknown) {
		t.Fatalf("expected ErrElementUnknown, got %v", err)
	}
}
