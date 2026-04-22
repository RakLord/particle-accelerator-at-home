package input

import (
	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
	"particleaccelerator/internal/ui"
)

const defaultSpawnInterval = 30

// PlaceFromTool writes the currently-selected Tool into the cell at pos.
// An existing component at pos is overwritten. Placing a Helium Injector
// while Helium is locked is a no-op.
func PlaceFromTool(s *sim.GameState, u *ui.UIState, pos sim.Position) {
	if !inBounds(pos) {
		return
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
		if !sim.IsElementUnlocked(s, sim.ElementHelium) {
			return
		}
		cell.Component = &components.Injector{
			Direction:     sim.DirEast,
			SpawnInterval: defaultSpawnInterval,
			Element:       sim.ElementHelium,
		}
		cell.IsCollector = false
	case ui.ToolAccelerator:
		cell.Component = &components.SimpleAccelerator{SpeedBonus: 1}
		cell.IsCollector = false
	case ui.ToolMeshGrid:
		cell.Component = &components.MeshGrid{}
		cell.IsCollector = false
	case ui.ToolMagnetiser:
		cell.Component = &components.Magnetiser{Bonus: bignum.One()}
		cell.IsCollector = false
	case ui.ToolRotator:
		cell.Component = &components.Rotator{Turn: components.TurnRight}
		cell.IsCollector = false
	case ui.ToolCollector:
		cell.Component = nil
		cell.IsCollector = true
	case ui.ToolErase:
		Erase(s, pos)
	}
}

// Erase clears the cell at pos.
func Erase(s *sim.GameState, pos sim.Position) {
	if !inBounds(pos) {
		return
	}
	cell := &s.Grid.Cells[pos.Y][pos.X]
	cell.Component = nil
	cell.IsCollector = false
}

// Reconfigure cycles the configuration of whatever is already at pos:
// Injector → next direction; Rotator → flip turn. No-op for other kinds.
func Reconfigure(s *sim.GameState, pos sim.Position) {
	if !inBounds(pos) {
		return
	}
	cell := &s.Grid.Cells[pos.Y][pos.X]
	switch c := cell.Component.(type) {
	case *components.Injector:
		c.Direction = (c.Direction + 1) % 4
	case *components.Rotator:
		if c.Turn == components.TurnLeft {
			c.Turn = components.TurnRight
		} else {
			c.Turn = components.TurnLeft
		}
	}
}

func inBounds(p sim.Position) bool {
	return p.X >= 0 && p.X < sim.GridSize && p.Y >= 0 && p.Y < sim.GridSize
}
