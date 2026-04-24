package components

import (
	"testing"

	"particleaccelerator/internal/bignum"
)

func TestEffectiveSpawnIntervalIdentity(t *testing.T) {
	cases := []struct {
		base int
		mul  bignum.Decimal
	}{
		{30, bignum.Zero()},    // zero rateMul treated as identity
		{30, bignum.One()},     // explicit 1×
		{30, bignum.MustParse("0.5")}, // below 1× won't slow down
	}
	for _, c := range cases {
		if got := effectiveSpawnInterval(c.base, c.mul); got != c.base {
			t.Fatalf("rateMul=%v: got %d want %d", c.mul, got, c.base)
		}
	}
}

func TestEffectiveSpawnIntervalShortensAtRateMulAboveOne(t *testing.T) {
	// rateMul=2 → interval 30 becomes 15.
	if got := effectiveSpawnInterval(30, bignum.MustParse("2")); got != 15 {
		t.Fatalf("rateMul=2: got %d want 15", got)
	}
	// rateMul=3 → 30 / 3 = 10.
	if got := effectiveSpawnInterval(30, bignum.MustParse("3")); got != 10 {
		t.Fatalf("rateMul=3: got %d want 10", got)
	}
	// Floor at 1 tick even if rateMul would drive effective below.
	if got := effectiveSpawnInterval(5, bignum.MustParse("100")); got != 1 {
		t.Fatalf("rateMul=100 floor: got %d want 1", got)
	}
}
