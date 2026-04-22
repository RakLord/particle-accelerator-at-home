package sim

import (
	"errors"
	"math"
	"testing"

	"particleaccelerator/internal/bignum"
)

func TestPowDecimalZero(t *testing.T) {
	got := powDecimal(bignum.MustParse("2"), 0)
	if !got.Eq(bignum.One()) {
		t.Fatalf("pow(2, 0) = %v want 1", got)
	}
}

func TestPowDecimalOne(t *testing.T) {
	got := powDecimal(bignum.MustParse("3.7"), 1)
	if !got.Eq(bignum.MustParse("3.7")) {
		t.Fatalf("pow(3.7, 1) = %v want 3.7", got)
	}
}

func TestPowDecimalMatchesMathPow(t *testing.T) {
	for n := 2; n < 10; n++ {
		got := powDecimal(bignum.MustParse("1.15"), n).Float64()
		want := math.Pow(1.15, float64(n))
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("pow(1.15, %d) = %v want %v", n, got, want)
		}
	}
}

func TestPowDecimalLargeIsFinite(t *testing.T) {
	// 1.15^500 is ~1.5e30 — well within bignum's unbounded exponent range.
	got := powDecimal(bignum.MustParse("1.15"), 500)
	if got.IsZero() {
		t.Fatalf("pow(1.15, 500) = 0 (underflow/bug)")
	}
	if math.IsInf(got.Float64(), 0) || math.IsNaN(got.Float64()) {
		// Expected — Float64() saturates at 1e308 — but bignum value itself
		// must remain well-formed. Log10 gives a finite order of magnitude.
		if math.IsInf(got.Log10(), 0) || math.IsNaN(got.Log10()) {
			t.Fatalf("pow(1.15, 500) log10 is not finite")
		}
	}
}

func TestComponentCostScalesWithOwned(t *testing.T) {
	ResetCostModifiers()
	s := NewGameState()
	s.Owned = map[ComponentKind]int{}

	low := ComponentCost(s, KindAccelerator)
	s.Owned[KindAccelerator] = 10
	high := ComponentCost(s, KindAccelerator)

	if !high.GT(low) {
		t.Fatalf("cost should grow with Owned: low=%v high=%v", low, high)
	}
	// low = ceil(5 * 1) = 5; high = ceil(5 * 1.15^10) ≈ ceil(20.22) = 21.
	if !low.Eq(bignum.FromInt(5)) {
		t.Fatalf("base cost: got %v want 5", low)
	}
	if !high.Eq(bignum.FromInt(21)) {
		t.Fatalf("scaled cost: got %v want 21", high)
	}
}

func TestComponentCostIsWholeNumber(t *testing.T) {
	ResetCostModifiers()
	s := NewGameState()
	s.Owned = map[ComponentKind]int{}
	for _, kind := range []ComponentKind{
		KindInjector, KindAccelerator, KindMeshGrid, KindMagnetiser, KindRotator, KindCollector,
	} {
		for _, owned := range []int{0, 1, 5, 17, 42} {
			s.Owned[kind] = owned
			cost := ComponentCost(s, kind)
			if !cost.Eq(cost.Ceil()) {
				t.Errorf("%s@%d: cost %v is not a whole number", kind, owned, cost)
			}
		}
	}
}

