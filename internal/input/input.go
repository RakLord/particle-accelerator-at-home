package input

import (
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

const defaultSpawnInterval = 30

// PlaceFromTool writes the currently-selected Tool into the cell at pos.
// An existing component at pos is overwritten.
func PlaceFromTool(s *sim.GameState, u *ui.UIState, pos sim.Position) {
	if !inBounds(pos) {
		return
	}
	cell := &s.Grid.Cells[pos.Y][pos.X]
	switch u.Selected {
	case ui.ToolInjector:
		cell.Component = &sim.Injector{
			Direction:     sim.DirEast,
			SpawnInterval: defaultSpawnInterval,
			Element:       sim.ElementHydrogen,
		}
		cell.IsCollector = false
	case ui.ToolAccelerator:
		cell.Component = &sim.SimpleAccelerator{SpeedBonus: 1}
		cell.IsCollector = false
	case ui.ToolRotator:
		cell.Component = &sim.Rotator{Turn: sim.TurnRight}
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
	case *sim.Injector:
		c.Direction = (c.Direction + 1) % 4
	case *sim.Rotator:
		if c.Turn == sim.TurnLeft {
			c.Turn = sim.TurnRight
		} else {
			c.Turn = sim.TurnLeft
		}
	}
}

func inBounds(p sim.Position) bool {
	return p.X >= 0 && p.X < sim.GridSize && p.Y >= 0 && p.Y < sim.GridSize
}
