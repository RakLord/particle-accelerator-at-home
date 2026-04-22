package sim

import "testing"

func TestTickMovesSubject(t *testing.T) {
	s := NewGameState()
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element:   ElementHydrogen,
		Mass:      1,
		Speed:     1,
		Direction: DirEast,
		Position:  Position{X: 1, Y: 2},
		Load:      1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if len(s.Grid.Subjects) != 1 {
		t.Fatalf("expected 1 subject, got %d", len(s.Grid.Subjects))
	}
	got := s.Grid.Subjects[0].Position
	if got != (Position{X: 2, Y: 2}) {
		t.Fatalf("expected position (2,2), got %v", got)
	}
}

func TestAcceleratorIncreasesSpeed(t *testing.T) {
	s := NewGameState()
	s.Grid.Cells[2][2].Component = &SimpleAccelerator{SpeedBonus: 2}
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element:   ElementHydrogen,
		Mass:      1,
		Speed:     1,
		Direction: DirEast,
		Position:  Position{X: 1, Y: 2},
		Load:      1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if s.Grid.Subjects[0].Speed != 3 {
		t.Fatalf("expected speed 3 after accelerator, got %d", s.Grid.Subjects[0].Speed)
	}
}

func TestRotatorChangesDirection(t *testing.T) {
	s := NewGameState()
	s.Grid.Cells[2][2].Component = &Rotator{Turn: TurnRight}
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element:   ElementHydrogen,
		Mass:      1,
		Speed:     1,
		Direction: DirEast,
		Position:  Position{X: 1, Y: 2},
		Load:      1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if got := s.Grid.Subjects[0].Direction; got != DirSouth {
		t.Fatalf("expected DirSouth after right turn, got %v", got)
	}
}

func TestCollectorAwardsUSD(t *testing.T) {
	s := NewGameState()
	s.Grid.Cells[2][2].IsCollector = true
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element:   ElementHydrogen,
		Mass:      2,
		Speed:     1,
		Direction: DirEast,
		Position:  Position{X: 1, Y: 2},
		Load:      1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if len(s.Grid.Subjects) != 0 {
		t.Fatalf("expected subject removed after collection, got %d", len(s.Grid.Subjects))
	}
	if s.USD != 2 {
		t.Fatalf("expected USD 2, got %v", s.USD)
	}
	if s.Research[ElementHydrogen] != 1 {
		t.Fatalf("expected research 1, got %d", s.Research[ElementHydrogen])
	}
	if s.CurrentLoad != 0 {
		t.Fatalf("expected CurrentLoad 0, got %d", s.CurrentLoad)
	}
}

func TestInjectorRespectsMaxLoad(t *testing.T) {
	s := NewGameState()
	s.MaxLoad = 2
	s.Grid.Cells[0][0].Component = &Injector{
		Direction:     DirEast,
		SpawnInterval: 1,
		Element:       ElementHydrogen,
	}
	for range 10 {
		s.Tick()
	}
	if s.CurrentLoad > s.MaxLoad {
		t.Fatalf("CurrentLoad %d exceeds MaxLoad %d", s.CurrentLoad, s.MaxLoad)
	}
}

func TestSubjectOffGridIsRemoved(t *testing.T) {
	s := NewGameState()
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element:   ElementHydrogen,
		Speed:     1,
		Direction: DirEast,
		Position:  Position{X: GridSize - 1, Y: 0},
		Load:      1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if len(s.Grid.Subjects) != 0 {
		t.Fatalf("expected subject removed after falling off grid")
	}
	if s.CurrentLoad != 0 {
		t.Fatalf("expected CurrentLoad 0, got %d", s.CurrentLoad)
	}
}
