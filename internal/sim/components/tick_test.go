package components_test

import (
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
)

func TestAcceleratorAddsTierBonusToSpeed(t *testing.T) {
	cases := []struct {
		tier     sim.Tier
		wantAdd  int
	}{
		{sim.BaseTier, 1}, // T1 = +1
		{sim.Tier(2), 2},
		{sim.Tier(3), 3},
	}
	for _, c := range cases {
		s := sim.NewGameState()
		if c.tier != sim.BaseTier {
			s.ComponentTiers = map[sim.ComponentKind]sim.Tier{sim.KindAccelerator: c.tier}
		}
		s.Grid.Cells[2][2].Component = &components.SimpleAccelerator{
			Orientation: sim.DirEast,
		}
		s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
			Element:     sim.ElementHydrogen,
			Mass:        bignum.One(),
			Speed:       sim.SpeedDivisor, // crosses exactly one cell, arriving at (2,2)
			Direction:   sim.DirEast,
			InDirection: sim.DirEast,
			Position:    sim.Position{X: 1, Y: 2},
			Load:        1,
		})
		s.CurrentLoad = 1
		s.Tick()
		want := sim.SpeedDivisor + c.wantAdd
		if got := s.Grid.Subjects[0].Speed; got != want {
			t.Fatalf("tier %d: got speed %d want %d", c.tier, got, want)
		}
	}
}

func TestAcceleratorRejectsSideEntry(t *testing.T) {
	s := sim.NewGameState()
	s.Grid.Cells[2][2].Component = &components.SimpleAccelerator{
		Orientation: sim.DirNorth,
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
	if len(s.Grid.Subjects) != 0 {
		t.Fatalf("expected subject destroyed on side entry, got %d subjects", len(s.Grid.Subjects))
	}
	if s.CurrentLoad != 0 {
		t.Fatalf("expected CurrentLoad 0 after destruction, got %d", s.CurrentLoad)
	}
}

func TestElbowChangesDirection(t *testing.T) {
	s := sim.NewGameState()
	s.Grid.Cells[2][2].Component = &components.Rotator{Orientation: sim.DirNorth}
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
	if got := s.Grid.Subjects[0].Direction; got != sim.DirNorth {
		t.Fatalf("expected DirNorth after elbow turn, got %v", got)
	}
}

func TestElbowRejectsDisconnectedEntry(t *testing.T) {
	s := sim.NewGameState()
	s.Grid.Cells[2][2].Component = &components.Rotator{Orientation: sim.DirNorth}
	s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
		Element:     sim.ElementHydrogen,
		Mass:        bignum.One(),
		Speed:       sim.SpeedDivisor,
		Direction:   sim.DirWest,
		InDirection: sim.DirWest,
		Position:    sim.Position{X: 3, Y: 2},
		Load:        1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if len(s.Grid.Subjects) != 0 {
		t.Fatalf("expected subject destroyed on disconnected elbow entry, got %d subjects", len(s.Grid.Subjects))
	}
	if s.CurrentLoad != 0 {
		t.Fatalf("expected CurrentLoad 0 after destruction, got %d", s.CurrentLoad)
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

func TestInjectorRespectsEffectiveMaxLoadWithBonus(t *testing.T) {
	// Base MaxLoad 2 + MaxLoadBonus 3 = effective 5. Injector should fill to 5.
	s := sim.NewGameState()
	s.MaxLoad = 2
	s.Modifiers.MaxLoadBonus = 3
	s.Grid.Cells[0][0].Component = &components.Injector{
		Direction:     sim.DirEast,
		SpawnInterval: 1,
		Element:       sim.ElementHydrogen,
	}
	for range 20 {
		s.Tick()
	}
	if s.CurrentLoad > s.EffectiveMaxLoad() {
		t.Fatalf("CurrentLoad %d exceeds EffectiveMaxLoad %d", s.CurrentLoad, s.EffectiveMaxLoad())
	}
	if s.CurrentLoad <= s.MaxLoad {
		t.Fatalf("CurrentLoad %d should exceed base MaxLoad %d when bonus is active", s.CurrentLoad, s.MaxLoad)
	}
}

func TestResearchPerCollectBonusAppliesOnCollection(t *testing.T) {
	// Place an Injector feeding directly into a Collector one cell east.
	// Each collection should increment research by 1 + bonus.
	s := sim.NewGameState()
	s.Modifiers.ResearchPerCollectBonus = 2
	s.Grid.Cells[0][0].Component = &components.Injector{
		Direction:     sim.DirEast,
		SpawnInterval: 1,
		Element:       sim.ElementHydrogen,
	}
	s.Grid.Cells[0][1].IsCollector = true
	for range 30 {
		s.Tick()
	}
	if s.Research[sim.ElementHydrogen] < 3 {
		t.Fatalf("expected research to grow by 3 per collection, got %d total", s.Research[sim.ElementHydrogen])
	}
	// With bonus=2 each collection yields +3 research. Must be a multiple of 3.
	if s.Research[sim.ElementHydrogen]%3 != 0 {
		t.Fatalf("research %d should be multiple of 3 with bonus=2", s.Research[sim.ElementHydrogen])
	}
}

func TestAcceleratorSpeedBonusAppliesFromModifiers(t *testing.T) {
	s := sim.NewGameState()
	s.Modifiers.AcceleratorSpeedBonus = 5
	// Tier 2 → +2 Speed. Combined with the +5 modifier the total gain is 7.
	s.ComponentTiers = map[sim.ComponentKind]sim.Tier{sim.KindAccelerator: sim.Tier(2)}
	s.Grid.Cells[2][2].Component = &components.SimpleAccelerator{
		Orientation: sim.DirEast,
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
	// T2 tier bonus (2), modifier adds 5, starting speed was SpeedDivisor.
	want := sim.SpeedDivisor + 2 + 5
	if got := s.Grid.Subjects[0].Speed; got != want {
		t.Fatalf("accelerator with modifier bonus: got %d want %d", got, want)
	}
}
