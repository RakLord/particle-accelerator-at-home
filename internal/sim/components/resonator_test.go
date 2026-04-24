package components_test

import (
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
)

func TestResonatorIsolatedIsInert(t *testing.T) {
	s := sim.NewGameState()
	s.Grid.Cells[2][2].Component = &components.Resonator{}
	s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
		Element:     sim.ElementHydrogen,
		Mass:        bignum.One(),
		Speed:       sim.SpeedDivisor,
		Direction:   sim.DirEast,
		InDirection: sim.DirEast,
		Position:    sim.Position{X: 1, Y: 2},
		Load:        1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if got := s.Grid.Subjects[0].Speed; got != sim.SpeedDivisor {
		t.Fatalf("isolated Resonator should be inert: got speed %d want %d", got, sim.SpeedDivisor)
	}
}

func TestResonatorCountsOrthogonalNeighbours(t *testing.T) {
	cases := []struct {
		placeN, placeE, placeS bool
		wantAdd                int
	}{
		{false, false, false, 0},
		{true, false, false, 1},
		{true, true, false, 2},
		{true, true, true, 3},
	}
	for _, c := range cases {
		s := sim.NewGameState()
		s.Grid.Cells[2][2].Component = &components.Resonator{}
		if c.placeN {
			s.Grid.Cells[1][2].Component = &components.Resonator{}
		}
		if c.placeE {
			s.Grid.Cells[2][3].Component = &components.Resonator{}
		}
		if c.placeS {
			s.Grid.Cells[3][2].Component = &components.Resonator{}
		}
		s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
			Element:     sim.ElementHydrogen,
			Mass:        bignum.One(),
			Speed:       sim.SpeedDivisor,
			Direction:   sim.DirEast,
			InDirection: sim.DirEast,
			Position:    sim.Position{X: 1, Y: 2},
			Load:        1,
		})
		s.CurrentLoad = 1
		s.Tick()
		// T1 bonus is +1 per neighbour.
		want := sim.SpeedDivisor + c.wantAdd
		if got := s.Grid.Subjects[0].Speed; got != want {
			t.Fatalf("neighbours N=%v E=%v S=%v: got %d want %d", c.placeN, c.placeE, c.placeS, got, want)
		}
	}
}

func TestResonatorTiersScalePerNeighbour(t *testing.T) {
	// Two neighbours, scanning T1→T3. Expected Speed gain = tierBonus × 2.
	for tier := sim.BaseTier; tier <= sim.Tier(3); tier++ {
		s := sim.NewGameState()
		s.ComponentTiers = map[sim.ComponentKind]sim.Tier{sim.KindResonator: tier}
		s.Grid.Cells[2][2].Component = &components.Resonator{}
		s.Grid.Cells[1][2].Component = &components.Resonator{}
		s.Grid.Cells[2][3].Component = &components.Resonator{}
		s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
			Element:     sim.ElementHydrogen,
			Mass:        bignum.One(),
			Speed:       sim.SpeedDivisor,
			Direction:   sim.DirEast,
			InDirection: sim.DirEast,
			Position:    sim.Position{X: 1, Y: 2},
			Load:        1,
		})
		s.CurrentLoad = 1
		s.Tick()
		want := sim.SpeedDivisor + 2*int(tier)
		if got := s.Grid.Subjects[0].Speed; got != want {
			t.Fatalf("tier %d with 2 neighbours: got %d want %d", tier, got, want)
		}
	}
}
