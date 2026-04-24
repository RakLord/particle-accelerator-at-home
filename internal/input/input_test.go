package input_test

import (
	"testing"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/input"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
	"particleaccelerator/internal/ui"
)

func newTestState() (*sim.GameState, *ui.UIState) {
	sim.ResetCostModifiers()
	s := sim.NewGameState()
	// Clear starter inventory so tests control exactly what's available.
	s.Owned = map[sim.ComponentKind]int{}
	u := ui.NewUIState()
	return s, u
}

func TestPlaceFromToolAutoPurchase(t *testing.T) {
	s, u := newTestState()
	s.USD = bignum.FromInt(100)
	u.Selected = ui.ToolAccelerator

	pos := sim.Position{X: 2, Y: 2}
	input.PlaceFromTool(s, u, pos)

	cell := s.Grid.Cells[pos.Y][pos.X]
	if cell.Component == nil {
		t.Fatalf("component not placed")
	}
	if cell.Component.Kind() != sim.KindAccelerator {
		t.Fatalf("placed wrong kind: %s", cell.Component.Kind())
	}
	if s.Owned[sim.KindAccelerator] != 1 {
		t.Fatalf("Owned[accelerator] = %d want 1", s.Owned[sim.KindAccelerator])
	}
	// USD decreased by the accelerator base cost (5).
	if s.USD.GTE(bignum.FromInt(100)) {
		t.Fatalf("USD not deducted: %v", s.USD)
	}
}

func TestPlaceFromToolNoFundsNoInventory(t *testing.T) {
	s, u := newTestState()
	s.USD = bignum.Zero()
	u.Selected = ui.ToolAccelerator

	pos := sim.Position{X: 0, Y: 0}
	input.PlaceFromTool(s, u, pos)

	if cell := s.Grid.Cells[pos.Y][pos.X]; cell.Component != nil {
		t.Fatalf("component placed despite zero inventory and zero USD")
	}
	if s.Owned[sim.KindAccelerator] != 0 {
		t.Fatalf("Owned incremented on failed placement")
	}
}

func TestPlaceFromToolUsesInventoryBeforePurchase(t *testing.T) {
	s, u := newTestState()
	s.USD = bignum.Zero()
	s.Owned[sim.KindAccelerator] = 1
	u.Selected = ui.ToolAccelerator

	pos := sim.Position{X: 1, Y: 1}
	input.PlaceFromTool(s, u, pos)

	if cell := s.Grid.Cells[pos.Y][pos.X]; cell.Component == nil {
		t.Fatalf("inventory placement failed")
	}
	if s.Owned[sim.KindAccelerator] != 1 {
		t.Fatalf("Owned changed despite placing from inventory: %d", s.Owned[sim.KindAccelerator])
	}
	if !s.USD.IsZero() {
		t.Fatalf("USD changed despite placing from inventory: %v", s.USD)
	}
}

func TestEraseReturnsToInventory(t *testing.T) {
	s, u := newTestState()
	s.USD = bignum.FromInt(100)
	u.Selected = ui.ToolAccelerator

	pos := sim.Position{X: 3, Y: 2}
	input.PlaceFromTool(s, u, pos)
	availAfterPlace := sim.CountAvailable(s, sim.KindAccelerator)

	input.Erase(s, pos)

	if cell := s.Grid.Cells[pos.Y][pos.X]; cell.Component != nil {
		t.Fatalf("component not erased")
	}
	availAfterErase := sim.CountAvailable(s, sim.KindAccelerator)
	if availAfterErase != availAfterPlace+1 {
		t.Fatalf("Available after erase: got %d want %d", availAfterErase, availAfterPlace+1)
	}
	if s.Owned[sim.KindAccelerator] != 1 {
		t.Fatalf("Owned changed on erase: %d", s.Owned[sim.KindAccelerator])
	}
}

