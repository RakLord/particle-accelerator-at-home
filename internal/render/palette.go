package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

// The right-side panel is a compact "currently selected" indicator plus
// an Open Inventory button. The full picker lives in the inventory modal
// (drawInventory). See docs/features/0014-inventory.md.

const (
	selectedCardX = paletteX + 24
	selectedCardY = paletteY + 56
	selectedCardW = paletteW - 48
	selectedCardH = 200

	selectedHalfGap = 16
	selectedHalfW   = (selectedCardW - selectedHalfGap) / 2
	selectedLeftX   = selectedCardX
	injectingHalfX  = selectedCardX + selectedHalfW + selectedHalfGap

	selectedIconSize = 96

	openInvBtnW = paletteW - 48
	openInvBtnH = 44
	openInvBtnX = paletteX + 24
	openInvBtnY = selectedCardY + selectedCardH + 16

	injectBtnW = paletteW - 48
	injectBtnH = 52
	injectBtnX = paletteX + 24
	injectBtnY = paletteY + paletteH - injectBtnH - 24
)

// openInvButtonHit reports whether (mx, my) is over the "Open Inventory"
// button rendered by drawPalette.
func openInvButtonHit(mx, my int) bool {
	return contains(mx, my, openInvBtnX, openInvBtnY, openInvBtnW, openInvBtnH)
}

func injectButtonHit(mx, my int) bool {
	return contains(mx, my, injectBtnX, injectBtnY, injectBtnW, injectBtnH)
}

func drawPalette(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	fillRect(dst, paletteX, paletteY, paletteW, paletteH, colorPaletteBG)

	drawText(dst, "Selected", selectedLeftX, paletteY+20, colorText)
	drawText(dst, "Injecting", injectingHalfX, paletteY+20, colorText)

	drawSelectedCard(dst, s, u)
	drawInjectingCard(dst, s)

	// Open Inventory button.
	fillRect(dst, openInvBtnX, openInvBtnY, openInvBtnW, openInvBtnH, colorButton)
	strokeRect(dst, openInvBtnX, openInvBtnY, openInvBtnW, openInvBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Open Inventory  (E)", openInvBtnX, openInvBtnY, openInvBtnW, openInvBtnH, colorText)

	drawPrestigePanel(dst, s, u)
	drawInjectButton(dst, s)
}

func drawInjectButton(dst *ebiten.Image, s *sim.GameState) {
	label, enabled := injectButtonLabel(s)
	cooling := !enabled && s.HasInjector() && s.CurrentLoad < s.EffectiveMaxLoad() && s.InjectionCooldownRemaining > 0

	bg := colorInjector
	fg := colorText
	if !enabled && !cooling {
		bg = color.RGBA{0x1a, 0x2a, 0x20, 0xff}
		fg = colorTextMuted
	}
	fillRect(dst, injectBtnX, injectBtnY, injectBtnW, injectBtnH, bg)
	if cooling {
		drawInjectCooldownOverlay(dst, s)
	}
	strokeRect(dst, injectBtnX, injectBtnY, injectBtnW, injectBtnH, 1, colorTextMuted)
	drawTextFaceCentered(dst, label, injectBtnX, injectBtnY+8, injectBtnW, 22, fontTitle, fg)
	drawTextCentered(dst, "Manual injection", injectBtnX, injectBtnY+32, injectBtnW, 14, colorTextMuted)
}

// drawInjectCooldownOverlay draws a centered dark rectangle whose width is
// proportional to the remaining cooldown, so that as the cooldown counts down
// both edges collapse toward the horizontal center of the button, revealing
// the ready-green background from the outside in.
func drawInjectCooldownOverlay(dst *ebiten.Image, s *sim.GameState) {
	total := s.EffectiveInjectionCooldownTicks()
	if total <= 0 {
		return
	}
	remain := s.InjectionCooldownRemaining
	if remain <= 0 {
		return
	}
	if remain > total {
		remain = total
	}
	overlayW := injectBtnW * remain / total
	if overlayW <= 0 {
		return
	}
	if overlayW > injectBtnW {
		overlayW = injectBtnW
	}
	overlayX := injectBtnX + (injectBtnW-overlayW)/2
	fillRect(dst, overlayX, injectBtnY, overlayW, injectBtnH, color.RGBA{0x1a, 0x2a, 0x20, 0xff})
}

func injectButtonLabel(s *sim.GameState) (string, bool) {
	if !s.HasInjector() {
		return "No Injector", false
	}
	if s.CurrentLoad >= s.EffectiveMaxLoad() {
		return "Max Load", false
	}
	if s.InjectionCooldownRemaining > 0 {
		return "Cooldown " + itoa(cooldownSecondsRemaining(s)) + "s", false
	}
	return "Inject", true
}

func cooldownSecondsRemaining(s *sim.GameState) int {
	tickRate := s.TickRate
	if tickRate <= 0 {
		tickRate = sim.DefaultTickRate
	}
	return (s.InjectionCooldownRemaining + tickRate - 1) / tickRate
}

func drawSelectedCard(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	x, y, w, h := selectedLeftX, selectedCardY, selectedHalfW, selectedCardH
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

func drawInjectingCard(dst *ebiten.Image, s *sim.GameState) {
	x, y, w, h := injectingHalfX, selectedCardY, selectedHalfW, selectedCardH
	bg := color.RGBA{0x14, 0x14, 0x24, 0xff}
	fillRect(dst, x, y, w, h, bg)
	strokeRect(dst, x, y, w, h, 1, colorTextMuted)

	e := s.InjectionElement
	info, ok := sim.ElementCatalog[e]
	symbolW := 80
	symbolX := x + 16
	symbolY := y + (h-symbolW)/2
	if !ok {
		drawTextFaceCentered(dst, "—", symbolX, symbolY, symbolW, symbolW, fontDisplay, colorTextMuted)
		drawTextFace(dst, "No element", symbolX+symbolW+14, y+16, fontBody, colorTextMuted)
		return
	}

	drawTextFaceCentered(dst, info.Symbol, symbolX, symbolY, symbolW, symbolW, fontDisplay, elementAccentColor(e))

	textX := symbolX + symbolW + 14
	drawTextFace(dst, info.Name, textX, y+16, fontBody, colorText)

	rowY := y + 48
	rowGap := 22
	rightX := x + w - 12
	rows := []struct {
		label string
		value string
	}{
		{"Research", itoa(s.Research[e])},
		{"Base Mass", formatNumber(info.BaseMass)},
		{"Speed", formatSpeed(info.BaseSpeed)},
		{"Mult", formatMultiplier(info.Multiplier)},
	}
	for i, r := range rows {
		ry := rowY + i*rowGap
		drawTextSmall(dst, r.label, textX, ry, colorTextMuted)
		vw, _ := measureTextSmall(r.value)
		drawTextSmall(dst, r.value, rightX-vw, ry, colorText)
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
	case ui.ToolBinder:
		return colorBinder
	case ui.ToolErase:
		return colorBG
	}
	return colorButton
}
