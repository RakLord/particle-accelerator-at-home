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
	s.BestStats[ElementHydrogen] = ElementBestStats{
		MaxSpeed:          SpeedFromInt(11),
		MaxMass:           bignum.FromInt(7),
		MaxCollectedValue: bignum.FromInt(99),
	}

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
	stats := loaded.BestStats[ElementHydrogen]
	if stats.MaxSpeed != SpeedFromInt(11) {
		t.Fatalf("Hydrogen MaxSpeed lost on round-trip: got %d", stats.MaxSpeed)
	}
	if !stats.MaxMass.Eq(bignum.FromInt(7)) {
		t.Fatalf("Hydrogen MaxMass lost on round-trip: got %v", stats.MaxMass)
	}
	if !stats.MaxCollectedValue.Eq(bignum.FromInt(99)) {
		t.Fatalf("Hydrogen MaxCollectedValue lost on round-trip: got %v", stats.MaxCollectedValue)
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
	if loaded.BestStats == nil {
		t.Fatalf("save should default BestStats map")
	}
}

func TestLoadV2SaveMigratesSpeedsToFixedPoint(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	grid := `{"Cells":[[{},{},{},{},{}],[{},{},{},{},{}],[{},{},{},{},{}],[{},{},{},{},{}],[{},{},{},{},{}]],"Subjects":[{"Element":"hydrogen","Mass":"1e0","Speed":2,"Direction":1,"Position":{"X":0,"Y":0},"Load":1}]}`
	state := `{"Layer":"genesis","Grid":` + grid + `,"USD":"0","Research":{},"BestStats":{"hydrogen":{"max_speed":3}},"collection_log":[{"element":"hydrogen","mass":"1e0","speed":4,"value":"4e0"}],"UnlockedElements":{"hydrogen":true},"MaxLoad":16,"CurrentLoad":1,"TickRate":10,"Ticks":0}`
	env := `{"version":2,"state":` + state + `}`
	if err := save.Write(saveKey, env); err != nil {
		t.Fatalf("seed v2 save: %v", err)
	}

	loaded, ok, err := Load()
	if err != nil || !ok {
		t.Fatalf("Load v2 save: ok=%v err=%v", ok, err)
	}
	if got := loaded.Grid.Subjects[0].Speed; got != SpeedFromInt(2) {
		t.Fatalf("Subject Speed after migration: got %d want %d", got, SpeedFromInt(2))
	}
	if got := loaded.BestStats[ElementHydrogen].MaxSpeed; got != SpeedFromInt(3) {
		t.Fatalf("BestStats MaxSpeed after migration: got %d want %d", got, SpeedFromInt(3))
	}
	if got := loaded.CollectionLog[0].Speed; got != SpeedFromInt(4) {
		t.Fatalf("CollectionLog Speed after migration: got %d want %d", got, SpeedFromInt(4))
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

func TestSaveLoadPreservesNotifications(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	s := NewGameState()
	s.RecordNotification("Inventory", "Press E to open the Inventory.", "09:41")
	s.MarkHelperMilestoneShown("first-five-usd")
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, ok, err := Load()
	if err != nil || !ok {
		t.Fatalf("Load: ok=%v err=%v", ok, err)
	}
	if got := len(loaded.NotificationLog); got != 1 {
		t.Fatalf("NotificationLog length: got %d want 1", got)
	}
	entry := loaded.NotificationLog[0]
	if entry.Header != "Inventory" || entry.TimeHHMM != "09:41" {
		t.Fatalf("notification mismatch: %+v", entry)
	}
	if !loaded.HasShownHelperMilestone("first-five-usd") {
		t.Fatalf("milestone shown flag was not preserved")
	}
}

func TestSaveLoadPreservesPrestigeFieldsAndResetsTransientAutoCounter(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	s := NewGameState()
	s.BinderReserves[ElementCarbon] = 12
	s.TokenInventory[ElementHydrogen] = 4
	s.BondsState[BondMethane] = true
	s.BondPoints = 3
	s.LaboratoryUpgrades[LabDensePacking] = 1
	s.AutoInjectActive = true
	s.AutoInjectTickCounter = 9
	s.RunCount = 2
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, ok, err := Load()
	if err != nil || !ok {
		t.Fatalf("Load: ok=%v err=%v", ok, err)
	}
	if loaded.BinderReserves[ElementCarbon] != 12 || loaded.TokenInventory[ElementHydrogen] != 4 {
		t.Fatalf("prestige run fields mismatch: reserves=%v tokens=%v", loaded.BinderReserves, loaded.TokenInventory)
	}
	if !loaded.BondsState[BondMethane] || loaded.BondPoints != 3 || loaded.RunCount != 2 || !loaded.AutoInjectActive {
		t.Fatalf("durable prestige fields mismatch: bonds=%v BP=%d run=%d auto=%v", loaded.BondsState, loaded.BondPoints, loaded.RunCount, loaded.AutoInjectActive)
	}
	if loaded.AutoInjectTickCounter != 0 {
		t.Fatalf("AutoInjectTickCounter should be transient, got %d", loaded.AutoInjectTickCounter)
	}
	if got := loaded.Modifiers.Normalized().CollectorValueMul; !got.Eq(bignum.MustParse("1.15")) {
		t.Fatalf("modifiers not rebuilt from bonds: got %v want 1.15", got)
	}
	if got := loaded.Modifiers.Normalized().BinderStoreCapacityMul; !got.Eq(bignum.FromInt(2)) {
		t.Fatalf("modifiers not rebuilt from lab: got %v want 2", got)
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
