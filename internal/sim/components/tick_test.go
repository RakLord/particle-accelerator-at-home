package components_test

import (
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
)

func TestAcceleratorIncreasesSpeed(t *testing.T) {
	s := sim.NewGameState()
	s.Grid.Cells[2][2].Component = &components.SimpleAccelerator{SpeedBonus: 2 * sim.SpeedDivisor}
	s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
		Element:     sim.ElementHydrogen,
		Mass:        bignum.One(),
		Speed:       sim.SpeedDivisor, // one cell per tick so the tick enters (2,2)
		Direction:   sim.DirEast,
		InDirection: sim.DirEast,
		Position:    sim.Position{X: 1, Y: 2},
		Load:        1,
	})
	s.CurrentLoad = 1
	s.Tick()
	// Accelerator added 2·SpeedDivisor on top of the starting SpeedDivisor.
	if s.Grid.Subjects[0].Speed != 3*sim.SpeedDivisor {
		t.Fatalf("expected speed %d after accelerator, got %d", 3*sim.SpeedDivisor, s.Grid.Subjects[0].Speed)
	}
}

func TestRotatorChangesDirection(t *testing.T) {
	s := sim.NewGameState()
	s.Grid.Cells[2][2].Component = &components.Rotator{Turn: components.TurnRight}
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
	if got := s.Grid.Subjects[0].Direction; got != sim.DirSouth {
		t.Fatalf("expected DirSouth after right turn, got %v", got)
	}
}

func TestInjectorRespectsMaxLoad(t *testing.T) {
	s := sim.NewGameState()
	s.MaxLoad = 2
	s.Grid.Cells[0][0].Component = &components.Injector{
		Direction:     sim.DirEast,
		SpawnInterval: 1,
		Element:       sim.ElementHydrogen,
	}
	for range 10 {
		s.Tick()
	}
	if s.CurrentLoad > s.MaxLoad {
		t.Fatalf("CurrentLoad %d exceeds MaxLoad %d", s.CurrentLoad, s.MaxLoad)
	}
}
