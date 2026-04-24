package components

import (
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

func TestCatalystActivatesAtThreshold(t *testing.T) {
	c := &Catalyst{}
	ctx := sim.NewTestApplyContext()
	ctx.Research = stubResearch{sim.ElementHydrogen: catalystThreshold()}
	out, _ := c.Apply(ctx, sim.Subject{Element: sim.ElementHydrogen, Mass: bignum.FromInt(2)})
	// T1 multiplier is ×1.5 → 2 × 1.5 = 3.
	if !out.Mass.Eq(bignum.FromInt(3)) {
		t.Fatalf("T1 at threshold: got %v want 3", out.Mass)
	}
}

func TestCatalystTiersScale(t *testing.T) {
	cases := []struct {
		tier sim.Tier
		want string
	}{
		{sim.BaseTier, "3"},    // 2 × 1.5
		{sim.Tier(2), "4"},     // 2 × 2
		{sim.Tier(3), "6"},     // 2 × 3
	}
	for _, c := range cases {
		cat := &Catalyst{}
		ctx := sim.NewTestApplyContext()
		ctx.Research = stubResearch{sim.ElementHydrogen: catalystThreshold()}
		ctx.Tiers = testTierView(map[sim.ComponentKind]sim.Tier{sim.KindCatalyst: c.tier})
		out, _ := cat.Apply(ctx, sim.Subject{Element: sim.ElementHydrogen, Mass: bignum.FromInt(2)})
		if !out.Mass.Eq(bignum.MustParse(c.want)) {
			t.Fatalf("tier %d: got %v want %s", c.tier, out.Mass, c.want)
		}
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
