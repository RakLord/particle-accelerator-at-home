package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

// The right-side panel is a compact "currently selected" indicator plus
// an Open Inventory button. The full picker lives in the inventory modal
// (drawInventory). See docs/features/inventory.md.

const (
	selectedCardX = paletteX + 24
	selectedCardY = paletteY + 56
	selectedCardW = paletteW - 48
	selectedCardH = 200

	selectedIconSize = 96

	openInvBtnW = paletteW - 48
	openInvBtnH = 44
	openInvBtnX = paletteX + 24
	openInvBtnY = selectedCardY + selectedCardH + 16
)

// openInvButtonHit reports whether (mx, my) is over the "Open Inventory"
// button rendered by drawPalette.
func openInvButtonHit(mx, my int) bool {
	return contains(mx, my, openInvBtnX, openInvBtnY, openInvBtnW, openInvBtnH)
}

func drawPalette(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	fillRect(dst, paletteX, paletteY, paletteW, paletteH, colorPaletteBG)

	drawText(dst, "Selected", paletteX+24, paletteY+20, colorText)

	drawSelectedCard(dst, s, u)

	// Open Inventory button.
	fillRect(dst, openInvBtnX, openInvBtnY, openInvBtnW, openInvBtnH, colorButton)
	strokeRect(dst, openInvBtnX, openInvBtnY, openInvBtnW, openInvBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Open Inventory  (E)", openInvBtnX, openInvBtnY, openInvBtnW, openInvBtnH, colorText)

	// Hint footer.
	hintY := paletteY + paletteH - 56
	drawText(dst, "Left-click: place / reconfigure", paletteX+24, hintY, colorTextMuted)
	drawText(dst, "Right-click: erase", paletteX+24, hintY+16, colorTextMuted)
	drawText(dst, "Q: pick tile · scroll: rotate", paletteX+24, hintY+32, colorTextMuted)
}

func drawSelectedCard(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	x, y, w, h := selectedCardX, selectedCardY, selectedCardW, selectedCardH
	bg := color.RGBA{0x14, 0x14, 0x24, 0xff}
	fillRect(dst, x, y, w, h, bg)
	strokeRect(dst, x, y, w, h, 1, colorTextMuted)

	if u.Selected == ui.ToolNone {
		drawTextCentered(dst, "Nothing selected", x, y+h/2-20, w, 18, colorTextMuted)
		drawTextCentered(dst, "Press E to open inventory", x, y+h/2+4, w, 18, colorTextMuted)
		return
	}

	info := ui.ToolInfoCatalog[u.Selected]
	kind := ui.KindForTool(u.Selected)

	// Icon on the left.
	iconX := x + 16
	iconY := y + (h-selectedIconSize)/2
	if sprite := tileSpriteForTool(u.Selected); sprite != nil {
		drawSpriteFitted(dst, sprite, iconX, iconY, selectedIconSize, selectedIconSize)
	} else {
		fillRect(dst, iconX, iconY, selectedIconSize, selectedIconSize, toolColor(u.Selected))
	}
	strokeRect(dst, iconX, iconY, selectedIconSize, selectedIconSize, 1, colorTextMuted)

	// Text block on the right.
	textX := iconX + selectedIconSize + 14
	textW := x + w - textX - 12
	drawTextFace(dst, info.Name, textX, y+16, fontBody, colorText)

	if info.Tagline != "" {
		drawTextWrapped(dst, info.Tagline, textX, y+38, textW, fontSmall, colorTextMuted)
	}

	if kind != "" {
		available := sim.CountAvailable(s, kind)
		cost := formatUSD(sim.ComponentCost(s, kind))
		statY := y + h - 38
		drawTextSmall(dst, "have: "+itoa(available), textX, statY, colorText)
		drawTextSmall(dst, "next: "+cost, textX, statY+16, colorTextMuted)
	}

	// Lock callout for future gated tools.
	if reason := ui.ToolLockReason(s, u.Selected); reason != "" {
		drawTextSmall(dst, "Locked", x+12, y+h-16, colorResetArmed)
	}
}

// toolColor is retained as the colour-swatch fallback used by the
// inventory modal and the selected-card icon when no sprite is available.
func toolColor(t ui.Tool) color.Color {
	switch t {
	case ui.ToolInjector:
		return colorInjector
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
