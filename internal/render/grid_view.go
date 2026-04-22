package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
)

func drawGrid(dst *ebiten.Image, s *sim.GameState, alpha float64, trail []trailSample) {
	fillRect(dst, gridAreaX, gridAreaY, gridAreaW, gridAreaH, colorBG)

	// Cells (background + component fill + collector overlay).
	for cy := range sim.GridSize {
		for cx := range sim.GridSize {
			x, y, w, h := cellRect(cx, cy)
			cell := s.Grid.Cells[cy][cx]
			if cell.Component != nil {
				fillRect(dst, x+2, y+2, w-4, h-4, componentColor(cell.Component))
				drawComponentGlyph(dst, cell.Component, x, y, w, h)
			}
			if cell.IsCollector {
				fillRect(dst, x+2, y+2, w-4, h-4, colorCollector)
				drawTextCentered(dst, "OUT", x, y, w, h, colorText)
			}
		}
	}

	// Grid lines.
	xLeft := float32(gridAreaX + gridPadding)
	xRight := float32(gridAreaX + gridPadding + sim.GridSize*cellSize)
	yTop := float32(gridAreaY + gridPadding)
	yBot := float32(gridAreaY + gridPadding + sim.GridSize*cellSize)
	for i := 0; i <= sim.GridSize; i++ {
		x := float32(gridAreaX + gridPadding + i*cellSize)
		y := float32(gridAreaY + gridPadding + i*cellSize)
		vector.StrokeLine(dst, x, yTop, x, yBot, 1, colorGridLine, false)
		vector.StrokeLine(dst, xLeft, y, xRight, y, 1, colorGridLine, false)
	}

	// Trail (below live Subjects so the current particle sits on top).
	drawTrail(dst, trail)

	// Subjects: interpolated along recorded Path with quarter arcs through rotators.
	for _, sub := range s.Grid.Subjects {
		cx, cy := subjectPixel(sub, alpha)
		fillCircle(dst, cx, cy, 10, subjectColor(sub.Element))
	}
}

func subjectColor(e sim.Element) color.Color {
	switch e {
	case sim.ElementHelium:
		return colorSubjectHelium
	}
	return colorSubject
}

func componentColor(c sim.Component) color.Color {
	switch v := c.(type) {
	case *components.Injector:
		if v.Element == sim.ElementHelium {
			return colorInjectorHelium
		}
		return colorInjector
	case *components.SimpleAccelerator:
		return colorAccelerator
	case *components.MeshGrid:
		return colorMeshGrid
	case *components.Magnetiser:
		return colorMagnetiser
	case *components.Rotator:
		return colorRotator
	}
	return colorButton
}

// drawComponentGlyph adds a direction arrow for Injector and a turn indicator
// for Rotator. Accelerator gets a simple label.
func drawComponentGlyph(dst *ebiten.Image, c sim.Component, x, y, w, h int) {
	switch v := c.(type) {
	case *components.Injector:
		symbol := sim.ElementCatalog[v.Element].Symbol
		if symbol == "" {
			symbol = "?"
		}
		drawTextCentered(dst, symbol+" "+arrowFor(v.Direction), x, y, w, h, colorText)
	case *components.SimpleAccelerator:
		drawTextCentered(dst, "+"+itoa(v.SpeedBonus), x, y, w, h, colorText)
	case *components.MeshGrid:
		drawTextCentered(dst, "×½", x, y, w, h, colorText)
	case *components.Magnetiser:
		drawTextCentered(dst, "M", x, y, w, h, colorText)
	case *components.Rotator:
		cx := float32(x) + float32(w)/2
		cy := float32(y) + float32(h)/2
		r := float32(w) * 0.28
		drawCircularArrow(dst, cx, cy, r, v.Turn == components.TurnRight, colorText)
	}
}

func arrowFor(d sim.Direction) string {
	switch d {
	case sim.DirNorth:
		return "^"
	case sim.DirEast:
		return ">"
	case sim.DirSouth:
		return "v"
	case sim.DirWest:
		return "<"
	}
	return "?"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
