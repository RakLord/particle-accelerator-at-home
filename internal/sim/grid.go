package sim

const GridSize = 11

type Cell struct {
	Upgrader Upgrader
}

type Grid struct {
	Cells [GridSize][GridSize]Cell
	Orbs  []Orb
}

func NewGrid() *Grid { return &Grid{} }
