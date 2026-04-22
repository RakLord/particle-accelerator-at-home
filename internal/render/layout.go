package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"

	"particleaccelerator/internal/sim"
)

// Logical resolution. See docs/overview.md — 1280×720 with Ebitengine
// scaling to the window.
const (
	screenW = 1280
	screenH = 720

	headerH = 48

	gridAreaX = 0
	gridAreaY = headerH
	gridAreaW = 672
	gridAreaH = screenH - headerH

	cellSize    = 120
	gridPadding = (gridAreaW - cellSize*sim.GridSize) / 2

	paletteX = gridAreaW
	paletteY = headerH
	paletteW = screenW - gridAreaW
	paletteH = screenH - headerH
)

var (
	colorBG          = color.RGBA{0x0a, 0x0a, 0x14, 0xff}
	colorHeaderBG    = color.RGBA{0x14, 0x14, 0x28, 0xff}
	colorPaletteBG   = color.RGBA{0x10, 0x10, 0x1e, 0xff}
	colorGridLine    = color.RGBA{0x33, 0x33, 0x5a, 0xff}
	colorText        = color.RGBA{0xf0, 0xf0, 0xff, 0xff}
	colorTextMuted   = color.RGBA{0x88, 0x88, 0xaa, 0xff}
	colorButton      = color.RGBA{0x22, 0x22, 0x3a, 0xff}
	colorButtonHover = color.RGBA{0x33, 0x33, 0x55, 0xff}
	colorSelected    = color.RGBA{0x55, 0x55, 0x88, 0xff}

	colorInjector    = color.RGBA{0x3e, 0xbc, 0x6c, 0xff}
	colorAccelerator = color.RGBA{0x4e, 0x8a, 0xff, 0xff}
	colorRotator     = color.RGBA{0xf0, 0xc4, 0x3a, 0xff}
	colorCollector   = color.RGBA{0xe0, 0x5a, 0x5a, 0xff}
	colorSubject     = color.RGBA{0xff, 0xff, 0xff, 0xff}
	colorModalBG     = color.RGBA{0x1a, 0x1a, 0x2a, 0xff}
	colorOverlay     = color.RGBA{0x00, 0x00, 0x00, 0xaa}
	colorResetArmed  = color.RGBA{0xff, 0x44, 0x44, 0xff}
)

var font7x13 = text.NewGoXFace(basicfont.Face7x13)

// cellRect returns the logical-pixel bounds of grid cell (cx, cy).
func cellRect(cx, cy int) (x, y, w, h int) {
	return gridAreaX + gridPadding + cx*cellSize,
		gridAreaY + gridPadding + cy*cellSize,
		cellSize, cellSize
}

// cellAt returns the grid cell at logical coordinates (x, y), or ok=false if
// the coordinates fall outside the grid area.
func cellAt(x, y int) (sim.Position, bool) {
	lx := x - gridAreaX - gridPadding
	ly := y - gridAreaY - gridPadding
	if lx < 0 || ly < 0 {
		return sim.Position{}, false
	}
	cx, cy := lx/cellSize, ly/cellSize
	if cx >= sim.GridSize || cy >= sim.GridSize {
		return sim.Position{}, false
	}
	return sim.Position{X: cx, Y: cy}, true
}
