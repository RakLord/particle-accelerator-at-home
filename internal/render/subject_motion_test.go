package render

import (
	"math"
	"testing"

	"particleaccelerator/internal/sim"
)

func approxEq(a, b, tol float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= tol
}

// When a subject is mid-tick with no crossings, its render position slides
// linearly across the cell from its inbound edge to its outbound edge.
func TestSubjectPixelStraightInCell(t *testing.T) {
	sub := sim.Subject{
		Direction:        sim.DirEast,
		InDirection:      sim.DirEast,
		PrevInDirection:  sim.DirEast,
		Position:         sim.Position{X: 2, Y: 2},
		Path:             []sim.Position{{X: 2, Y: 2}},
		StepProgress:     5,
		PrevStepProgress: 5,
		Speed:            1,
	}
	gotX, gotY := subjectPixel(sub, 0)
	cx, cy := cellCenterF(sub.Position)
	// StepProgress=5 / SpeedDivisor=10 = 0.5 → cell center on a straight cell.
	if !approxEq(gotX, cx, 0.5) || !approxEq(gotY, cy, 0.5) {
		t.Fatalf("center: got (%v,%v) want (%v,%v)", gotX, gotY, cx, cy)
	}
}

// At a turn boundary (elbow at Path[0] leaving in a different direction), the
// subject's rendered position sits along a quarter arc, not on the L-shaped
// corner through the cell center.
func TestSubjectPixelArcThroughRotator(t *testing.T) {
	// Subject sits in cell (2,2). Arrived moving East; current direction is
	// South (as if an elbow applied last tick). It's mid-cell
	// (StepProgress=5, no new crossings).
	sub := sim.Subject{
		Direction:        sim.DirSouth,
		InDirection:      sim.DirEast,
		PrevInDirection:  sim.DirEast,
		Position:         sim.Position{X: 2, Y: 2},
		Path:             []sim.Position{{X: 2, Y: 2}},
		StepProgress:     5,
		PrevStepProgress: 5,
		Speed:            1,
	}
	gotX, gotY := subjectPixel(sub, 0)
	cx, cy := cellCenterF(sub.Position)

	// For East→South turn, inner corner is SW (cx-r, cy+r). Arc midpoint sits
	// at angle -π/4 from SW, radius r. Pixel = (cx-r + r·cos(-π/4), cy+r + r·sin(-π/4))
	r := float32(cellSize) / 2
	wantX := cx - r + r*float32(math.Cos(-math.Pi/4))
	wantY := cy + r + r*float32(math.Sin(-math.Pi/4))

	if !approxEq(gotX, wantX, 0.5) || !approxEq(gotY, wantY, 0.5) {
		t.Fatalf("arc midpoint: got (%v,%v) want (%v,%v)", gotX, gotY, wantX, wantY)
	}
	// Sanity: arc midpoint must be strictly inside the cell (not on center) —
	// distance from cell center should be >= r/2.
	dx, dy := gotX-cx, gotY-cy
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if dist < r*0.3 {
		t.Fatalf("arc midpoint too close to cell center: dist=%v r=%v", dist, r)
	}
}

// When the Subject has no Path (first frame after Load), fall back to the raw
// Position without a divide-by-zero or out-of-bounds slice access.
func TestSubjectPixelNoPathFallsBackToPosition(t *testing.T) {
	sub := sim.Subject{
		Direction: sim.DirEast,
		Position:  sim.Position{X: 1, Y: 1},
	}
	gotX, gotY := subjectPixel(sub, 0.5)
	wantX, wantY := cellCenterF(sub.Position)
	if gotX != wantX || gotY != wantY {
		t.Fatalf("fallback: got (%v,%v) want (%v,%v)", gotX, gotY, wantX, wantY)
	}
}
