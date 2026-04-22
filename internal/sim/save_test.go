package sim

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTripDesktop(t *testing.T) {
	// Isolate the save directory used by internal/save/file_desktop.go
	// (via os.UserConfigDir which respects XDG_CONFIG_HOME).
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir) // fallback on some platforms

	s := NewGameState()
	s.USD = 42
	s.Grid.Cells[0][0].Component = &Injector{Direction: DirEast, SpawnInterval: 5, Element: ElementHydrogen}
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, ok, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !ok {
		t.Fatalf("expected saved state to be present")
	}
	if got.USD != 42 {
		t.Fatalf("USD mismatch: got %v", got.USD)
	}
	if _, isInjector := got.Grid.Cells[0][0].Component.(*Injector); !isInjector {
		t.Fatalf("injector lost on round-trip")
	}
}

func TestSavePropagatesWriteErrors(t *testing.T) {
	// Point the save dir at a non-writable path so WriteFile fails.
	// Using a file-as-directory trick that works on Linux: create a regular
	// file and use it as XDG_CONFIG_HOME, so MkdirAll under it fails.
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", blocker)
	t.Setenv("HOME", blocker)

	s := NewGameState()
	err := s.Save()
	if err == nil {
		t.Fatalf("expected save to propagate a write error, got nil")
	}
}

func TestCellRoundTrip(t *testing.T) {
	cells := []Cell{
		{},
		{IsCollector: true},
		{Component: &Injector{Direction: DirSouth, SpawnInterval: 20, Element: ElementHydrogen, TickCounter: 5}},
		{Component: &SimpleAccelerator{SpeedBonus: 3}},
		{Component: &Rotator{Turn: TurnLeft}},
	}
	for i, c := range cells {
		blob, err := json.Marshal(c)
		if err != nil {
			t.Fatalf("cell %d marshal: %v", i, err)
		}
		var got Cell
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
	s := NewGameState()
	s.USD = 1234.5
	s.Research[ElementHydrogen] = 7
	s.Grid.Cells[0][0].Component = &Injector{Direction: DirEast, SpawnInterval: 30, Element: ElementHydrogen, TickCounter: 12}
	s.Grid.Cells[2][2].Component = &SimpleAccelerator{SpeedBonus: 1}
	s.Grid.Cells[4][4].IsCollector = true
	s.Grid.Subjects = append(s.Grid.Subjects, Subject{
		Element: ElementHydrogen, Mass: 1, Speed: 2, Direction: DirEast,
		Position: Position{X: 1, Y: 0}, Load: 1,
	})
	s.CurrentLoad = 1

	blob, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	loaded := NewGameState()
	if err := json.Unmarshal(blob, loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if loaded.USD != s.USD {
		t.Fatalf("USD mismatch: got %v want %v", loaded.USD, s.USD)
	}
	if loaded.Research[ElementHydrogen] != 7 {
		t.Fatalf("research mismatch: %d", loaded.Research[ElementHydrogen])
	}
	if loaded.Grid.Cells[0][0].Component.Kind() != KindInjector {
		t.Fatalf("injector kind lost")
	}
	if inj, ok := loaded.Grid.Cells[0][0].Component.(*Injector); !ok || inj.TickCounter != 12 {
		t.Fatalf("injector TickCounter lost")
	}
	if !loaded.Grid.Cells[4][4].IsCollector {
		t.Fatalf("collector flag lost")
	}
	if len(loaded.Grid.Subjects) != 1 || loaded.Grid.Subjects[0].Position != (Position{X: 1, Y: 0}) {
		t.Fatalf("subject lost or malformed")
	}
}