func TestPickToolAtSelectsHoveredComponent(t *testing.T) {
	s, u := newTestState()
	pos := sim.Position{X: 2, Y: 2}
	s.Grid.Cells[pos.Y][pos.X].Component = &components.MeshGrid{Orientation: sim.DirEast}
	u.Selected = ui.ToolInjector

	input.PickToolAt(s, u, pos)

	if u.Selected != ui.ToolMeshGrid {
		t.Fatalf("Selected = %v, want ToolMeshGrid", u.Selected)
	}
}

func TestPickToolAtSelectsCollector(t *testing.T) {
	s, u := newTestState()
	pos := sim.Position{X: 2, Y: 2}
	s.Grid.Cells[pos.Y][pos.X].IsCollector = true

	input.PickToolAt(s, u, pos)

	if u.Selected != ui.ToolCollector {
		t.Fatalf("Selected = %v, want ToolCollector", u.Selected)
	}
}

func TestPickToolAtIgnoresEmptyCell(t *testing.T) {
	s, u := newTestState()
	u.Selected = ui.ToolAccelerator

	input.PickToolAt(s, u, sim.Position{X: 0, Y: 0})

	if u.Selected != ui.ToolAccelerator {
		t.Fatalf("Selected changed on empty cell: %v", u.Selected)
	}
}

func TestPlaceFromToolOverwriteReturnsOld(t *testing.T) {
	s, u := newTestState()
	s.USD = bignum.FromInt(1000)
	pos := sim.Position{X: 2, Y: 2}

	u.Selected = ui.ToolAccelerator
	input.PlaceFromTool(s, u, pos) // buys Accelerator, places it
	availAccAfterA := sim.CountAvailable(s, sim.KindAccelerator)

	u.Selected = ui.ToolElbow
	input.PlaceFromTool(s, u, pos) // buys Rotator, overwrites cell
	availAccAfterB := sim.CountAvailable(s, sim.KindAccelerator)

	if availAccAfterB != availAccAfterA+1 {
		t.Fatalf("overwritten Accelerator did not return to inventory: before=%d after=%d",
			availAccAfterA, availAccAfterB)
	}
	if cell := s.Grid.Cells[pos.Y][pos.X]; cell.Component == nil ||
		cell.Component.Kind() != sim.KindRotator {
		t.Fatalf("overwrite did not install Rotator")
	}
	if sim.CountAvailable(s, sim.KindRotator) != 0 {
		t.Fatalf("newly-placed Rotator should not also be in inventory")
	}
}

func TestPlacePhase4Components(t *testing.T) {
	cases := []struct {
		tool     ui.Tool
		kind     sim.ComponentKind
		typeName string
	}{
		{ui.ToolResonator, sim.KindResonator, "*components.Resonator"},
		{ui.ToolCatalyst, sim.KindCatalyst, "*components.Catalyst"},
		{ui.ToolDuplicator, sim.KindDuplicator, "*components.Duplicator"},
	}
	for _, c := range cases {
		s, u := newTestState()
		s.USD = bignum.FromInt(100000)
		u.Selected = c.tool
		pos := sim.Position{X: 1, Y: 1}
		input.PlaceFromTool(s, u, pos)
		got := s.Grid.Cells[pos.Y][pos.X].Component
		if got == nil || got.Kind() != c.kind {
			t.Fatalf("%v: got %+v want kind %s", c.tool, got, c.kind)
		}
		if s.Owned[c.kind] != 1 {
			t.Fatalf("%v: Owned not incremented: %d", c.tool, s.Owned[c.kind])
		}
	}
}

func TestDuplicatorReconfigureCyclesOrientation(t *testing.T) {
	s, u := newTestState()
	s.USD = bignum.FromInt(100000)
	u.Selected = ui.ToolDuplicator
	pos := sim.Position{X: 1, Y: 1}
	input.PlaceFromTool(s, u, pos)
	d, ok := s.Grid.Cells[pos.Y][pos.X].Component.(*components.Duplicator)
	if !ok {
		t.Fatalf("Duplicator not placed")
	}
	initial := d.Orientation
	input.Reconfigure(s, pos)
	if d.Orientation == initial {
		t.Fatalf("Duplicator Reconfigure did not rotate Orientation")
	}
}

