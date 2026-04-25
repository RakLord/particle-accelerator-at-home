package input

import (
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
	"particleaccelerator/internal/ui"
)

const defaultSpawnInterval = 30

// PlaceFromTool writes the currently-selected Tool into the cell at pos.
// An existing component at pos is overwritten; the displaced component
// returns to the available inventory automatically (Owned is monotonic,
// so decrementing the placed count for the old kind bumps its Available).
// Placing when inventory is empty auto-purchases if affordable; otherwise no-op.
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
	kind := ui.KindForTool(u.Selected)
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
	case ui.ToolInjector:
		cell.Component = &components.Injector{
			Direction:     sim.DirEast,
			SpawnInterval: defaultSpawnInterval,
		}
		cell.IsCollector = false
	case ui.ToolAccelerator:
		cell.Component = &components.SimpleAccelerator{Orientation: sim.DirNorth}
		cell.IsCollector = false
	case ui.ToolMeshGrid:
		cell.Component = &components.MeshGrid{Orientation: sim.DirEast}
		cell.IsCollector = false
	case ui.ToolMagnetiser:
		cell.Component = &components.Magnetiser{Orientation: sim.DirEast}
		cell.IsCollector = false
	case ui.ToolElbow:
		cell.Component = &components.Rotator{Orientation: sim.DirNorth}
		cell.IsCollector = false
	case ui.ToolPipe:
		cell.Component = &components.Pipe{Orientation: sim.DirEast}
		cell.IsCollector = false
	case ui.ToolCollector:
		cell.Component = nil
		cell.IsCollector = true
	case ui.ToolResonator:
		cell.Component = &components.Resonator{}
		cell.IsCollector = false
	case ui.ToolCatalyst:
		cell.Component = &components.Catalyst{}
		cell.IsCollector = false
	case ui.ToolDuplicator:
		cell.Component = &components.Duplicator{Orientation: sim.DirWest}
		cell.IsCollector = false
	case ui.ToolCompressor:
		cell.Component = &components.Compressor{Orientation: sim.DirEast}
		cell.IsCollector = false
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

// PickToolAt selects the inventory tool matching the cell at pos. Empty cells
// are ignored so an accidental pick over blank grid doesn't clear selection.
func PickToolAt(s *sim.GameState, u *ui.UIState, pos sim.Position) {
	if !inBounds(pos) {
		return
	}
	cell := s.Grid.Cells[pos.Y][pos.X]
	if cell.IsCollector {
		u.Selected = ui.ToolCollector
		return
	}
	switch cell.Component.(type) {
	case *components.Injector:
		u.Selected = ui.ToolInjector
	case *components.SimpleAccelerator:
		u.Selected = ui.ToolAccelerator
	case *components.MeshGrid:
		u.Selected = ui.ToolMeshGrid
	case *components.Magnetiser:
		u.Selected = ui.ToolMagnetiser
	case *components.Rotator:
		u.Selected = ui.ToolElbow
	case *components.Pipe:
		u.Selected = ui.ToolPipe
	case *components.Resonator:
		u.Selected = ui.ToolResonator
	case *components.Catalyst:
		u.Selected = ui.ToolCatalyst
	case *components.Duplicator:
		u.Selected = ui.ToolDuplicator
	case *components.Compressor:
		u.Selected = ui.ToolCompressor
	}
}

// Reconfigure cycles the orientation of directional tiles already at pos.
func Reconfigure(s *sim.GameState, pos sim.Position) {
	ReconfigureBy(s, pos, 1)
}

// ReconfigureBy rotates directional tiles already at pos by steps quarter-turns.
// Positive steps match Reconfigure's existing cycle direction; negative steps
// rotate the other way.
func ReconfigureBy(s *sim.GameState, pos sim.Position, steps int) {
	if !inBounds(pos) {
		return
	}
	if steps == 0 {
		return
	}
	cell := &s.Grid.Cells[pos.Y][pos.X]
	switch c := cell.Component.(type) {
	case *components.Injector:
		c.Direction = rotateDirection(c.Direction, steps)
	case *components.SimpleAccelerator:
		c.Orientation = rotateDirection(c.Orientation, steps)
	case *components.MeshGrid:
		c.Orientation = rotateDirection(c.Orientation, steps)
	case *components.Magnetiser:
		c.Orientation = rotateDirection(c.Orientation, steps)
	case *components.Rotator:
		c.Orientation = rotateDirection(c.Orientation, steps)
	case *components.Pipe:
		c.Orientation = rotateDirection(c.Orientation, steps)
	case *components.Duplicator:
		c.Orientation = rotateDirection(c.Orientation, steps)
	case *components.Compressor:
		c.Orientation = rotateDirection(c.Orientation, steps)
	}
}

func rotateDirection(d sim.Direction, steps int) sim.Direction {
	return sim.Direction((int(d) + steps%4 + 4) % 4)
}

func inBounds(p sim.Position) bool {
	return p.X >= 0 && p.X < sim.GridSize && p.Y >= 0 && p.Y < sim.GridSize
}
