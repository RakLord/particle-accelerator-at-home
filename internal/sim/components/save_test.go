package components_test

import (
	"encoding/json"
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
)

func TestSaveLoadRoundTripDesktop(t *testing.T) {
	// Isolate the save directory used by internal/save/file_desktop.go.
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	s := sim.NewGameState()
	s.USD = bignum.FromInt(42)
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
	if !got.USD.Eq(bignum.FromInt(42)) {
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
		{Component: &components.SimpleAccelerator{Orientation: sim.DirWest}},
		{Component: &components.Rotator{Orientation: sim.DirSouth}},
		{Component: &components.Pipe{Orientation: sim.DirEast}},
		{Component: &components.MeshGrid{Orientation: sim.DirWest}},
		{Component: &components.Magnetiser{}},
		{Component: &components.Compressor{}},
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
		switch want := c.Component.(type) {
		case *components.SimpleAccelerator:
			gotAcc, ok := got.Component.(*components.SimpleAccelerator)
			if !ok || gotAcc.Orientation != want.Orientation {
				t.Fatalf("cell %d accelerator mismatch: got %#v want %#v", i, got.Component, want)
			}
		case *components.Rotator:
			gotElbow, ok := got.Component.(*components.Rotator)
			if !ok || gotElbow.Orientation != want.Orientation {
				t.Fatalf("cell %d elbow mismatch: got %#v want %#v", i, got.Component, want)
			}
		case *components.MeshGrid:
			gotMesh, ok := got.Component.(*components.MeshGrid)
			if !ok || gotMesh.Orientation != want.Orientation {
				t.Fatalf("cell %d mesh grid mismatch: got %#v want %#v", i, got.Component, want)
			}
		}
	}
}

func TestGameStateRoundTrip(t *testing.T) {
	s := sim.NewGameState()
	s.USD = bignum.MustParse("1234.5")
	s.Research[sim.ElementHydrogen] = 7
	s.UnlockedElements[sim.ElementHelium] = true
	s.InjectionElement = sim.ElementHelium
	s.Grid.Cells[0][0].Component = &components.Injector{
		Direction: sim.DirEast, SpawnInterval: 30, Element: sim.ElementHydrogen, TickCounter: 12,
	}
	s.Grid.Cells[2][2].Component = &components.SimpleAccelerator{Orientation: sim.DirNorth}
	s.Grid.Cells[3][3].Component = &components.MeshGrid{Orientation: sim.DirSouth}
	s.Grid.Cells[4][4].IsCollector = true
	s.Grid.Subjects = append(s.Grid.Subjects, sim.Subject{
		Element: sim.ElementHydrogen, Mass: bignum.One(), Speed: sim.SpeedFromInt(2), Direction: sim.DirEast,
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

	if !loaded.USD.Eq(s.USD) {
		t.Fatalf("USD mismatch: got %v want %v", loaded.USD, s.USD)
	}
	if loaded.Research[sim.ElementHydrogen] != 7 {
		t.Fatalf("research mismatch: %d", loaded.Research[sim.ElementHydrogen])
	}
	if loaded.InjectionElement != sim.ElementHelium {
		t.Fatalf("InjectionElement mismatch: got %q", loaded.InjectionElement)
	}
	if loaded.Grid.Cells[0][0].Component.Kind() != sim.KindInjector {
		t.Fatalf("injector kind lost")
	}
	if inj, ok := loaded.Grid.Cells[0][0].Component.(*components.Injector); !ok || inj.TickCounter != 12 {
		t.Fatalf("injector TickCounter lost")
	}
	if acc, ok := loaded.Grid.Cells[2][2].Component.(*components.SimpleAccelerator); !ok || acc.Orientation != sim.DirNorth {
		t.Fatalf("accelerator orientation lost: %#v", loaded.Grid.Cells[2][2].Component)
	}
	if mesh, ok := loaded.Grid.Cells[3][3].Component.(*components.MeshGrid); !ok || mesh.Orientation != sim.DirSouth {
		t.Fatalf("mesh orientation lost: %#v", loaded.Grid.Cells[3][3].Component)
	}
	if !loaded.Grid.Cells[4][4].IsCollector {
		t.Fatalf("collector flag lost")
	}
	if len(loaded.Grid.Subjects) != 1 || loaded.Grid.Subjects[0].Position != (sim.Position{X: 1, Y: 0}) {
		t.Fatalf("subject lost or malformed")
	}
}

func TestLegacySpeedBonusAndBonusFieldsIgnoredOnLoad(t *testing.T) {
	// Saves from before Phase 3 carried per-instance SpeedBonus/Bonus fields on
	// Accelerator/Magnetiser. Tier drives these now; the legacy fields must
	// unmarshal harmlessly (Go's encoding/json drops unknown fields by default).
	legacy := `{
		"component": {"speed_bonus": 2, "orientation": 0},
		"kind": "accelerator"
	}`
	var c sim.Cell
	if err := json.Unmarshal([]byte(legacy), &c); err != nil {
		t.Fatalf("unmarshal legacy accelerator: %v", err)
	}
	if _, ok := c.Component.(*components.SimpleAccelerator); !ok {
		t.Fatalf("legacy accelerator not reconstructed: %#v", c.Component)
	}

	legacyMag := `{
		"component": {"bonus": "1.5"},
		"kind": "magnetiser"
	}`
	var m sim.Cell
	if err := json.Unmarshal([]byte(legacyMag), &m); err != nil {
		t.Fatalf("unmarshal legacy magnetiser: %v", err)
	}
	if _, ok := m.Component.(*components.Magnetiser); !ok {
		t.Fatalf("legacy magnetiser not reconstructed: %#v", m.Component)
	}
}

func TestLoadV2SeedsOwnedFromGridComponents(t *testing.T) {
	// Migration path: a save that predates the Owned field must be
	// grandfathered in by seeding Owned from the grid contents. This test
	// exercises the Component.Kind() arm of the migration (the sim package
	// covers the IsCollector arm on its own).
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	s := sim.NewGameState()
	s.Owned = nil // Force the migration path on reload.
	s.UnlockedElements[sim.ElementHelium] = true
	s.InjectionElement = "" // Force the legacy per-Injector Element migration path.
	s.Grid.Cells[0][0].Component = &components.Injector{
		Direction: sim.DirEast, SpawnInterval: 30, Element: sim.ElementHelium,
	}
	s.Grid.Cells[1][0].Component = &components.SimpleAccelerator{Orientation: sim.DirNorth}
	s.Grid.Cells[2][0].Component = &components.SimpleAccelerator{Orientation: sim.DirEast}
	s.Grid.Cells[3][0].Component = &components.Magnetiser{}
	s.Grid.Cells[4][0].IsCollector = true

	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, ok, err := sim.Load()
	if err != nil || !ok {
		t.Fatalf("Load: ok=%v err=%v", ok, err)
	}
	expected := map[sim.ComponentKind]int{
		sim.KindInjector:    1,
		sim.KindAccelerator: 2,
		sim.KindMagnetiser:  1,
		sim.KindCollector:   1,
	}
	for kind, want := range expected {
		if got := loaded.Owned[kind]; got != want {
			t.Errorf("Owned[%s] after migration: got %d want %d", kind, got, want)
		}
	}
	if loaded.Owned[sim.KindRotator] != 0 {
		t.Errorf("Owned[rotator]: got %d want 0", loaded.Owned[sim.KindRotator])
	}
	if loaded.InjectionElement != sim.ElementHelium {
		t.Errorf("InjectionElement after migration: got %q want %q", loaded.InjectionElement, sim.ElementHelium)
	}
}
