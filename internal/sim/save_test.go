package sim

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/save"
)

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

func TestGameStateRoundTripUnlockedElements(t *testing.T) {
	s := NewGameState()
	s.Research[ElementHydrogen] = 15
	s.UnlockedElements[ElementHelium] = true

	blob, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	loaded := NewGameState()
	if err := json.Unmarshal(blob, loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !loaded.UnlockedElements[ElementHelium] {
		t.Fatalf("Helium unlock flag lost on round-trip")
	}
	if !loaded.UnlockedElements[ElementHydrogen] {
		t.Fatalf("Hydrogen unlock flag lost on round-trip")
	}
}

func TestLoadV2SaveDefaultsUnlockedElements(t *testing.T) {
	// Simulate a current-version save payload with no UnlockedElements field.
	// The nil-guard in Load() must default Hydrogen on.
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	legacyState := `{"Grid":{"Cells":[[{},{},{},{},{}],[{},{},{},{},{}],[{},{},{},{},{}],[{},{},{},{},{}],[{},{},{},{},{}]],"Subjects":null},"USD":100,"Research":{"hydrogen":3},"MaxLoad":16,"CurrentLoad":0,"TickRate":10,"Ticks":0}`
	env := `{"version":2,"state":` + legacyState + `}`
	if err := save.Write(saveKey, env); err != nil {
		t.Fatalf("seed current save: %v", err)
	}

	loaded, ok, err := Load()
	if err != nil || !ok {
		t.Fatalf("Load current save: ok=%v err=%v", ok, err)
	}
	if !loaded.UnlockedElements[ElementHydrogen] {
		t.Fatalf("save should default Hydrogen to unlocked")
	}
	if loaded.UnlockedElements[ElementHelium] {
		t.Fatalf("save should not unlock Helium")
	}
	if !loaded.USD.Eq(bignum.FromInt(100)) {
		t.Fatalf("USD mismatch: got %v want 100", loaded.USD)
	}
	if loaded.Research[ElementHydrogen] != 3 {
		t.Fatalf("research mismatch: got %d want 3", loaded.Research[ElementHydrogen])
	}
}

func TestSaveLoadPreservesOwned(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	s := NewGameState()
	s.Owned = map[ComponentKind]int{
		KindInjector:    3,
		KindAccelerator: 7,
		KindCollector:   2,
	}
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, ok, err := Load()
	if err != nil || !ok {
		t.Fatalf("Load: ok=%v err=%v", ok, err)
	}
	for kind, want := range s.Owned {
		if got := loaded.Owned[kind]; got != want {
			t.Errorf("Owned[%s]: got %d want %d", kind, got, want)
		}
	}
}

func TestLoadV2SaveSeedsOwnedCollectorsFromGrid(t *testing.T) {
	// Craft a v2 save with two Collector cells and no `owned` field.
	// Collectors don't go through componentRegistry (they're cell.IsCollector),
	// so this test doesn't need the components package imported. The
	// Component.Kind() path is exercised in the components_test package.
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	row0 := `[{"is_collector":true},{},{"is_collector":true},{},{}]`
	empty := `[{},{},{},{},{}]`
	grid := `{"Cells":[` + row0 + `,` + empty + `,` + empty + `,` + empty + `,` + empty + `],"Subjects":null}`
	state := `{"Layer":"genesis","Grid":` + grid + `,"USD":"0","Research":{},"UnlockedElements":{"hydrogen":true},"MaxLoad":16,"CurrentLoad":0,"TickRate":10,"Ticks":0}`
	env := `{"version":2,"state":` + state + `}`
	if err := save.Write(saveKey, env); err != nil {
		t.Fatalf("seed save: %v", err)
	}

	loaded, ok, err := Load()
	if err != nil || !ok {
		t.Fatalf("Load: ok=%v err=%v", ok, err)
	}
	if loaded.Owned[KindCollector] != 2 {
		t.Errorf("Owned[collector] after migration: got %d want 2", loaded.Owned[KindCollector])
	}
}

func TestLoadRejectsV1Save(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	legacyState := `{"Grid":{"Cells":[[{},{},{},{},{}],[{},{},{},{},{}],[{},{},{},{},{}],[{},{},{},{},{}],[{},{},{},{},{}]],"Subjects":null},"USD":100,"Research":{"hydrogen":3},"MaxLoad":16,"CurrentLoad":0,"TickRate":10,"Ticks":0}`
	env := `{"version":1,"state":` + legacyState + `}`
	if err := save.Write(saveKey, env); err != nil {
		t.Fatalf("seed v1 save: %v", err)
	}

	_, ok, err := Load()
	if err == nil {
		t.Fatalf("expected version 1 save to be rejected")
	}
	if ok {
		t.Fatalf("expected ok=false for rejected save")
	}
}
