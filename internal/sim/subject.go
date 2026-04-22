package sim

type Direction uint8

const (
	DirNorth Direction = iota
	DirEast
	DirSouth
	DirWest
)

func (d Direction) Step() (dx, dy int) {
	switch d {
	case DirNorth:
		return 0, -1
	case DirEast:
		return 1, 0
	case DirSouth:
		return 0, 1
	case DirWest:
		return -1, 0
	}
	return 0, 0
}

func (d Direction) Left() Direction  { return (d + 3) % 4 }
func (d Direction) Right() Direction { return (d + 1) % 4 }

type Position struct {
	X, Y int
}

type Subject struct {
	Element   Element
	Mass      float64
	Speed     int
	Magnetism float64
	Direction Direction
	Position  Position
	Load      int
}
