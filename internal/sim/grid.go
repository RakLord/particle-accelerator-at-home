package sim

const GridSize = 5

type Cell struct {
	Component   Component
	IsCollector bool
}

type Grid struct {
	Cells    [GridSize][GridSize]Cell
	Subjects []Subject
}

func NewGrid() *Grid { return &Grid{} }
