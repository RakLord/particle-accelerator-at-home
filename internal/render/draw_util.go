package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func fillRect(dst *ebiten.Image, x, y, w, h int, c color.Color) {
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), c, false)
}

func strokeRect(dst *ebiten.Image, x, y, w, h int, thickness float32, c color.Color) {
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), thickness, c, false)
}

func fillCircle(dst *ebiten.Image, cx, cy float32, r float32, c color.Color) {
	vector.FillCircle(dst, cx, cy, r, c, true)
}

// drawText renders text at the given top-left position.
func drawText(dst *ebiten.Image, s string, x, y int, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, font7x13, op)
}

// drawTextCentered renders text centered within (x, y, w, h).
func drawTextCentered(dst *ebiten.Image, s string, x, y, w, h int, c color.Color) {
	op := &text.DrawOptions{}
	tw, th := text.Measure(s, font7x13, 0)
	op.GeoM.Translate(float64(x)+(float64(w)-tw)/2, float64(y)+(float64(h)-th)/2)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, font7x13, op)
}

func contains(px, py, x, y, w, h int) bool {
	return px >= x && py >= y && px < x+w && py < y+h
}

// drawCircularArrow draws a 300° arc with an arrowhead at the leading end,
// centered on (cx, cy) with radius r. The gap sits at the top of the circle.
// When clockwise is true, the arrow reads as ↻; otherwise as ↺.
func drawCircularArrow(dst *ebiten.Image, cx, cy, r float32, clockwise bool, col color.Color) {
	const (
		arcSteps = 40
		gapRad   = float32(math.Pi / 3) // 60° gap at the top
		headLen  = float32(0.6)         // arrowhead length as fraction of r
		headHalf = float32(0.35)        // arrowhead half-width as fraction of r
		strokeW  = float32(3)
	)
	topAngle := float32(-math.Pi / 2)

	var startA, endA, dirSign float32
	if clockwise {
		startA = topAngle + gapRad/2
		endA = topAngle + 2*math.Pi - gapRad/2
		dirSign = 1
	} else {
		startA = topAngle - gapRad/2
		endA = topAngle - (2*math.Pi - gapRad/2)
		dirSign = -1
	}

	cos := func(a float32) float32 { return float32(math.Cos(float64(a))) }
	sin := func(a float32) float32 { return float32(math.Sin(float64(a))) }

	// Polyline arc (anti-aliased line segments).
	px := cx + r*cos(startA)
	py := cy + r*sin(startA)
	for i := 1; i <= arcSteps; i++ {
		t := float32(i) / float32(arcSteps)
		a := startA + (endA-startA)*t
		nx := cx + r*cos(a)
		ny := cy + r*sin(a)
		vector.StrokeLine(dst, px, py, nx, ny, strokeW, col, true)
		px, py = nx, ny
	}

	// Arrowhead triangle at the leading end.
	baseX := cx + r*cos(endA)
	baseY := cy + r*sin(endA)
	tangentA := endA + dirSign*float32(math.Pi/2) // tangent points in direction of motion
	tipX := baseX + r*headLen*cos(tangentA)
	tipY := baseY + r*headLen*sin(tangentA)
	// Wing base sits slightly before the tip, spread perpendicular to the
	// tangent (i.e., along the radial direction).
	perpA := tangentA + float32(math.Pi/2)
	w1x := baseX + r*headHalf*cos(perpA)
	w1y := baseY + r*headHalf*sin(perpA)
	w2x := baseX - r*headHalf*cos(perpA)
	w2y := baseY - r*headHalf*sin(perpA)

	path := &vector.Path{}
	path.MoveTo(tipX, tipY)
	path.LineTo(w1x, w1y)
	path.LineTo(w2x, w2y)
	path.Close()
	var cs ebiten.ColorScale
	cs.ScaleWithColor(col)
	vector.FillPath(dst, path, nil, &vector.DrawPathOptions{
		AntiAlias:  true,
		ColorScale: cs,
	})
}
