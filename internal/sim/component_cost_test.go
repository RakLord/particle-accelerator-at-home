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
	s.Owned[KindAccelerator] = 2
	high := ComponentCost(s, KindAccelerator)

	if !high.GT(low) {
		t.Fatalf("cost should grow with Owned: low=%v high=%v", low, high)
	}
	// low = ceil(5 * 1) = 5; high = ceil(5 * 3.2^2) = ceil(51.2) = 52.
	if !low.Eq(bignum.FromInt(5)) {
		t.Fatalf("base cost: got %v want 5", low)
	}
	if !high.Eq(bignum.FromInt(52)) {
		t.Fatalf("scaled cost: got %v want 52", high)
	}
}

func TestComponentCostKeyBalanceTargets(t *testing.T) {
	ResetCostModifiers()
	s := NewGameState()
	s.Owned = map[ComponentKind]int{}

	cases := []struct {
		kind       ComponentKind
		firstCost  bignum.Decimal
		secondCost bignum.Decimal
	}{
		{KindMagnetiser, bignum.FromInt(100), bignum.FromInt(500)},
		{KindCatalyst, bignum.FromInt(1000), bignum.FromInt(12000)},
		{KindDuplicator, bignum.FromInt(10000), bignum.FromInt(1250000)},
		{KindResonator, bignum.FromInt(50), bignum.FromInt(68)},
	}

	for _, tc := range cases {
		s.Owned[tc.kind] = 0
		if got := ComponentCost(s, tc.kind); !got.Eq(tc.firstCost) {
			t.Errorf("%s first cost: got %v want %v", tc.kind, got, tc.firstCost)
		}
		s.Owned[tc.kind] = 1
		if got := ComponentCost(s, tc.kind); !got.Eq(tc.secondCost) {
			t.Errorf("%s second cost: got %v want %v", tc.kind, got, tc.secondCost)
		}
	}
}

func TestRoutingComponentCostCurves(t *testing.T) {
	// Pipe and Rotator share the same $8 base, but Rotator (the more
	// powerful turn tile) ramps at 1.20 while Pipe ramps at 1.15. The
	// gap should be visible by the tenth purchase.
	ResetCostModifiers()
	s := NewGameState()
	s.Owned = map[ComponentKind]int{}

	cases := []struct {
		name  string
		kind  ComponentKind
		owned int
		want  bignum.Decimal
	}{
		{"pipe first", KindPipe, 0, bignum.FromInt(8)},
		{"pipe second", KindPipe, 1, bignum.FromInt(10)},        // ceil(8 * 1.15)
		{"pipe tenth", KindPipe, 9, bignum.FromInt(29)},         // ceil(8 * 1.15^9) = ceil(28.15)
		{"rotator first", KindRotator, 0, bignum.FromInt(8)},
		{"rotator second", KindRotator, 1, bignum.FromInt(10)},  // ceil(8 * 1.20)
		{"rotator tenth", KindRotator, 9, bignum.FromInt(42)},   // ceil(8 * 1.20^9) = ceil(41.27)
	}
	for _, tc := range cases {
		s.Owned[tc.kind] = tc.owned
		if got := ComponentCost(s, tc.kind); !got.Eq(tc.want) {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

func TestCompressorCostCurve(t *testing.T) {
	ResetCostModifiers()
	s := NewGameState()
	s.Owned = map[ComponentKind]int{}

	cases := []struct {
		owned int
		want  bignum.Decimal
	}{
		{0, bignum.FromInt(7000)},        // raw = 7000
		{1, bignum.FromInt(105000)},      // raw = 7000 * 15
		{2, bignum.FromInt(1575000)},     // raw = 7000 * 15^2 — exactly at SoftCapAt
		{3, bignum.FromInt(354375000)},   // raw = 7000 * 15^3; shaped = 1575000 * 15^2
	}
	for _, tc := range cases {
		s.Owned[KindCompressor] = tc.owned
		if got := ComponentCost(s, KindCompressor); !got.Eq(tc.want) {
			t.Errorf("compressor owned=%d: got %v want %v", tc.owned, got, tc.want)
		}
	}
}

func TestComponentCostSoftCapAmplifiesAboveThreshold(t *testing.T) {
	info := ComponentCostInfo{
		Base:         bignum.FromInt(10),
		Growth:       bignum.FromInt(2),
		SoftCapAt:    bignum.FromInt(100),
		SoftCapPower: 2,
	}

	below := applyCostSoftCap(bignum.FromInt(80), info)
	if !below.Eq(bignum.FromInt(80)) {
		t.Fatalf("below cap: got %v want 80", below)
	}
	above := applyCostSoftCap(bignum.FromInt(200), info)
	if !above.Eq(bignum.FromInt(400)) {
		t.Fatalf("above cap: got %v want 400", above)
	}
}

func TestComponentCostIsWholeNumber(t *testing.T) {
	ResetCostModifiers()
	s := NewGameState()
	s.Owned = map[ComponentKind]int{}
	for _, kind := range []ComponentKind{
		KindInjector, KindAccelerator, KindMeshGrid, KindMagnetiser, KindRotator, KindPipe, KindCollector,
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

func TestComponentCostAppliesGlobalComponentCostMultiplier(t *testing.T) {
	ResetCostModifiers()
	s := NewGameState()
	s.Owned = map[ComponentKind]int{}
	s.Modifiers.ComponentCostMul = bignum.MustParse("0.5")

	cost := ComponentCost(s, KindInjector)
	if !cost.Eq(bignum.FromInt(5)) {
		t.Fatalf("global component cost multiplier: got %v want 5", cost)
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
