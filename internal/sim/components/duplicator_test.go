package components_test

import (
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
)

// Feeding a subject into the west side of a Duplicator oriented West should
// emit two Subjects heading North and South respectively.
func TestDuplicatorT1EmitsTwoPerpendicularOutputs(t *testing.T) {
	s := sim.NewGameState()
	s.MaxLoad = 16
	s.Grid.Cells[2][2].Component = &components.Duplicator{Orientation: sim.DirWest}
	s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
		Element:     sim.ElementHydrogen,
		Mass:        bignum.FromInt(4),
		Speed:       sim.SpeedFromInt(sim.SpeedDivisor),
		Magnetism:   bignum.FromInt(2),
		Direction:   sim.DirEast,
		InDirection: sim.DirEast,
		Position:    sim.Position{X: 1, Y: 2},
		Load:        1,
	})
	s.CurrentLoad = 1
	s.Tick()

	if len(s.Grid.Subjects) != 2 {
		t.Fatalf("expected 2 Subjects after duplicator, got %d", len(s.Grid.Subjects))
	}

	// T1 halves Mass on each output. Magnetism and Speed copied unchanged.
	dirs := map[sim.Direction]bool{}
	for _, sub := range s.Grid.Subjects {
		if !sub.Mass.Eq(bignum.FromInt(2)) {
			t.Errorf("T1 output Mass: got %v want 2", sub.Mass)
		}
		if sub.Speed != sim.SpeedFromInt(sim.SpeedDivisor) {
			t.Errorf("output Speed: got %d want %d", sub.Speed, sim.SpeedFromInt(sim.SpeedDivisor))
		}
		if !sub.Magnetism.Eq(bignum.FromInt(2)) {
			t.Errorf("output Magnetism: got %v want 2", sub.Magnetism)
		}
		dirs[sub.Direction] = true
	}
	// West-oriented Duplicator's perpendicular outputs are North and South.
	if !dirs[sim.DirNorth] || !dirs[sim.DirSouth] {
		t.Fatalf("expected outputs N and S; got dir set %v", dirs)
	}
}

func TestDuplicatorTiersScaleMassFraction(t *testing.T) {
	// Incoming Mass 10 × per-output fraction yields per-output masses.
	cases := []struct {
		tier       sim.Tier
		wantPerOut string
	}{
		{sim.BaseTier, "5"},  // 10 × 0.5
		{sim.Tier(2), "6"},   // 10 × 0.6
		{sim.Tier(3), "7.5"}, // 10 × 0.75
	}
	for _, c := range cases {
		s := sim.NewGameState()
		s.MaxLoad = 16
		s.ComponentTiers = map[sim.ComponentKind]sim.Tier{sim.KindDuplicator: c.tier}
		s.Grid.Cells[2][2].Component = &components.Duplicator{Orientation: sim.DirWest}
		s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
			Element:     sim.ElementHydrogen,
			Mass:        bignum.FromInt(10),
			Speed:       sim.SpeedFromInt(sim.SpeedDivisor),
			Direction:   sim.DirEast,
			InDirection: sim.DirEast,
			Position:    sim.Position{X: 1, Y: 2},
			Load:        1,
		})
		s.CurrentLoad = 1
		s.Tick()
		if len(s.Grid.Subjects) != 2 {
			t.Fatalf("tier %d: expected 2 outputs, got %d", c.tier, len(s.Grid.Subjects))
		}
		for _, sub := range s.Grid.Subjects {
			if !sub.Mass.Eq(bignum.MustParse(c.wantPerOut)) {
				t.Errorf("tier %d: output Mass %v want %s", c.tier, sub.Mass, c.wantPerOut)
			}
		}
	}
}

// Entry from a closed side must be destroyed — not emit outputs.
func TestDuplicatorRejectsWrongSideEntry(t *testing.T) {
	s := sim.NewGameState()
	s.MaxLoad = 16
	// Duplicator expects input from WEST. Send a subject from the SOUTH side.
	s.Grid.Cells[2][2].Component = &components.Duplicator{Orientation: sim.DirWest}
	s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
		Element:     sim.ElementHydrogen,
		Mass:        bignum.One(),
		Speed:       sim.SpeedFromInt(sim.SpeedDivisor),
		Direction:   sim.DirNorth,
		InDirection: sim.DirNorth,
		Position:    sim.Position{X: 2, Y: 3},
		Load:        1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if len(s.Grid.Subjects) != 0 {
		t.Fatalf("expected subject destroyed on wrong-side entry, got %d", len(s.Grid.Subjects))
	}
	if s.CurrentLoad != 0 {
		t.Fatalf("expected CurrentLoad 0 after destruction, got %d", s.CurrentLoad)
	}
}

// A full grid should only admit as many extras as MaxLoad allows. The input
// Subject's Load is freed when it's consumed, so a MaxLoad=1 grid after the
// Duplicator runs can admit exactly one extra — the second is dropped.
func TestDuplicatorRespectsEffectiveMaxLoad(t *testing.T) {
	s := sim.NewGameState()
	s.MaxLoad = 1
	s.Grid.Cells[2][2].Component = &components.Duplicator{Orientation: sim.DirWest}
	s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
		Element:     sim.ElementHydrogen,
		Mass:        bignum.FromInt(4),
		Speed:       sim.SpeedFromInt(sim.SpeedDivisor),
		Direction:   sim.DirEast,
		InDirection: sim.DirEast,
		Position:    sim.Position{X: 1, Y: 2},
		Load:        1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if len(s.Grid.Subjects) != 1 {
		t.Fatalf("expected 1 emitted Subject under MaxLoad=1, got %d", len(s.Grid.Subjects))
	}
	if s.CurrentLoad != 1 {
		t.Fatalf("CurrentLoad: got %d want 1", s.CurrentLoad)
	}
}
