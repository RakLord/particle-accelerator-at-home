package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"particleaccelerator/internal/sim"
)

func drawGrid(dst *ebiten.Image, s *sim.GameState, alpha float64, trail []trailSample) {
	fillRect(dst, gridAreaX, gridAreaY, gridAreaW, gridAreaH, colorBG)
	drawLoadBar(dst, s)

	// Cells: background pass.
	for cy := range sim.GridSize {
		for cx := range sim.GridSize {
			x, y, w, h := cellRect(cx, cy)
			drawSpriteFitted(dst, sprites.emptyTile, x, y, w, h)
		}
	}

	// Component base pass below live Subjects.
	for cy := range sim.GridSize {
		for cx := range sim.GridSize {
			x, y, w, h := cellRect(cx, cy)
			cell := s.Grid.Cells[cy][cx]
			if cell.Component != nil {
				layers := spriteLayersForComponent(cell.Component)
				if len(layers.base) > 0 {
					drawSpriteLayers(dst, layers.base, x, y, w, h)
				}
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

	// Subjects: interpolated along recorded Path with quarter arcs through elbows.
	for _, sub := range s.Grid.Subjects {
		cx, cy := subjectPixel(sub, alpha)
		fillCircle(dst, cx, cy, 10, subjectColor(sub.Element))
	}

	// Component top pass so Subjects render inside the tube where top art exists.
	for cy := range sim.GridSize {
		for cx := range sim.GridSize {
			cell := s.Grid.Cells[cy][cx]
			if cell.Component == nil {
				continue
			}
			layers := spriteLayersForComponent(cell.Component)
			if len(layers.top) == 0 {
				continue
			}
			x, y, w, h := cellRect(cx, cy)
			drawSpriteLayers(dst, layers.top, x, y, w, h)
		}
	}

	// Collector pass above subjects: transparent regions of the collector art
	// reveal subjects and tails passing underneath.
	for cy := range sim.GridSize {
		for cx := range sim.GridSize {
			cell := s.Grid.Cells[cy][cx]
			if !cell.IsCollector {
				continue
			}
			x, y, w, h := cellRect(cx, cy)
			drawSpriteCenteredRotated(dst, sprites.collector, x, y, w, h, collectorRotation(cell.CollectorDirection))
		}
	}
}

func drawLoadBar(dst *ebiten.Image, s *sim.GameState) {
	barX := gridAreaX + gridPadding
	barY := gridAreaY + 10
	barW := sim.GridSize * cellSize
	barH := 20
	cap := s.EffectiveMaxLoad()
	fillRect(dst, barX, barY, barW, barH, colorButton)
	fillW := loadBarFillWidth(s.CurrentLoad, cap, barW)
	if fillW > 0 {
		fillRect(dst, barX, barY, fillW, barH, colorPurchaseActive)
	}
	strokeRect(dst, barX, barY, barW, barH, 1, colorTextMuted)
	drawTextCentered(dst, "Load: "+itoa(s.CurrentLoad)+"/"+itoa(cap), barX, barY+1, barW, barH, colorText)
}

func loadBarFillWidth(current, cap, width int) int {
	if cap <= 0 || current <= 0 || width <= 0 {
		return 0
	}
	if current >= cap {
		return width
	}
	return width * current / cap
}

func drawSpriteLayers(dst *ebiten.Image, layers []spriteLayer, x, y, w, h int) {
	for _, layer := range layers {
		drawSpriteCenteredRotated(dst, layer.image, x, y, w, h, layer.rotation)
	}
}

func subjectColor(e sim.Element) color.Color {
	palette := [...]color.RGBA{
		colorSubject,
		colorSubjectHelium,
		{0xff, 0x9f, 0x6e, 0xff},
		{0xff, 0xcf, 0x70, 0xff},
		{0x9f, 0xe8, 0x72, 0xff},
		{0x6f, 0xe0, 0xb0, 0xff},
		{0x73, 0xde, 0xff, 0xff},
		{0x82, 0xb7, 0xff, 0xff},
		{0xb2, 0x94, 0xff, 0xff},
		{0xf0, 0x90, 0xff, 0xff},
	}
	if info, ok := sim.ElementCatalog[e]; ok && info.AtomicNumber > 0 {
		return palette[(info.AtomicNumber-1)%len(palette)]
	}
	return colorSubject
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
