package components_test

import (
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
)

func TestAcceleratorAddsTierBonusToSpeed(t *testing.T) {
	cases := []struct {
		tier    sim.Tier
		wantAdd int
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
			Speed:       sim.SpeedFromInt(sim.SpeedDivisor), // crosses exactly one cell, arriving at (2,2)
			Direction:   sim.DirEast,
			InDirection: sim.DirEast,
			Position:    sim.Position{X: 1, Y: 2},
			Load:        1,
		})
		s.CurrentLoad = 1
		s.Tick()
		want := sim.SpeedFromInt(sim.SpeedDivisor + c.wantAdd)
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
		Speed:       sim.SpeedFromInt(sim.SpeedDivisor),
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
		Speed:       sim.SpeedFromInt(sim.SpeedDivisor),
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

func TestPipePassesSubjectStraightThrough(t *testing.T) {
	cases := []struct {
		name        string
		orientation sim.Direction
		dir         sim.Direction
	}{
		{"horizontal pipe, east-bound subject", sim.DirEast, sim.DirEast},
		{"horizontal pipe, west-bound subject", sim.DirEast, sim.DirWest},
		{"vertical pipe, north-bound subject", sim.DirNorth, sim.DirNorth},
		{"vertical pipe, south-bound subject", sim.DirNorth, sim.DirSouth},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := sim.NewGameState()
			s.Grid.Cells[2][2].Component = &components.Pipe{Orientation: c.orientation}
			entryPos := sim.Position{X: 2, Y: 2}
			switch c.dir {
			case sim.DirEast:
				entryPos = sim.Position{X: 1, Y: 2}
			case sim.DirWest:
				entryPos = sim.Position{X: 3, Y: 2}
			case sim.DirNorth:
				entryPos = sim.Position{X: 2, Y: 3}
			case sim.DirSouth:
				entryPos = sim.Position{X: 2, Y: 1}
			}
			s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
				Element:     sim.ElementHydrogen,
				Mass:        bignum.One(),
				Speed:       sim.SpeedFromInt(sim.SpeedDivisor),
				Direction:   c.dir,
				InDirection: c.dir,
				Position:    entryPos,
				Load:        1,
			})
			s.CurrentLoad = 1
			s.Tick()
			if len(s.Grid.Subjects) != 1 {
				t.Fatalf("subject destroyed by straight pipe: got %d subjects", len(s.Grid.Subjects))
			}
			if got := s.Grid.Subjects[0].Direction; got != c.dir {
				t.Fatalf("direction changed by straight pipe: got %v want %v", got, c.dir)
			}
		})
	}
}

func TestPipeRejectsPerpendicularEntry(t *testing.T) {
	s := sim.NewGameState()
	// Horizontal pipe at (2,2); subject approaches from above moving south.
	s.Grid.Cells[2][2].Component = &components.Pipe{Orientation: sim.DirEast}
	s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
		Element:     sim.ElementHydrogen,
		Mass:        bignum.One(),
		Speed:       sim.SpeedFromInt(sim.SpeedDivisor),
		Direction:   sim.DirSouth,
		InDirection: sim.DirSouth,
		Position:    sim.Position{X: 2, Y: 1},
		Load:        1,
	})
	s.CurrentLoad = 1
	s.Tick()
	if len(s.Grid.Subjects) != 0 {
		t.Fatalf("expected perpendicular subject destroyed, got %d subjects", len(s.Grid.Subjects))
	}
	if s.CurrentLoad != 0 {
		t.Fatalf("expected CurrentLoad 0 after destruction, got %d", s.CurrentLoad)
	}
}

