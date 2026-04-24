package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

const (
	invModalW = 960
	invModalH = 600

	invPadding     = 24
	invHeaderH     = 56
	invDescW       = 280
	invCardCols    = 4
	invCardRows    = 3
	invCardSize    = 140
	invCardGap     = 12
	invCardTopH    = 22
	invCardBottomH = 22
)

func invModalX() int { return (screenW - invModalW) / 2 }
func invModalY() int { return (screenH - invModalH) / 2 }

func invGridX() int { return invModalX() + invPadding }
func invGridY() int { return invModalY() + invHeaderH }

func invGridW() int {
	return invCardCols*invCardSize + (invCardCols-1)*invCardGap
}

func invDescX() int { return invModalX() + invModalW - invDescW - invPadding }
func invDescY() int { return invGridY() }
func invDescH() int { return invModalH - invHeaderH - invPadding }

func invCloseX() int { return invModalX() + invModalW - closeBtnW - 12 }
func invCloseY() int { return invModalY() + 12 }

// invCardRect returns the on-screen rect for the i-th card (row-major,
// 0-indexed). Out-of-range indices return a zero rect.
func invCardRect(i int) (x, y, w, h int) {
	if i < 0 || i >= invCardCols*invCardRows {
		return 0, 0, 0, 0
	}
	col := i % invCardCols
	row := i / invCardCols
	x = invGridX() + col*(invCardSize+invCardGap)
	y = invGridY() + row*(invCardSize+invCardGap)
	return x, y, invCardSize, invCardSize
}

// invToolAt returns the Tool whose card contains (mx, my), or ToolNone if
// the cursor is not over any card.
func invToolAt(mx, my int) ui.Tool {
	for i, t := range ui.PlaceableTools {
		if i >= invCardCols*invCardRows {
			break
		}
		x, y, w, h := invCardRect(i)
		if contains(mx, my, x, y, w, h) {
			return t
		}
	}
	return ui.ToolNone
}

// invInPanel reports whether (mx, my) falls inside the inventory modal's
// outer panel. Used to detect click-outside-to-close.
func invInPanel(mx, my int) bool {
	return contains(mx, my, invModalX(), invModalY(), invModalW, invModalH)
}

