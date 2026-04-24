package components

import (
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

func TestCompressorNoOpAtOrAboveOneSpeed(t *testing.T) {
	c := &Compressor{}
	ctx := sim.NewTestApplyContext()
	for _, speed := range []sim.Speed{sim.SpeedScale, sim.SpeedFromInt(2), sim.SpeedFromInt(5)} {
		in := sim.Subject{Speed: speed, Mass: bignum.FromInt(10)}
		out, lost := c.Apply(ctx, in)
		if lost {
			t.Fatalf("speed=%d: compressor should never destroy subjects", speed)
		}
		if !out.Mass.Eq(in.Mass) {
			t.Fatalf("speed=%d: mass should be unchanged at >=1.00 speed, got %v", speed, out.Mass)
		}
	}
}

func TestCompressorNoOpOnZeroSpeed(t *testing.T) {
	c := &Compressor{}
	ctx := sim.NewTestApplyContext()
	in := sim.Subject{Speed: 0, Mass: bignum.FromInt(10)}
	out, _ := c.Apply(ctx, in)
	if !out.Mass.Eq(in.Mass) {
		t.Fatalf("zero-speed subject should pass through untouched, got mass %v", out.Mass)
	}
}

func TestCompressorTenXAtPointOneSpeed(t *testing.T) {
	// Speed 0.10 = 10 in fixed-point; T1 coefficient = 1; multiplier = 10×.
	c := &Compressor{}
	ctx := sim.NewTestApplyContext()
	in := sim.Subject{Speed: sim.Speed(10), Mass: bignum.FromInt(3)}
	out, _ := c.Apply(ctx, in)
	if !out.Mass.Eq(bignum.FromInt(30)) {
		t.Fatalf("T1 @ speed 0.10: got mass %v want 30", out.Mass)
	}
}

func TestCompressorHalfSpeedDoublesMass(t *testing.T) {
	c := &Compressor{}
	ctx := sim.NewTestApplyContext()
	in := sim.Subject{Speed: sim.SpeedFromRatio(1, 2), Mass: bignum.FromInt(5)}
	out, _ := c.Apply(ctx, in)
	if !out.Mass.Eq(bignum.FromInt(10)) {
		t.Fatalf("T1 @ speed 0.50: got mass %v want 10", out.Mass)
	}
}

func TestCompressorTiersScale(t *testing.T) {
	cases := []struct {
		tier sim.Tier
		want bignum.Decimal
	}{
		{sim.BaseTier, bignum.FromInt(10)},        // coef 1 × 10 × mass 1
		{sim.Tier(2), bignum.MustParse("15")},     // coef 1.5 × 10 × mass 1
		{sim.Tier(3), bignum.FromInt(20)},         // coef 2   × 10 × mass 1
	}
	for _, tc := range cases {
		c := &Compressor{}
		ctx := sim.NewTestApplyContext()
		ctx.Tiers = testTierView(map[sim.ComponentKind]sim.Tier{sim.KindCompressor: tc.tier})
		// Speed 0.10 → ratio 10.
		out, _ := c.Apply(ctx, sim.Subject{Speed: sim.Speed(10), Mass: bignum.One()})
		if !out.Mass.Eq(tc.want) {
			t.Fatalf("tier %d @ speed 0.10: got %v want %v", tc.tier, out.Mass, tc.want)
		}
	}
}