func TestElbowRejectsDisconnectedEntry(t *testing.T) {
	s := sim.NewGameState()
	s.Grid.Cells[2][2].Component = &components.Rotator{Orientation: sim.DirNorth}
	s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
		Element:     sim.ElementHydrogen,
		Mass:        bignum.One(),
		Speed:       sim.SpeedFromInt(sim.SpeedDivisor),
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
	for x := 0; x < 3; x++ {
		s.Grid.Cells[0][x].Component = &components.Injector{Direction: sim.DirEast}
	}

	if got := s.Inject(); got != 2 {
		t.Fatalf("Inject admitted %d Subjects, want 2", got)
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
	for x := 0; x < 5; x++ {
		s.Grid.Cells[0][x].Component = &components.Injector{Direction: sim.DirEast}
	}

	if got := s.Inject(); got != 5 {
		t.Fatalf("Inject admitted %d Subjects, want 5", got)
	}
	if s.CurrentLoad > s.EffectiveMaxLoad() {
		t.Fatalf("CurrentLoad %d exceeds EffectiveMaxLoad %d", s.CurrentLoad, s.EffectiveMaxLoad())
	}
	if s.CurrentLoad <= s.MaxLoad {
		t.Fatalf("CurrentLoad %d should exceed base MaxLoad %d when bonus is active", s.CurrentLoad, s.MaxLoad)
	}
}

func TestInjectorUsesGlobalInjectionElement(t *testing.T) {
	s := sim.NewGameState()
	s.MaxLoad = 2
	s.UnlockedElements[sim.ElementHelium] = true
	s.InjectionElement = sim.ElementHelium
	s.Grid.Cells[0][0].Component = &components.Injector{
		Direction:     sim.DirEast,
		SpawnInterval: 1,
		Element:       sim.ElementHydrogen, // legacy per-injector value is ignored.
	}

	s.Inject()
	if len(s.Grid.Subjects) != 1 {
		t.Fatalf("expected one spawned Subject, got %d", len(s.Grid.Subjects))
	}
	if got := s.Grid.Subjects[0].Element; got != sim.ElementHelium {
		t.Fatalf("spawned Element = %q, want %q", got, sim.ElementHelium)
	}
	if got, want := s.Grid.Subjects[0].Mass, sim.ElementCatalog[sim.ElementHelium].BaseMass; !got.Eq(want) {
		t.Fatalf("spawned Mass = %v, want %v", got, want)
	}
	if got, want := s.Grid.Subjects[0].Speed, sim.ElementCatalog[sim.ElementHelium].BaseSpeed; got != want {
		t.Fatalf("spawned Speed = %d, want %d", got, want)
	}

	s.InjectionElement = sim.ElementHydrogen
	s.InjectionCooldownRemaining = 0
	s.Inject()
	if len(s.Grid.Subjects) != 2 {
		t.Fatalf("expected second spawned Subject, got %d", len(s.Grid.Subjects))
	}
	if got := s.Grid.Subjects[1].Element; got != sim.ElementHydrogen {
		t.Fatalf("spawned Element after selection change = %q, want %q", got, sim.ElementHydrogen)
	}
}

func TestInjectorUsesElementBaseSpawnStats(t *testing.T) {
	inj := &components.Injector{Direction: sim.DirEast}

	hydrogen, ok := inj.Spawn(sim.NewTestApplyContext(), sim.Position{})
	if !ok {
		t.Fatalf("expected Hydrogen spawn")
	}
	if got, want := hydrogen.Mass, sim.ElementCatalog[sim.ElementHydrogen].BaseMass; !got.Eq(want) {
		t.Fatalf("Hydrogen Mass = %v, want %v", got, want)
	}
	if got, want := hydrogen.Speed, sim.ElementCatalog[sim.ElementHydrogen].BaseSpeed; got != want {
		t.Fatalf("Hydrogen Speed = %d, want %d", got, want)
	}

	ctx := sim.NewTestApplyContext()
	ctx.InjectionElement = sim.ElementCalcium
	calcium, ok := inj.Spawn(ctx, sim.Position{})
	if !ok {
		t.Fatalf("expected Calcium spawn")
	}
	if got, want := calcium.Mass, sim.ElementCatalog[sim.ElementCalcium].BaseMass; !got.Eq(want) {
		t.Fatalf("Calcium Mass = %v, want %v", got, want)
	}
	if got, want := calcium.Speed, sim.ElementCatalog[sim.ElementCalcium].BaseSpeed; got != want {
		t.Fatalf("Calcium Speed = %d, want %d", got, want)
	}
	if !calcium.Mass.GT(hydrogen.Mass) || calcium.Speed >= hydrogen.Speed {
		t.Fatalf("expected Calcium to be heavier and slower than Hydrogen: H=%v/%d Ca=%v/%d", hydrogen.Mass, hydrogen.Speed, calcium.Mass, calcium.Speed)
	}
}

func TestTickDoesNotAutoInject(t *testing.T) {
	s := sim.NewGameState()
	s.Grid.Cells[0][0].Component = &components.Injector{Direction: sim.DirEast}
	for range 20 {
		s.Tick()
	}
	if len(s.Grid.Subjects) != 0 {
		t.Fatalf("Tick auto-injected %d Subjects", len(s.Grid.Subjects))
	}
}

func TestInjectStartsAndRespectsCooldown(t *testing.T) {
	s := sim.NewGameState()
	s.Grid.Cells[0][0].Component = &components.Injector{Direction: sim.DirEast}

	if got := s.Inject(); got != 1 {
		t.Fatalf("first Inject admitted %d Subjects, want 1", got)
	}
	wantCooldown := s.EffectiveInjectionCooldownTicks()
	if s.InjectionCooldownRemaining != wantCooldown {
		t.Fatalf("cooldown = %d, want %d", s.InjectionCooldownRemaining, wantCooldown)
	}
	if got := s.Inject(); got != 0 {
		t.Fatalf("second Inject during cooldown admitted %d Subjects, want 0", got)
	}
	for range wantCooldown {
		s.Tick()
	}
	if s.InjectionCooldownRemaining != 0 {
		t.Fatalf("cooldown after waiting = %d, want 0", s.InjectionCooldownRemaining)
	}
}

func TestInjectDoesNotStartCooldownWhenBlockedByLoad(t *testing.T) {
	s := sim.NewGameState()
	s.MaxLoad = 0
	s.Grid.Cells[0][0].Component = &components.Injector{Direction: sim.DirEast}

	if got := s.Inject(); got != 0 {
		t.Fatalf("Inject admitted %d Subjects under MaxLoad 0, want 0", got)
	}
	if s.InjectionCooldownRemaining != 0 {
		t.Fatalf("blocked Inject started cooldown %d", s.InjectionCooldownRemaining)
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
	s.Grid.Cells[0][1].CollectorDirection = sim.DirEast
	s.Inject()
	for range sim.SpeedDivisor {
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
		Speed:       sim.SpeedFromInt(sim.SpeedDivisor),
		Direction:   sim.DirEast,
		InDirection: sim.DirEast,
		Position:    sim.Position{X: 1, Y: 2},
		Load:        1,
	})
	s.CurrentLoad = 1
	s.Tick()
	// T2 tier bonus (2), modifier adds 5, starting speed was SpeedDivisor.
	want := sim.SpeedFromInt(sim.SpeedDivisor + 2 + 5)
	if got := s.Grid.Subjects[0].Speed; got != want {
		t.Fatalf("accelerator with modifier bonus: got %d want %d", got, want)
	}
}
