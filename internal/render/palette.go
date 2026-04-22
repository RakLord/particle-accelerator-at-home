package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/ui"
)

const (
	paletteBtnW   = paletteW - 48
	paletteBtnH   = 64
	paletteBtnGap = 12
	paletteBtnX   = paletteX + 24
)

type paletteEntry struct {
	tool  ui.Tool
	label string
}

var paletteEntries = []paletteEntry{
	{ui.ToolInjector, "Injector (source)"},
	{ui.ToolAccelerator, "Accelerator (+1 Speed)"},
	{ui.ToolRotator, "Rotator (turn)"},
	{ui.ToolCollector, "Collector (endpoint)"},
	{ui.ToolErase, "Erase"},
}

func paletteButtonY(i int) int {
	// Leave room for a heading label at the top.
	return paletteY + 56 + i*(paletteBtnH+paletteBtnGap)
}

func paletteButtonAt(mx, my int) (ui.Tool, bool) {
	for i, e := range paletteEntries {
		if contains(mx, my, paletteBtnX, paletteButtonY(i), paletteBtnW, paletteBtnH) {
			return e.tool, true
		}
	}
	return ui.ToolNone, false
}

func drawPalette(dst *ebiten.Image, u *ui.UIState) {
	fillRect(dst, paletteX, paletteY, paletteW, paletteH, colorPaletteBG)

	drawText(dst, "Components", paletteX+24, paletteY+24, colorText)

	for i, e := range paletteEntries {
		y := paletteButtonY(i)
		bg := colorButton
		if u.Selected == e.tool {
			bg = colorSelected
		}
		fillRect(dst, paletteBtnX, y, paletteBtnW, paletteBtnH, bg)
		strokeRect(dst, paletteBtnX, y, paletteBtnW, paletteBtnH, 1, colorTextMuted)

		swatch := toolColor(e.tool)
		fillRect(dst, paletteBtnX+12, y+16, 32, 32, swatch)
		strokeRect(dst, paletteBtnX+12, y+16, 32, 32, 1, colorTextMuted)

		drawText(dst, e.label, paletteBtnX+60, y+24, colorText)
	}

	// Hint at the bottom.
	hintY := paletteY + paletteH - 64
	drawText(dst, "Left-click: place / reconfigure", paletteX+24, hintY, colorTextMuted)
	drawText(dst, "Right-click: erase", paletteX+24, hintY+16, colorTextMuted)
}

func toolColor(t ui.Tool) interface{ RGBA() (r, g, b, a uint32) } {
	switch t {
	case ui.ToolInjector:
		return colorInjector
	case ui.ToolAccelerator:
		return colorAccelerator
	case ui.ToolRotator:
		return colorRotator
	case ui.ToolCollector:
		return colorCollector
	case ui.ToolErase:
		return colorBG
	}
	return colorButton
}
