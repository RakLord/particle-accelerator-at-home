package render

import (
	"math"

	"particleaccelerator/internal/sim"
)

// subjectPixel returns the interpolated pixel position of a Subject at render
// alpha α ∈ [0,1] within the current tick. Uses sim.Path, StepProgress, and
// the inbound-direction snapshot to walk along a mixed-segment polycurve where
// elbow-turn cells are rendered as quarter arcs and all other cells as straight
// lines. See docs/features/smooth-motion.md for the geometric model.
func subjectPixel(sub sim.Subject, alpha float64) (float32, float32) {
	n := len(sub.Path)
	if n == 0 {
		// First frame after Load (Path is transient/unsaved) — snap to cell.
		cx, cy := cellCenterF(sub.Position)
		return cx, cy
	}

	var cellIdx int
	var localT float64

	if n == 1 {
		// No boundary crossings this tick. Progress slides from PrevStepProgress
		// to StepProgress within Path[0].
		progress := float64(sub.PrevStepProgress) + (float64(sub.StepProgress)-float64(sub.PrevStepProgress))*alpha
		cellIdx = 0
		localT = clamp01(progress / float64(sim.StepProgressPerCell))
	} else {
		// At least one crossing. Walk through Path consuming virtual progress,
		// with the first cell starting at PrevStepProgress fraction.
		virtual := alpha * float64(sub.Speed)
		firstCap := float64(sim.StepProgressPerCell - sub.PrevStepProgress)
		if virtual <= firstCap {
			cellIdx = 0
			localT = (float64(sub.PrevStepProgress) + virtual) / float64(sim.StepProgressPerCell)
		} else {
			virtual -= firstCap
			cellIdx = 1
			for cellIdx < n-1 && virtual > float64(sim.StepProgressPerCell) {
				virtual -= float64(sim.StepProgressPerCell)
				cellIdx++
			}
			localT = clamp01(virtual / float64(sim.StepProgressPerCell))
		}
	}

	dirIn, dirOut := cellDirections(sub, cellIdx)
	return cellInternalPos(sub.Path[cellIdx], dirIn, dirOut, localT)
}

// cellDirections resolves the inbound and outbound directions for the cell at
// sub.Path[cellIdx]. The first cell uses sub.PrevInDirection (snapshot of how
// the subject arrived there); the last cell uses sub.Direction (post-Apply
// direction at tick end).
func cellDirections(sub sim.Subject, cellIdx int) (sim.Direction, sim.Direction) {
	n := len(sub.Path)
	var dirIn, dirOut sim.Direction

	if cellIdx == 0 {
		dirIn = sub.PrevInDirection
	} else {
		dirIn = dirBetween(sub.Path[cellIdx-1], sub.Path[cellIdx])
	}

	if cellIdx == n-1 {
		dirOut = sub.Direction
	} else {
		dirOut = dirBetween(sub.Path[cellIdx], sub.Path[cellIdx+1])
	}

	return dirIn, dirOut
}

// dirBetween returns the cardinal direction from adjacent cell a to cell b.
// Undefined when a and b are not orthogonally adjacent.
func dirBetween(a, b sim.Position) sim.Direction {
	switch {
	case b.X > a.X:
		return sim.DirEast
	case b.X < a.X:
		return sim.DirWest
	case b.Y > a.Y:
		return sim.DirSouth
	case b.Y < a.Y:
		return sim.DirNorth
	}
	return sim.DirEast
}

// cellInternalPos returns the pixel position inside a cell at fraction t from
// its inbound edge midpoint (t=0) to its outbound edge midpoint (t=1).
// Straight when dirIn == dirOut (line through cell center), quarter arc
// otherwise (centered on the cell corner shared by the inbound and outbound
// edges, radius = cellSize/2).
func cellInternalPos(cell sim.Position, dirIn, dirOut sim.Direction, t float64) (float32, float32) {
	x, y, w, _ := cellRect(cell.X, cell.Y)
	cx := float32(x) + float32(w)/2
	cy := float32(y) + float32(w)/2
	r := float32(w) / 2

	inSide := opposite(dirIn)
	ix, iy := edgeMidpoint(cx, cy, r, inSide)
	ox, oy := edgeMidpoint(cx, cy, r, dirOut)

	if dirIn == dirOut {
		return lerp32(ix, ox, float32(t)), lerp32(iy, oy, float32(t))
	}

	// Quarter arc centered on the corner shared by inSide and dirOut edges.
	innerX, innerY := cornerPixel(cx, cy, r, inSide, dirOut)
	startA := math.Atan2(float64(iy-innerY), float64(ix-innerX))
	endA := math.Atan2(float64(oy-innerY), float64(ox-innerX))
	// Take the short arc (|diff| should be π/2).
	diff := endA - startA
	if diff > math.Pi {
		diff -= 2 * math.Pi
	} else if diff < -math.Pi {
		diff += 2 * math.Pi
	}
	a := startA + diff*t
	px := float32(float64(innerX) + float64(r)*math.Cos(a))
	py := float32(float64(innerY) + float64(r)*math.Sin(a))
	return px, py
}

// opposite returns the cardinal direction 180° from d.
func opposite(d sim.Direction) sim.Direction { return (d + 2) % 4 }

// edgeMidpoint returns the midpoint of the cell's edge on the given side,
// expressed as a direction from the cell center.
func edgeMidpoint(cx, cy, r float32, side sim.Direction) (float32, float32) {
	dx, dy := side.Step()
	return cx + float32(dx)*r, cy + float32(dy)*r
}

// cornerPixel returns the pixel position of the corner of the cell shared by
// the edges on sides a and b. a and b must be perpendicular cardinal directions.
func cornerPixel(cx, cy, r float32, a, b sim.Direction) (float32, float32) {
	ax, ay := a.Step()
	bx, by := b.Step()
	return cx + float32(ax+bx)*r, cy + float32(ay+by)*r
}

func cellCenterF(p sim.Position) (float32, float32) {
	x, y, w, _ := cellRect(p.X, p.Y)
	return float32(x) + float32(w)/2, float32(y) + float32(w)/2
}

func lerp32(a, b, t float32) float32 { return a + (b-a)*t }

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