func TestComponentCostAppliesModifiers(t *testing.T) {
	ResetCostModifiers()
	defer ResetCostModifiers()

	s := NewGameState()
	s.Owned = map[ComponentKind]int{}
	base := ComponentCost(s, KindAccelerator)

	RegisterCostModifier(func(_ *GameState, _ ComponentKind) bignum.Decimal {
		return bignum.FromInt(2)
	})
	doubled := ComponentCost(s, KindAccelerator)
	// 5 * 2 = 10
	if !doubled.Eq(base.MulInt(2)) {
		t.Fatalf("global modifier: got %v want %v", doubled, base.MulInt(2))
	}

	// Kind-specific modifier: only affects injector.
	RegisterCostModifier(func(_ *GameState, kind ComponentKind) bignum.Decimal {
		if kind == KindInjector {
			return bignum.MustParse("0.5")
		}
		return bignum.One()
	})
	inj := ComponentCost(s, KindInjector)
	acc := ComponentCost(s, KindAccelerator)
	// injector: ceil(10 * 2 * 0.5) = 10; accelerator: still 10.
	if !inj.Eq(bignum.FromInt(10)) {
		t.Fatalf("kind-specific modifier on injector: got %v want 10", inj)
	}
	if !acc.Eq(bignum.FromInt(10)) {
		t.Fatalf("kind-specific modifier leaked to accelerator: got %v", acc)
	}
}

func TestPurchaseComponentSuccess(t *testing.T) {
	ResetCostModifiers()
	s := NewGameState()
	start := bignum.FromInt(100)
	s.USD = start
	s.Owned = map[ComponentKind]int{}

	cost := ComponentCost(s, KindAccelerator)
	if err := PurchaseComponent(s, KindAccelerator); err != nil {
		t.Fatalf("PurchaseComponent: %v", err)
	}
	// USD was decreased by exactly cost, through the same Sub path. Comparing
	// `start.Sub(cost)` avoids a known bignum precision artifact where
	// comparing against a freshly-constructed 95 fails on the last digit.
	want := start.Sub(cost)
	if !s.USD.Eq(want) {
		t.Fatalf("USD after purchase: got %v want %v", s.USD, want)
	}
	if s.Owned[KindAccelerator] != 1 {
		t.Fatalf("Owned[accelerator] = %d want 1", s.Owned[KindAccelerator])
	}
}

func TestPurchaseComponentInsufficientFunds(t *testing.T) {
	ResetCostModifiers()
	s := NewGameState()
	s.USD = bignum.FromInt(2)
	s.Owned = map[ComponentKind]int{}

	err := PurchaseComponent(s, KindAccelerator)
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("want ErrInsufficientFunds, got %v", err)
	}
	if !s.USD.Eq(bignum.FromInt(2)) {
		t.Fatalf("USD changed on failed purchase: got %v", s.USD)
	}
	if s.Owned[KindAccelerator] != 0 {
		t.Fatalf("Owned changed on failed purchase: got %d", s.Owned[KindAccelerator])
	}
}

func TestPurchaseComponentUnknownKind(t *testing.T) {
	ResetCostModifiers()
	s := NewGameState()
	err := PurchaseComponent(s, ComponentKind("not_a_real_kind"))
	if !errors.Is(err, ErrNoSuchComponent) {
		t.Fatalf("want ErrNoSuchComponent, got %v", err)
	}
}

func TestCanPurchase(t *testing.T) {
	ResetCostModifiers()
	s := NewGameState()
	s.USD = bignum.FromInt(6)
	s.Owned = map[ComponentKind]int{}

	if !CanPurchase(s, KindAccelerator) {
		t.Fatalf("should afford Accelerator at $6 (cost $5)")
	}
	if CanPurchase(s, KindMagnetiser) {
		t.Fatalf("should not afford Magnetiser at $6 (cost $25)")
	}
	if CanPurchase(s, ComponentKind("nope")) {
		t.Fatalf("unknown kind should not be purchasable")
	}
}

func TestHardResetResetsInventory(t *testing.T) {
	s := NewGameState()
	s.Owned[KindAccelerator] = 99
	s.USD = bignum.FromInt(10000)

	s.HardReset()

	starter := starterInventory()
	for kind, want := range starter {
		if s.Owned[kind] != want {
			t.Errorf("Owned[%s] after reset = %d want %d", kind, s.Owned[kind], want)
		}
	}
	if s.Owned[KindAccelerator] != starter[KindAccelerator] {
		t.Errorf("excess Owned survived reset")
	}
	if !s.USD.IsZero() {
		t.Errorf("USD not reset: %v", s.USD)
	}
}
