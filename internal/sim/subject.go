package sim

import "particleaccelerator/internal/bignum"

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
	Mass      bignum.Decimal
	Speed     int
	Magnetism bignum.Decimal
	Direction Direction
	Position  Position
	Load      int

	// Transient motion state — not persisted. Render-side interpolation uses
	// these to smoothly place the Subject between ticks; on Load they zero-init
	// and the first frame draws the Subject at Position with no mid-cell offset.
	//
	// Motion model: StepProgress accumulates Speed each tick; a cell boundary
	// crosses when it reaches SpeedDivisor. Render treats t = StepProgress/
	// SpeedDivisor as the fraction through the current cell from its inbound
	// edge midpoint (t=0) to its outbound edge midpoint (t=1), passing through
	// the cell center at t=0.5 for straight cells, or along a quarter arc for
	// elbow-turn cells.
	InDirection      Direction  `json:"-"` // direction subject entered its current cell from; persists until next crossing
	PrevInDirection  Direction  `json:"-"` // snapshot of InDirection at tick start
	PrevPosition     Position   `json:"-"` // snapshot of Position at tick start (== Path[0] when Path is non-empty)
	Path             []Position `json:"-"` // cells visited this tick, starting with PrevPosition
	StepProgress     int        `json:"-"` // accumulator in [0, SpeedDivisor)
	PrevStepProgress int        `json:"-"` // snapshot at tick start
}
