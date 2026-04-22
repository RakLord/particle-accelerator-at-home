package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

const (
	paletteBtnW   = paletteW - 48
	paletteBtnH   = 56
	paletteBtnGap = 8
	paletteBtnX   = paletteX + 24
)

type paletteEntry struct {
	tool  ui.Tool
	label string
}

var paletteEntries = []paletteEntry{
	{ui.ToolInjectorHydrogen, "Injector · Hydrogen"},
	{ui.ToolInjectorHelium, "Injector · Helium"},
	{ui.ToolAccelerator, "Accelerator (+1 Speed)"},
	{ui.ToolMeshGrid, "Mesh Grid (×½ Speed)"},
	{ui.ToolMagnetiser, "Magnetiser (+1 Mag)"},
	{ui.ToolRotator, "Rotator (turn)"},
	{ui.ToolCollector, "Collector (endpoint)"},
	{ui.ToolErase, "Erase"},
}

func paletteButtonY(i int) int {
	// Leave room for a heading label at the top.
	return paletteY + 40 + i*(paletteBtnH+paletteBtnGap)
}

func paletteButtonAt(mx, my int) (ui.Tool, bool) {
	for i, e := range paletteEntries {
		if contains(mx, my, paletteBtnX, paletteButtonY(i), paletteBtnW, paletteBtnH) {
			return e.tool, true
		}
	}
	return ui.ToolNone, false
}

func drawPalette(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	fillRect(dst, paletteX, paletteY, paletteW, paletteH, colorPaletteBG)

	drawText(dst, "Components", paletteX+24, paletteY+16, colorText)

	for i, e := range paletteEntries {
		y := paletteButtonY(i)
		locked := e.tool == ui.ToolInjectorHelium && !sim.IsElementUnlocked(s, sim.ElementHelium)

		bg := colorButton
		if u.Selected == e.tool {
			bg = colorSelected
		}
		fillRect(dst, paletteBtnX, y, paletteBtnW, paletteBtnH, bg)
		strokeRect(dst, paletteBtnX, y, paletteBtnW, paletteBtnH, 1, colorTextMuted)

		swatch := toolColor(e.tool)
		if locked {
			swatch = colorButton
		}
		fillRect(dst, paletteBtnX+12, y+12, 32, 32, swatch)
		strokeRect(dst, paletteBtnX+12, y+12, 32, 32, 1, colorTextMuted)

		labelColor := colorText
		if locked {
			labelColor = colorTextMuted
		}
		drawText(dst, e.label, paletteBtnX+60, y+14, labelColor)

		if sub := subLabel(s, e.tool); sub != "" {
			drawText(dst, sub, paletteBtnX+60, y+32, colorTextMuted)
		}
	}

	// Hint at the bottom.
	hintY := paletteY + paletteH - 56
	drawText(dst, "Left-click: place / reconfigure", paletteX+24, hintY, colorTextMuted)
	drawText(dst, "Right-click: erase", paletteX+24, hintY+16, colorTextMuted)
	drawText(dst, "Codex in header opens the Periodic Table.", paletteX+24, hintY+32, colorTextMuted)
}

// subLabel returns the secondary status line for a palette entry, or "" when
// there's nothing to show. Currently only the Helium Injector has state.
func subLabel(s *sim.GameState, t ui.Tool) string {
	if t != ui.ToolInjectorHelium {
		return ""
	}
	if sim.IsElementUnlocked(s, sim.ElementHelium) {
		return "Unlocked"
	}
	if sim.IsElementPurchasable(s, sim.ElementHelium) {
		return "Purchase in Codex"
	}
	info := sim.ElementCatalog[sim.ElementHelium]
	need := info.ResearchThreshold - s.Research[info.UnlocksFrom]
	if need < 0 {
		need = 0
	}
	return "Locked · need " + itoa(need) + " more H research"
}

func toolColor(t ui.Tool) color.Color {
	switch t {
	case ui.ToolInjectorHydrogen:
		return colorInjector
	case ui.ToolInjectorHelium:
		return colorInjectorHelium
	case ui.ToolAccelerator:
		return colorAccelerator
	case ui.ToolMeshGrid:
		return colorMeshGrid
	case ui.ToolMagnetiser:
		return colorMagnetiser
	case ui.ToolRotator:
		return colorRotator
	case ui.ToolCollector:
		return colorCollector
	case ui.ToolErase:
		return colorBG
	}
	return colorButton
}
