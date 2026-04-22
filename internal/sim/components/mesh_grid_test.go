package components

import (
	"testing"

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
	mg := &MeshGrid{}
	for _, c := range cases {
		got := mg.Apply(sim.Subject{Speed: c.in}).Speed
		if got != c.out {
			t.Fatalf("Speed %d: got %d want %d", c.in, got, c.out)
		}
	}
}

func TestMeshGridPreservesOtherFields(t *testing.T) {
	mg := &MeshGrid{}
	in := sim.Subject{
		Element:   sim.ElementHelium,
		Mass:      2,
		Speed:     4,
		Magnetism: 3,
		Direction: sim.DirWest,
		Position:  sim.Position{X: 1, Y: 2},
	}
	out := mg.Apply(in)
	if out.Element != in.Element || out.Mass != in.Mass || out.Magnetism != in.Magnetism ||
		out.Direction != in.Direction || out.Position != in.Position {
		t.Fatalf("MeshGrid mutated unrelated fields: %+v", out)
	}
}
