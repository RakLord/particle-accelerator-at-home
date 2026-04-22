package sim

import (
	"errors"
	"math"
	"testing"
)

func TestCollectValueHydrogenBaseline(t *testing.T) {
	sub := Subject{Element: ElementHydrogen, Mass: 2, Speed: 3}
	got := collectValue(sub, 0)
	want := 2.0 * 3.0 * 1.0 // speedK=1, research=0 → multiplier 1.0, Mg=0
	if got != want {
		t.Fatalf("Hydrogen baseline: got %v want %v", got, want)
	}
}

func TestCollectValueHeliumMultiplier(t *testing.T) {
	sub := Subject{Element: ElementHelium, Mass: 1, Speed: 1}
	got := collectValue(sub, 0)
	if got != 2.5 {
		t.Fatalf("Helium baseline: got %v want 2.5", got)
	}
}

func TestCollectValueResearchMultiplier(t *testing.T) {
	sub := Subject{Element: ElementHydrogen, Mass: 1, Speed: 1}
	// research=50 → 1 + 50/50 = 2×
	got := collectValue(sub, 50)
	if got != 2.0 {
		t.Fatalf("research multiplier: got %v want 2.0", got)
	}
}

func TestCollectValueMagnetismCoefficient(t *testing.T) {
	withMag := Subject{Element: ElementHydrogen, Mass: 1, Speed: 1, Magnetism: 4}
	without := Subject{Element: ElementHydrogen, Mass: 1, Speed: 1}
	got := collectValue(withMag, 0) - collectValue(without, 0)
	want := 4.0 * 0.5
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("magnetism delta: got %v want %v", got, want)
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
	s.USD = 1000
	if err := PurchaseElement(s, ElementHelium); !errors.Is(err, ErrResearchTooLow) {
		t.Fatalf("expected ErrResearchTooLow, got %v", err)
	}

	// Meets research, insufficient USD.
	s.Research[ElementHydrogen] = 10
	s.USD = 100
	if err := PurchaseElement(s, ElementHelium); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}
	if s.UnlockedElements[ElementHelium] {
		t.Fatalf("Helium should not be unlocked on insufficient funds")
	}

	// Succeeds.
	s.USD = 750
	if err := PurchaseElement(s, ElementHelium); err != nil {
		t.Fatalf("expected purchase success, got %v", err)
	}
	if s.USD != 250 {
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
