package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/sim"
)

// trailLifetime is how many Draw frames a trail sample lives before it fades
// out completely. ~45 frames ≈ 0.75 s at 60 FPS.
const trailLifetime = 45

// trailSpawnEveryNFrames controls sampling density along the trail. 1 = push
// every frame (densest); higher values space samples farther apart.
const trailSpawnEveryNFrames = 1

type trailSample struct {
	X, Y    float32
	Element sim.Element
	Age     int // Draw frames since spawn
}

// updateTrail advances existing trail ages, drops expired samples, and pushes a
// new sample per live Subject when trails are enabled. It's called once per
// Draw from Game.Draw.
func (g *Game) updateTrail(alpha float64) {
	if !g.ui.TrailsEnabled {
		// Don't just zero the slice on toggle-off — that's handled by the hotkey
		// handler. But don't age or append either.
		return
	}

	// Age + compact.
	dst := g.trail[:0]
	for _, s := range g.trail {
		s.Age++
		if s.Age < trailLifetime {
			dst = append(dst, s)
		}
	}
	g.trail = dst

	// Push one sample per Subject at its interpolated pixel position.
	for _, sub := range g.state.Grid.Subjects {
		cx, cy := subjectPixel(sub, alpha)
		g.trail = append(g.trail, trailSample{
			X:       cx,
			Y:       cy,
			Element: sub.Element,
		})
	}
}

// drawTrail renders all trail samples as small fading circles. Called from
// drawGrid before the live Subjects so Subjects sit on top.
func drawTrail(dst *ebiten.Image, samples []trailSample) {
	for _, s := range samples {
		alpha := 1 - float64(s.Age)/float64(trailLifetime)
		if alpha <= 0 {
			continue
		}
		c := subjectColor(s.Element)
		fillCircleFade(dst, s.X, s.Y, 5, c, float32(alpha))
	}
}

// fillCircleFade draws a circle with its color multiplied by alpha — the
// ebiten vector primitives take a solid color, so we premultiply manually.
func fillCircleFade(dst *ebiten.Image, cx, cy, r float32, c color.Color, alpha float32) {
	cr, cg, cb, ca := c.RGBA()
	// Premultiplied RGBA in 8-bit range.
	a := float32(ca>>8) * alpha
	rr := float32(cr>>8) * alpha
	gg := float32(cg>>8) * alpha
	bb := float32(cb>>8) * alpha
	faded := color.RGBA{R: uint8(rr), G: uint8(gg), B: uint8(bb), A: uint8(a)}
	fillCircle(dst, cx, cy, r, faded)
}
