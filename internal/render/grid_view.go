package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"particleaccelerator/internal/sim"
)

func drawGrid(dst *ebiten.Image, s *sim.GameState) {
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

	// Subjects on top.
	for _, sub := range s.Grid.Subjects {
		x, y, w, h := cellRect(sub.Position.X, sub.Position.Y)
		cx := float32(x) + float32(w)/2
		cy := float32(y) + float32(h)/2
		fillCircle(dst, cx, cy, 10, colorSubject)
	}
}

func componentColor(c sim.Component) color.Color {
	switch c.(type) {
	case *sim.Injector:
		return colorInjector
	case *sim.SimpleAccelerator:
		return colorAccelerator
	case *sim.Rotator:
		return colorRotator
	}
	return colorButton
}

// drawComponentGlyph adds a direction arrow for Injector and a turn indicator
// for Rotator. Accelerator gets a simple label.
func drawComponentGlyph(dst *ebiten.Image, c sim.Component, x, y, w, h int) {
	switch v := c.(type) {
	case *sim.Injector:
		drawTextCentered(dst, "IN "+arrowFor(v.Direction), x, y, w, h, colorText)
	case *sim.SimpleAccelerator:
		drawTextCentered(dst, "+"+itoa(v.SpeedBonus), x, y, w, h, colorText)
	case *sim.Rotator:
		cx := float32(x) + float32(w)/2
		cy := float32(y) + float32(h)/2
		r := float32(w) * 0.28
		drawCircularArrow(dst, cx, cy, r, v.Turn == sim.TurnRight, colorText)
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
