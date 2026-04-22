package input

import (
	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
	"particleaccelerator/internal/ui"
)

const defaultSpawnInterval = 30

// toolKind maps a Tool selection to the ComponentKind that represents it for
// cost / ownership purposes. Returns "" for tools that don't participate in
// the inventory system (ToolNone, ToolErase). Both injector tools resolve to
// KindInjector — element variants are free to pick, only the kind costs.
func toolKind(t ui.Tool) sim.ComponentKind {
	switch t {
	case ui.ToolInjectorHydrogen, ui.ToolInjectorHelium:
		return sim.KindInjector
	case ui.ToolAccelerator:
		return sim.KindAccelerator
	case ui.ToolMeshGrid:
		return sim.KindMeshGrid
	case ui.ToolMagnetiser:
		return sim.KindMagnetiser
	case ui.ToolElbow:
		return sim.KindRotator
	case ui.ToolCollector:
		return sim.KindCollector
	}
	return ""
}

// PlaceFromTool writes the currently-selected Tool into the cell at pos.
// An existing component at pos is overwritten; the displaced component
// returns to the available inventory automatically (Owned is monotonic,
// so decrementing the placed count for the old kind bumps its Available).
// Placing a Helium Injector while Helium is locked is a no-op. Placing
// when inventory is empty auto-purchases if affordable; otherwise no-op.
func PlaceFromTool(s *sim.GameState, u *ui.UIState, pos sim.Position) {
	if !inBounds(pos) {
		return
	}
	if u.Selected == ui.ToolErase {
		Erase(s, pos)
		return
	}
	if u.Selected == ui.ToolNone {
		return
	}
	// Element-gate Helium injector before spending inventory.
	if u.Selected == ui.ToolInjectorHelium && !sim.IsElementUnlocked(s, sim.ElementHelium) {
		return
	}
	kind := toolKind(u.Selected)
	if kind == "" {
		return
	}
	if sim.CountAvailable(s, kind) <= 0 {
		if err := sim.PurchaseComponent(s, kind); err != nil {
			return
		}
	}

	cell := &s.Grid.Cells[pos.Y][pos.X]
	switch u.Selected {
	case ui.ToolInjectorHydrogen:
		cell.Component = &components.Injector{
			Direction:     sim.DirEast,
			SpawnInterval: defaultSpawnInterval,
			Element:       sim.ElementHydrogen,
		}
		cell.IsCollector = false
	case ui.ToolInjectorHelium:
		cell.Component = &components.Injector{
			Direction:     sim.DirEast,
			SpawnInterval: defaultSpawnInterval,
			Element:       sim.ElementHelium,
		}
		cell.IsCollector = false
	case ui.ToolAccelerator:
		cell.Component = &components.SimpleAccelerator{SpeedBonus: 1, Orientation: sim.DirNorth}
		cell.IsCollector = false
	case ui.ToolMeshGrid:
		cell.Component = &components.MeshGrid{}
		cell.IsCollector = false
	case ui.ToolMagnetiser:
		cell.Component = &components.Magnetiser{Bonus: bignum.One()}
		cell.IsCollector = false
	case ui.ToolElbow:
		cell.Component = &components.Rotator{Orientation: sim.DirNorth}
		cell.IsCollector = false
	case ui.ToolCollector:
		cell.Component = nil
		cell.IsCollector = true
	}
}

// Erase clears the cell at pos. The component that was there returns to
// available inventory automatically (Owned is monotonic, placed count drops).
func Erase(s *sim.GameState, pos sim.Position) {
	if !inBounds(pos) {
		return
	}
	cell := &s.Grid.Cells[pos.Y][pos.X]
	cell.Component = nil
	cell.IsCollector = false
}

// Reconfigure cycles the orientation of directional tiles already at pos.
func Reconfigure(s *sim.GameState, pos sim.Position) {
	if !inBounds(pos) {
		return
	}
	cell := &s.Grid.Cells[pos.Y][pos.X]
	switch c := cell.Component.(type) {
	case *components.Injector:
		c.Direction = (c.Direction + 1) % 4
	case *components.SimpleAccelerator:
		c.Orientation = (c.Orientation + 1) % 4
	case *components.Rotator:
		c.Orientation = (c.Orientation + 1) % 4
	}
}

func inBounds(p sim.Position) bool {
	return p.X >= 0 && p.X < sim.GridSize && p.Y >= 0 && p.Y < sim.GridSize
}
