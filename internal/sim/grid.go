package sim

const GridSize = 5

type Cell struct {
	Component          Component
	IsCollector        bool
	CollectorDirection Direction
}

type Grid struct {
	Cells    [GridSize][GridSize]Cell
	Subjects []Subject
}

func NewGrid() *Grid { return &Grid{} }

// gridView is the unexported read-only wrapper passed to components via
// ApplyContext. It hands out Cells by value and returns fresh slices from
// SubjectsAt so callers cannot mutate live grid data.
type gridView struct{ g *Grid }

func newGridView(g *Grid) GridView { return gridView{g: g} }

func (v gridView) CellAt(p Position) (Cell, bool) {
	if !v.InBounds(p) {
		return Cell{}, false
	}
	return v.g.Cells[p.Y][p.X], true
}

func (v gridView) SubjectsAt(p Position) []Subject {
	if !v.InBounds(p) {
		return nil
	}
	var out []Subject
	for _, s := range v.g.Subjects {
		if s.Position == p {
			out = append(out, s)
		}
	}
	return out
}

func (v gridView) InBounds(p Position) bool {
	return p.X >= 0 && p.X < GridSize && p.Y >= 0 && p.Y < GridSize
}

func (v gridView) Size() int { return GridSize }
