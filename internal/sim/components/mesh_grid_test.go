package components

import (
	"encoding/json"
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

func TestMeshGridHalvesSpeed(t *testing.T) {
	cases := []struct {
		in, out sim.Speed
	}{
		{sim.SpeedFromInt(6), sim.SpeedFromInt(3)},
		{sim.SpeedFromInt(5), sim.SpeedFromRatio(5, 2)},
		{sim.SpeedFromInt(4), sim.SpeedFromInt(2)},
		{sim.SpeedFromInt(3), sim.SpeedFromRatio(3, 2)},
		{sim.SpeedFromInt(2), sim.SpeedFromInt(1)},
		{sim.SpeedFromInt(1), sim.SpeedFromRatio(1, 2)},
		{0, 0},
	}
	mg := &MeshGrid{Orientation: sim.DirEast}
	for _, c := range cases {
		out, lost := mg.Apply(sim.NewTestApplyContext(), sim.Subject{Speed: c.in, InDirection: sim.DirEast})
		if lost {
			t.Fatal("mesh grid should never destroy subjects")
		}
		got := out.Speed
		if got != c.out {
			t.Fatalf("Speed %d: got %d want %d", c.in, got, c.out)
		}
	}
}

func TestMeshGridPreservesOtherFields(t *testing.T) {
	mg := &MeshGrid{Orientation: sim.DirEast}
	in := sim.Subject{
		Element:     sim.ElementHelium,
		Mass:        bignum.FromInt(2),
		Speed:       sim.SpeedFromInt(4),
		Magnetism:   bignum.FromInt(3),
		Direction:   sim.DirWest,
		InDirection: sim.DirEast,
		Position:    sim.Position{X: 1, Y: 2},
	}
	out, lost := mg.Apply(sim.NewTestApplyContext(), in)
	if lost {
		t.Fatal("mesh grid should never destroy subjects")
	}
	if out.Element != in.Element || !out.Mass.Eq(in.Mass) || !out.Magnetism.Eq(in.Magnetism) ||
		out.Direction != in.Direction || out.InDirection != in.InDirection || out.Position != in.Position {
		t.Fatalf("MeshGrid mutated unrelated fields: %+v", out)
	}
}

func TestMeshGridRejectsSideEntry(t *testing.T) {
	mg := &MeshGrid{Orientation: sim.DirNorth}
	_, lost := mg.Apply(sim.NewTestApplyContext(), sim.Subject{Speed: sim.SpeedFromInt(4), InDirection: sim.DirEast})
	if !lost {
		t.Fatal("expected mesh grid to reject side entry")
	}
}

func TestMeshGridTiersChangeDivisor(t *testing.T) {
	cases := []struct {
		tier    sim.Tier
		inSpeed sim.Speed
		want    sim.Speed
	}{
		{sim.BaseTier, sim.SpeedFromInt(6), sim.SpeedFromInt(3)},     // T1 halves
		{sim.Tier(2), sim.SpeedFromInt(6), sim.SpeedFromInt(2)},      // T2 thirds
		{sim.Tier(3), sim.SpeedFromInt(8), sim.SpeedFromInt(2)},      // T3 quarters
		{sim.Tier(2), sim.SpeedFromInt(2), sim.SpeedFromRatio(2, 3)}, // fractional output
		{sim.Tier(3), sim.SpeedFromInt(3), sim.SpeedFromRatio(3, 4)}, // fractional output
		{sim.BaseTier, sim.MinSpeed, sim.MinSpeed},                   // non-zero speeds never become trapped
	}
	for _, c := range cases {
		mg := &MeshGrid{Orientation: sim.DirEast}
		ctx := sim.NewTestApplyContext()
		ctx.Tiers = testTierView(map[sim.ComponentKind]sim.Tier{sim.KindMeshGrid: c.tier})
		out, lost := mg.Apply(ctx, sim.Subject{Speed: c.inSpeed, InDirection: sim.DirEast})
		if lost {
			t.Fatal("mesh grid should never destroy subjects")
		}
		if out.Speed != c.want {
			t.Fatalf("tier %d speed %d: got %d want %d", c.tier, c.inSpeed, out.Speed, c.want)
		}
	}
}

func TestMeshGridLegacyJSONDefaultsHorizontal(t *testing.T) {
	var mg MeshGrid
	if err := json.Unmarshal([]byte(`{}`), &mg); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if mg.Orientation != sim.DirEast {
		t.Fatalf("legacy Orientation = %v, want DirEast", mg.Orientation)
	}
}
