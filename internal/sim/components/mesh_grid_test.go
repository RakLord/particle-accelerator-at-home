package components

import (
	"encoding/json"
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
)

func TestMeshGridHalvesSpeed(t *testing.T) {
	cases := []struct {
		in, out int
	}{
		{6, 3},
		{5, 2},
		{4, 2},
		{3, 1},
		{2, 1},
		{1, 1}, // band-gated: below meshGridMinSpeed, no-op
		{0, 0},
	}
	mg := &MeshGrid{Orientation: sim.DirEast}
	for _, c := range cases {
		out, lost := mg.Apply(sim.ApplyContext{}, sim.Subject{Speed: c.in, InDirection: sim.DirEast})
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
		Speed:       4,
		Magnetism:   bignum.FromInt(3),
		Direction:   sim.DirWest,
		InDirection: sim.DirEast,
		Position:    sim.Position{X: 1, Y: 2},
	}
	out, lost := mg.Apply(sim.ApplyContext{}, in)
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
	_, lost := mg.Apply(sim.ApplyContext{}, sim.Subject{Speed: 4, InDirection: sim.DirEast})
	if !lost {
		t.Fatal("expected mesh grid to reject side entry")
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
