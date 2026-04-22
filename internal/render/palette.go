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
	{ui.ToolElbow, "Elbow (turn)"},
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

		kind := kindForTool(e.tool)
		unaffordable := kind != "" &&
			sim.CountAvailable(s, kind) == 0 &&
			!sim.CanPurchase(s, kind)
		dimmed := locked || unaffordable

		bg := colorButton
		if u.Selected == e.tool {
			bg = colorSelected
		}
		fillRect(dst, paletteBtnX, y, paletteBtnW, paletteBtnH, bg)
		strokeRect(dst, paletteBtnX, y, paletteBtnW, paletteBtnH, 1, colorTextMuted)

		iconX, iconY, iconSize := paletteBtnX+12, y+12, 32
		if sprite := tileSpriteForTool(e.tool); sprite != nil {
			drawSpriteFitted(dst, sprite, iconX, iconY, iconSize, iconSize)
			if dimmed {
				fillRect(dst, iconX, iconY, iconSize, iconSize, colorOverlay)
			}
		} else {
			swatch := toolColor(e.tool)
			if dimmed {
				swatch = colorButton
			}
			fillRect(dst, iconX, iconY, iconSize, iconSize, swatch)
		}
		strokeRect(dst, iconX, iconY, iconSize, iconSize, 1, colorTextMuted)

		labelColor := colorText
		if dimmed {
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
// there's nothing to show. Helium Injector has its own lock-status line;
// every other purchasable tool shows "have: N · $cost".
func subLabel(s *sim.GameState, t ui.Tool) string {
	if t == ui.ToolInjectorHelium {
		if !sim.IsElementUnlocked(s, sim.ElementHelium) {
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
		// Helium unlocked: fall through to the inventory line.
	}
	kind := kindForTool(t)
	if kind == "" {
		return ""
	}
	have := sim.CountAvailable(s, kind)
	cost := sim.ComponentCost(s, kind)
	return "have: " + itoa(have) + " · next: " + formatUSD(cost)
}

// kindForTool mirrors input.toolKind but is duplicated here because the input
// package isn't a dependency of render. Both injector tools share a single
// kind for inventory/cost accounting.
func kindForTool(t ui.Tool) sim.ComponentKind {
	switch t {
	case ui.ToolInjectorHydrogen, ui.ToolInjectorHelium:
		return sim.KindInjector
	case ui.ToolAccelerator:
		return sim.KindAccelerator
	case ui.ToolMeshGrid:
		return sim.KindMeshGrid
	case ui.ToolMagnetiser:
		return sim.KindMagnetiser
	case ui.ToolElbow:
		return sim.KindRotator
	case ui.ToolCollector:
		return sim.KindCollector
	}
	return ""
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
	case ui.ToolElbow:
		return colorRotator
	case ui.ToolCollector:
		return colorCollector
	case ui.ToolErase:
		return colorBG
	}
	return colorButton
}