func drawInventory(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	fillRect(dst, 0, 0, screenW, screenH, colorOverlay)

	x, y := invModalX(), invModalY()
	fillRect(dst, x, y, invModalW, invModalH, colorModalBG)
	strokeRect(dst, x, y, invModalW, invModalH, 2, colorTextMuted)

	drawTextFaceCentered(dst, "Inventory", x, y+12, invModalW, 28, fontTitle, colorText)
	drawTextCentered(dst, "Click a component to select it · E or Close to dismiss", x, y+38, invModalW, 16, colorTextMuted)

	cx, cy := invCloseX(), invCloseY()
	fillRect(dst, cx, cy, closeBtnW, closeBtnH, colorButton)
	strokeRect(dst, cx, cy, closeBtnW, closeBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Close", cx, cy, closeBtnW, closeBtnH, colorText)

	for i, t := range ui.PlaceableTools {
		if i >= invCardCols*invCardRows {
			break
		}
		drawInventoryCard(dst, s, u, t, i)
	}

	drawInventoryDescription(dst, s, u)
}

func drawInventoryCard(dst *ebiten.Image, s *sim.GameState, u *ui.UIState, t ui.Tool, i int) {
	x, y, w, h := invCardRect(i)
	kind := ui.KindForTool(t)
	unlocked := ui.IsToolUnlocked(s, t)
	available := 0
	if kind != "" {
		available = sim.CountAvailable(s, kind)
	}
	affordable := kind != "" && sim.CanPurchase(s, kind)
	dimmed := !unlocked || (available == 0 && !affordable)

	bg := colorButton
	if u.Selected == t {
		bg = colorSelected
	}
	fillRect(dst, x, y, w, h, bg)
	border := colorTextMuted
	if u.InventoryHovered == t {
		border = colorText
	}
	strokeRect(dst, x, y, w, h, 1, border)

	// Top strip: quantity available, right-aligned.
	topY := y + 4
	if unlocked && kind != "" {
		qty := "x" + itoa(available)
		qw, _ := measureTextSmall(qty)
		drawTextSmall(dst, qty, x+w-qw-8, topY, colorTextMuted)
	} else {
		drawTextSmall(dst, "—", x+w-16, topY, colorTextMuted)
	}

	// Middle: icon centred in the card body.
	iconSize := 80
	iconX := x + (w-iconSize)/2
	iconY := y + invCardTopH + ((h-invCardTopH-invCardBottomH)-iconSize)/2
	if sprite := logoSpriteForTool(t); sprite != nil {
		drawSpriteFitted(dst, sprite, iconX, iconY, iconSize, iconSize)
	} else {
		fillRect(dst, iconX, iconY, iconSize, iconSize, toolColor(t))
	}
	if dimmed {
		fillRect(dst, x+1, y+1, w-2, h-2, colorOverlay)
	}

	// Bottom strip: cost or lock label.
	bottomY := y + h - invCardBottomH + 4
	switch {
	case !unlocked:
		drawTextCentered(dst, "Locked", x, bottomY-2, w, 16, colorTextMuted)
	case kind == "":
		// Tools without a cost (none in PlaceableTools today, but keep robust).
		drawTextCentered(dst, "—", x, bottomY-2, w, 16, colorTextMuted)
	default:
		cost := formatUSD(sim.ComponentCost(s, kind))
		col := colorText
		if !affordable && available == 0 {
			col = colorResetArmed
		}
		drawTextCentered(dst, cost, x, bottomY-2, w, 16, col)
	}
}

func drawInventoryDescription(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	x, y, w, h := invDescX(), invDescY(), invDescW, invDescH()
	bg := color.RGBA{0x10, 0x10, 0x20, 0xf0}
	fillRect(dst, x, y, w, h, bg)
	strokeRect(dst, x, y, w, h, 1, colorTextMuted)

	t := u.InventoryHovered
	if t == ui.ToolNone {
		drawTextCentered(dst, "Hover a component for details.", x+12, y+h/2-10, w-24, 20, colorTextMuted)
		return
	}

	info, ok := ui.ToolInfoCatalog[t]
	if !ok {
		drawTextCentered(dst, "(no info)", x+12, y+h/2-10, w-24, 20, colorTextMuted)
		return
	}

	innerX := x + 14
	innerW := w - 28
	cur := y + 16

	drawTextFace(dst, info.Name, innerX, cur, fontTitle, colorText)
	cur += 30

	if info.Tagline != "" {
		drawTextWrapped(dst, info.Tagline, innerX, cur, innerW, fontSmall, colorHeaderIncome)
		cur += wrappedHeight(info.Tagline, innerW, fontSmall) + 12
	}

	if info.Description != "" {
		drawTextWrapped(dst, info.Description, innerX, cur, innerW, fontSmall, colorText)
		cur += wrappedHeight(info.Description, innerW, fontSmall) + 16
	}

	if reason := ui.ToolLockReason(s, t); reason != "" {
		drawTextWrapped(dst, "Locked: "+reason, innerX, cur, innerW, fontSmall, colorResetArmed)
		cur += wrappedHeight("Locked: "+reason, innerW, fontSmall) + 12
	}

	// Stat strip at the bottom of the panel.
	statsY := y + h - 96
	kind := ui.KindForTool(t)
	if kind == "" {
		return
	}
	owned := 0
	if s.Owned != nil {
		owned = s.Owned[kind]
	}
	placed := sim.CountPlaced(s, kind)
	available := sim.CountAvailable(s, kind)

	drawTextSmall(dst, "Owned", innerX, statsY, colorTextMuted)
	drawTextSmall(dst, itoa(owned), innerX+90, statsY, colorText)
	drawTextSmall(dst, "Placed", innerX, statsY+18, colorTextMuted)
	drawTextSmall(dst, itoa(placed), innerX+90, statsY+18, colorText)
	drawTextSmall(dst, "Available", innerX, statsY+36, colorTextMuted)
	drawTextSmall(dst, itoa(available), innerX+90, statsY+36, colorText)
	drawTextSmall(dst, "Next", innerX, statsY+54, colorTextMuted)
	drawTextSmall(dst, formatUSD(sim.ComponentCost(s, kind)), innerX+90, statsY+54, colorText)
}
