package components

import (
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

func TestMagnetiserT1AddsBaseBonus(t *testing.T) {
	m := &Magnetiser{}
	ctx := sim.NewTestApplyContext()
	// T1 bonus is +1 (from magnetiserBonusByTier[1]).
	out, lost := m.Apply(ctx, sim.Subject{Speed: 2, Magnetism: bignum.MustParse("0.5")})
	if lost {
		t.Fatal("magnetiser should never destroy subjects")
	}
	if !out.Magnetism.Eq(bignum.MustParse("1.5")) {
		t.Fatalf("T1 bonus: got %v want 1.5", out.Magnetism)
	}
}

func TestMagnetiserTiersScale(t *testing.T) {
	cases := []struct {
		tier    sim.Tier
		wantAdd int
	}{
		{sim.BaseTier, 1},
		{sim.Tier(2), 2},
		{sim.Tier(3), 3},
	}
	for _, c := range cases {
		m := &Magnetiser{}
		ctx := sim.NewTestApplyContext()
		ctx.Tiers = testTierView(map[sim.ComponentKind]sim.Tier{sim.KindMagnetiser: c.tier})
		out, lost := m.Apply(ctx, sim.Subject{Speed: 2, Magnetism: bignum.Zero()})
		if lost {
			t.Fatal("magnetiser should never destroy subjects")
		}
		if !out.Magnetism.Eq(bignum.FromInt(c.wantAdd)) {
			t.Fatalf("tier %d: got %v want %d", c.tier, out.Magnetism, c.wantAdd)
		}
	}
}

func TestMagnetiserAppliesGlobalMultiplier(t *testing.T) {
	m := &Magnetiser{}
	ctx := sim.NewTestApplyContext()
	ctx.Modifiers = sim.GlobalModifiers{MagnetiserBonusMul: bignum.MustParse("1.5")}.Normalized()
	// T1 bonus (1) × modifier (1.5) = 1.5, added to starting 0.
	out, lost := m.Apply(ctx, sim.Subject{Speed: 2, Magnetism: bignum.Zero()})
	if lost {
		t.Fatal("magnetiser should never destroy subjects")
	}
	if !out.Magnetism.Eq(bignum.MustParse("1.5")) {
		t.Fatalf("MagnetiserBonusMul 1.5× on T1: got %v want 1.5", out.Magnetism)
	}
}

func TestMagnetiserBandGate(t *testing.T) {
	m := &Magnetiser{}
	// Speed=0 is below the band — inert.
	out, lost := m.Apply(sim.NewTestApplyContext(), sim.Subject{Speed: 0, Magnetism: bignum.Zero()})
	if lost {
		t.Fatal("magnetiser should never destroy subjects")
	}
	if !out.Magnetism.IsZero() {
		t.Fatalf("band gate failed: got %v want 0", out.Magnetism)
	}
}

// testTierView is a helper so tests can wire a fake Tiers accessor without
// reaching into the unexported tierView type.
type testTierView map[sim.ComponentKind]sim.Tier

func (v testTierView) For(kind sim.ComponentKind) sim.Tier {
	if t, ok := v[kind]; ok && t >= sim.BaseTier {
		return t
	}
	return sim.BaseTier
}
