package components_test

import (
	"encoding/json"
	"testing"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
)

func TestSaveLoadRoundTripDesktop(t *testing.T) {
	// Isolate the save directory used by internal/save/file_desktop.go.
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	s := sim.NewGameState()
	s.USD = 42
	s.Grid.Cells[0][0].Component = &components.Injector{
		Direction: sim.DirEast, SpawnInterval: 5, Element: sim.ElementHydrogen,
	}
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, ok, err := sim.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !ok {
		t.Fatalf("expected saved state to be present")
	}
	if got.USD != 42 {
		t.Fatalf("USD mismatch: got %v", got.USD)
	}
	if _, isInjector := got.Grid.Cells[0][0].Component.(*components.Injector); !isInjector {
		t.Fatalf("injector lost on round-trip")
	}
}

func TestCellRoundTrip(t *testing.T) {
	cells := []sim.Cell{
		{},
		{IsCollector: true},
		{Component: &components.Injector{Direction: sim.DirSouth, SpawnInterval: 20, Element: sim.ElementHydrogen, TickCounter: 5}},
		{Component: &components.SimpleAccelerator{SpeedBonus: 3}},
		{Component: &components.Rotator{Turn: components.TurnLeft}},
		{Component: &components.MeshGrid{}},
		{Component: &components.Magnetiser{Bonus: 1.5}},
	}
	for i, c := range cells {
		blob, err := json.Marshal(c)
		if err != nil {
			t.Fatalf("cell %d marshal: %v", i, err)
		}
		var got sim.Cell
		if err := json.Unmarshal(blob, &got); err != nil {
			t.Fatalf("cell %d unmarshal: %v", i, err)
		}
		if got.IsCollector != c.IsCollector {
			t.Fatalf("cell %d IsCollector mismatch", i)
		}
		if (got.Component == nil) != (c.Component == nil) {
			t.Fatalf("cell %d Component nil-ness mismatch", i)
		}
		if c.Component != nil && got.Component.Kind() != c.Component.Kind() {
			t.Fatalf("cell %d kind mismatch: %s vs %s", i, got.Component.Kind(), c.Component.Kind())
		}
	}
}

func TestGameStateRoundTrip(t *testing.T) {
	s := sim.NewGameState()
	s.USD = 1234.5
	s.Research[sim.ElementHydrogen] = 7
	s.Grid.Cells[0][0].Component = &components.Injector{
		Direction: sim.DirEast, SpawnInterval: 30, Element: sim.ElementHydrogen, TickCounter: 12,
	}
	s.Grid.Cells[2][2].Component = &components.SimpleAccelerator{SpeedBonus: 1}
	s.Grid.Cells[4][4].IsCollector = true
	s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
		Element: sim.ElementHydrogen, Mass: 1, Speed: 2, Direction: sim.DirEast,
		Position: sim.Position{X: 1, Y: 0}, Load: 1,
	})
	s.CurrentLoad = 1

	blob, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	loaded := sim.NewGameState()
	if err := json.Unmarshal(blob, loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if loaded.USD != s.USD {
		t.Fatalf("USD mismatch: got %v want %v", loaded.USD, s.USD)
	}
	if loaded.Research[sim.ElementHydrogen] != 7 {
		t.Fatalf("research mismatch: %d", loaded.Research[sim.ElementHydrogen])
	}
	if loaded.Grid.Cells[0][0].Component.Kind() != components.KindInjector {
		t.Fatalf("injector kind lost")
	}
	if inj, ok := loaded.Grid.Cells[0][0].Component.(*components.Injector); !ok || inj.TickCounter != 12 {
		t.Fatalf("injector TickCounter lost")
	}
	if !loaded.Grid.Cells[4][4].IsCollector {
		t.Fatalf("collector flag lost")
	}
	if len(loaded.Grid.Subjects) != 1 || loaded.Grid.Subjects[0].Position != (sim.Position{X: 1, Y: 0}) {
		t.Fatalf("subject lost or malformed")
	}
}