func TestReconfigureIsFree(t *testing.T) {
	s, u := newTestState()
	s.USD = bignum.FromInt(1000)
	u.Selected = ui.ToolElbow
	pos := sim.Position{X: 0, Y: 0}
	input.PlaceFromTool(s, u, pos)

	usdBefore := s.USD
	ownedBefore := s.Owned[sim.KindRotator]

	input.Reconfigure(s, pos)
	input.Reconfigure(s, pos)

	if !s.USD.Eq(usdBefore) {
		t.Fatalf("USD changed on Reconfigure: before=%v after=%v", usdBefore, s.USD)
	}
	if s.Owned[sim.KindRotator] != ownedBefore {
		t.Fatalf("Owned changed on Reconfigure: before=%d after=%d",
			ownedBefore, s.Owned[sim.KindRotator])
	}
	// Sanity: Reconfigure actually mutated the elbow's Orientation.
	r, ok := s.Grid.Cells[pos.Y][pos.X].Component.(*components.Rotator)
	if !ok {
		t.Fatalf("expected Rotator at placed cell")
	}
	if r.Orientation == sim.DirNorth {
		t.Fatalf("expected elbow orientation to change after reconfigure")
	}
}

func TestReconfigureByRotatesBothDirections(t *testing.T) {
	s, u := newTestState()
	s.USD = bignum.FromInt(1000)
	u.Selected = ui.ToolElbow
	pos := sim.Position{X: 0, Y: 0}
	input.PlaceFromTool(s, u, pos)
	r, ok := s.Grid.Cells[pos.Y][pos.X].Component.(*components.Rotator)
	if !ok {
		t.Fatalf("expected Rotator at placed cell")
	}

	input.ReconfigureBy(s, pos, -1)
	if r.Orientation != sim.DirWest {
		t.Fatalf("reverse rotate orientation = %v, want DirWest", r.Orientation)
	}

	input.ReconfigureBy(s, pos, 1)
	if r.Orientation != sim.DirNorth {
		t.Fatalf("forward rotate orientation = %v, want DirNorth", r.Orientation)
	}
}

func TestMeshGridPlacementAndReconfigureOrientation(t *testing.T) {
	s, u := newTestState()
	s.USD = bignum.FromInt(1000)
	u.Selected = ui.ToolMeshGrid
	pos := sim.Position{X: 1, Y: 1}

	input.PlaceFromTool(s, u, pos)
	mesh, ok := s.Grid.Cells[pos.Y][pos.X].Component.(*components.MeshGrid)
	if !ok {
		t.Fatalf("expected MeshGrid at placed cell")
	}
	if mesh.Orientation != sim.DirEast {
		t.Fatalf("new mesh orientation = %v, want DirEast", mesh.Orientation)
	}

	input.Reconfigure(s, pos)
	if mesh.Orientation != sim.DirSouth {
		t.Fatalf("reconfigured mesh orientation = %v, want DirSouth", mesh.Orientation)
	}
}

func TestPlaceInjectorUsesSingleTool(t *testing.T) {
	s, u := newTestState()
	s.USD = bignum.FromInt(1000)
	u.Selected = ui.ToolInjector

	pos := sim.Position{X: 0, Y: 0}
	input.PlaceFromTool(s, u, pos)

	cell := s.Grid.Cells[pos.Y][pos.X]
	if cell.Component == nil || cell.Component.Kind() != sim.KindInjector {
		t.Fatalf("Injector not placed: %#v", cell.Component)
	}
	if s.Owned[sim.KindInjector] != 1 {
		t.Fatalf("Injector purchase not recorded")
	}
	if s.USD.GTE(bignum.FromInt(1000)) {
		t.Fatalf("Injector purchase did not deduct USD: %v", s.USD)
	}
}
