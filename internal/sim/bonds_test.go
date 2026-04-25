package sim

import (
	"errors"
	"math"
	"testing"

	"particleaccelerator/internal/bignum"
)

func TestCrystallisationCostFormulaMatchesEarlyTargets(t *testing.T) {
	want := []int{5, 10, 15, 50, 50, 100, 200, 400}
	for owned, expected := range want {
		if got := CrystallisationCost(ElementCarbon, owned); got != expected {
			t.Fatalf("CrystallisationCost owned=%d: got %d want %d", owned, got, expected)
		}
	}
}

func TestCrystalliseTokenConsumesReserve(t *testing.T) {
	s := NewGameState()
	s.BinderReserves[ElementCarbon] = 12
	if err := CrystalliseToken(s, ElementCarbon); err != nil {
		t.Fatalf("CrystalliseToken: %v", err)
	}
	if got := s.BinderReserves[ElementCarbon]; got != 7 {
		t.Fatalf("reserve after crystallise: got %d want 7", got)
	}
	if got := s.TokenInventory[ElementCarbon]; got != 1 {
		t.Fatalf("tokens after crystallise: got %d want 1", got)
	}
}

func TestCrystalliseTokenRejectsInsufficientReserve(t *testing.T) {
	s := NewGameState()
	s.BinderReserves[ElementHydrogen] = 4
	err := CrystalliseToken(s, ElementHydrogen)
	if !errors.Is(err, ErrInsufficientReserve) {
		t.Fatalf("CrystalliseToken err = %v want ErrInsufficientReserve", err)
	}
	if got := s.TokenInventory[ElementHydrogen]; got != 0 {
		t.Fatalf("token changed on failed crystallise: got %d", got)
	}
}

func TestSynthesiseBondConsumesTokensAwardsBPAndRebuildsModifiers(t *testing.T) {
	s := NewGameState()
	s.TokenInventory[ElementCarbon] = 1
	s.TokenInventory[ElementHydrogen] = 4
	if err := SynthesiseBond(s, BondMethane); err != nil {
		t.Fatalf("SynthesiseBond: %v", err)
	}
	if !s.BondsState[BondMethane] {
		t.Fatalf("Methane not marked owned")
	}
	if got := s.TokenInventory[ElementCarbon]; got != 0 {
		t.Fatalf("Carbon tokens after synthesis: got %d want 0", got)
	}
	if got := s.TokenInventory[ElementHydrogen]; got != 0 {
		t.Fatalf("Hydrogen tokens after synthesis: got %d want 0", got)
	}
	if got := s.BondPoints; got != 1 {
		t.Fatalf("BondPoints: got %d want 1", got)
	}
	if got := s.Modifiers.Normalized().CollectorValueMul; !got.Eq(bignum.MustParse("1.15")) {
		t.Fatalf("CollectorValueMul: got %v want 1.15", got)
	}
}

func TestSynthesiseBondIsAllOrNothing(t *testing.T) {
	s := NewGameState()
	s.TokenInventory[ElementCarbon] = 1
	err := SynthesiseBond(s, BondMethane)
	if !errors.Is(err, ErrInsufficientTokens) {
		t.Fatalf("SynthesiseBond err = %v want ErrInsufficientTokens", err)
	}
	if got := s.TokenInventory[ElementCarbon]; got != 1 {
		t.Fatalf("Carbon token was partially deducted: got %d", got)
	}
	if s.BondsState[BondMethane] || s.BondPoints != 0 {
		t.Fatalf("failed synthesis mutated state: bonds=%v BP=%d", s.BondsState, s.BondPoints)
	}
}

func TestAcetyleneUsesRateMultiplierSemantics(t *testing.T) {
	s := NewGameState()
	s.TokenInventory[ElementCarbon] = 2
	s.TokenInventory[ElementHydrogen] = 2
	if err := SynthesiseBond(s, BondAcetylene); err != nil {
		t.Fatalf("SynthesiseBond Acetylene: %v", err)
	}
	got := s.Modifiers.Normalized().InjectorRateMul.Float64()
	if math.Abs(got-(4.0/3.0)) > 1e-9 {
		t.Fatalf("InjectorRateMul = %v want 4/3", got)
	}
	base := DefaultInjectionCooldownSeconds * DefaultTickRate
	if cd := s.EffectiveInjectionCooldownTicks(); cd >= base {
		t.Fatalf("Acetylene did not shorten cooldown: got %d base %d", cd, base)
	}
}
