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
