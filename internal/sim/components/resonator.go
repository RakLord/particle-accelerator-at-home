package components

import "particleaccelerator/internal/sim"

// Resonator adds a tier-driven Speed bonus per orthogonally-adjacent Resonator
// on the grid. Isolated Resonators are inert; clusters compound. See
// docs/features/0010-component-resonator.md.
type Resonator struct{}

// resonatorBonusPerNeighbourByTier is the per-neighbour Speed bonus at each
// tier. Index 0 unused.
var resonatorBonusPerNeighbourByTier = []sim.Speed{0, sim.SpeedFromInt(1), sim.SpeedFromInt(2), sim.SpeedFromInt(3)}

func (*Resonator) Kind() sim.ComponentKind { return sim.KindResonator }

func (r *Resonator) Apply(ctx sim.ApplyContext, s sim.Subject) (sim.Subject, bool) {
	neighbours := countAdjacentResonators(ctx.Grid, ctx.Pos)
	if neighbours == 0 {
		return s, false
	}
	tier := sim.ClampTier(ctx.Tiers, sim.KindResonator, len(resonatorBonusPerNeighbourByTier)-1)
	s.Speed += sim.Speed(neighbours) * resonatorBonusPerNeighbourByTier[tier]
	return s, false
}

// countAdjacentResonators returns the number of N/S/E/W neighbours that hold a
// Resonator. Diagonals do not count. A nil GridView yields 0.
func countAdjacentResonators(g sim.GridView, pos sim.Position) int {
	if g == nil {
		return 0
	}
	n := 0
	for _, d := range []sim.Direction{sim.DirNorth, sim.DirEast, sim.DirSouth, sim.DirWest} {
		dx, dy := d.Step()
		cell, ok := g.CellAt(sim.Position{X: pos.X + dx, Y: pos.Y + dy})
		if !ok || cell.Component == nil {
			continue
		}
		if cell.Component.Kind() == sim.KindResonator {
			n++
		}
	}
	return n
}

func init() {
	sim.RegisterComponent(sim.KindResonator, func() sim.Component { return &Resonator{} })
}
