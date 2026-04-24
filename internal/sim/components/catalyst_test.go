package components

import (
	"math"
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

// catalystThreshold exposes the unexported threshold for test-side assertions.
func catalystThreshold() int { return catalystResearchThreshold }

func TestCatalystInertBelowThreshold(t *testing.T) {
	c := &Catalyst{}
	ctx := sim.NewTestApplyContext()
	ctx.Research = stubResearch{sim.ElementHydrogen: catalystThreshold() - 1}
	in := sim.Subject{Element: sim.ElementHydrogen, Mass: bignum.FromInt(2)}
	out, lost := c.Apply(ctx, in)
	if lost {
		t.Fatal("catalyst should never destroy subjects")
	}
	if !out.Mass.Eq(in.Mass) {
		t.Fatalf("below threshold should be inert: got %v want %v", out.Mass, in.Mass)
	}
}

// At exactly the threshold the log-curve multiplier is 1.0 — the component
// activates but the effective factor is unity (no cliff).
func TestCatalystFlatAtThreshold(t *testing.T) {
	c := &Catalyst{}
	ctx := sim.NewTestApplyContext()
	ctx.Research = stubResearch{sim.ElementHydrogen: catalystThreshold()}
	in := sim.Subject{Element: sim.ElementHydrogen, Mass: bignum.FromInt(2)}
	out, _ := c.Apply(ctx, in)
	if got := out.Mass.Float64(); math.Abs(got-2) > 1e-9 {
		t.Fatalf("at threshold multiplier must be 1.0: got %v want 2", got)
	}
}

// Above the threshold, higher tiers steepen the curve. Pick research such
// that log10(research - 24) = 2 so expected values are integer multiples of
// k: T1=1+0.7·2=2.4, T2=1+0.95·2=2.9, T3=1+1.25·2=3.5.
func TestCatalystTiersScale(t *testing.T) {
	cases := []struct {
		tier sim.Tier
		want float64
	}{
		{sim.BaseTier, 2.4 * 2},
		{sim.Tier(2), 2.9 * 2},
		{sim.Tier(3), 3.5 * 2},
	}
	for _, c := range cases {
		cat := &Catalyst{}
		ctx := sim.NewTestApplyContext()
		ctx.Research = stubResearch{sim.ElementHydrogen: catalystThreshold() + 99} // 124 - 24 = 100
		ctx.Tiers = testTierView(map[sim.ComponentKind]sim.Tier{sim.KindCatalyst: c.tier})
		out, _ := cat.Apply(ctx, sim.Subject{Element: sim.ElementHydrogen, Mass: bignum.FromInt(2)})
		got := out.Mass.Float64()
		if math.Abs(got-c.want) > 1e-9 {
			t.Fatalf("tier %d: got %v want %v", c.tier, got, c.want)
		}
	}
}

// The multiplier must increase monotonically with research above the
// threshold — the whole point of switching off a flat cliff.
func TestCatalystMonotonicAboveThreshold(t *testing.T) {
	c := &Catalyst{}
	levels := []int{catalystThreshold(), catalystThreshold() + 5, 50, 100, 500, 1000}
	prev := 0.0
	for _, r := range levels {
		ctx := sim.NewTestApplyContext()
		ctx.Research = stubResearch{sim.ElementHydrogen: r}
		out, _ := c.Apply(ctx, sim.Subject{Element: sim.ElementHydrogen, Mass: bignum.FromInt(1)})
		got := out.Mass.Float64()
		if got < prev {
			t.Fatalf("research=%d: got %v < previous %v (should be monotonic)", r, got, prev)
		}
		prev = got
	}
}

func TestCatalystReadsElementSpecificResearch(t *testing.T) {
	// Hydrogen subject with Hydrogen research below threshold but Helium above.
	// Catalyst must read Hydrogen research, not Helium, so it's inert.
	c := &Catalyst{}
	ctx := sim.NewTestApplyContext()
	ctx.Research = stubResearch{
		sim.ElementHydrogen: 0,
		sim.ElementHelium:   catalystThreshold() + 100,
	}
	in := sim.Subject{Element: sim.ElementHydrogen, Mass: bignum.FromInt(2)}
	out, _ := c.Apply(ctx, in)
	if !out.Mass.Eq(in.Mass) {
		t.Fatalf("should read Hydrogen research, not Helium: got %v want %v", out.Mass, in.Mass)
	}
}

// stubResearch implements sim.ResearchView for tests.
type stubResearch map[sim.Element]int

func (r stubResearch) Level(e sim.Element) int { return r[e] }
