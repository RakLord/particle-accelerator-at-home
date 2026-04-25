package render

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

const (
	prestigePanelX = paletteX + 24
	prestigePanelY = openInvBtnY + openInvBtnH + 14
	prestigePanelW = paletteW - 48
	prestigePanelH = injectBtnY - prestigePanelY - 14

	prestigeTabH   = 28
	prestigeTabGap = 8

	prestigeRowH       = 30
	prestigeButtonW    = 96
	prestigeButtonH    = 24
	prestigeButtonGapY = 8

	prestigeConfirmW       = 520
	prestigeConfirmH       = 340
	prestigeConfirmButtonW = 132
	prestigeConfirmButtonH = 38
)

func prestigeConfirmX() int { return (screenW - prestigeConfirmW) / 2 }
func prestigeConfirmY() int { return (screenH - prestigeConfirmH) / 2 }

func drawPrestigePanel(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	fillRect(dst, prestigePanelX, prestigePanelY, prestigePanelW, prestigePanelH, color.RGBA{0x12, 0x12, 0x22, 0xff})
	strokeRect(dst, prestigePanelX, prestigePanelY, prestigePanelW, prestigePanelH, 1, colorTextMuted)

	drawText(dst, "Carbon Loop", prestigePanelX+12, prestigePanelY+10, colorText)
	if s.RunCount > 0 {
		run := "Run #" + itoa(s.RunCount+1)
		drawTextSmall(dst, run, prestigePanelX+132, prestigePanelY+13, colorTextMuted)
	}

	drawPrestigeTabs(dst, s, u)
	contentY := prestigePanelY + 76
	contentH := prestigePanelH - 88
	if u.PrestigeTab == ui.PrestigeTabBonds && canShowBondsTab(s) {
		drawBondsPanel(dst, s, contentY, contentH)
	} else {
		drawBinderStorePanel(dst, s, contentY, contentH)
	}
	drawPrestigeNoticeAndButton(dst, s, u)
}

func drawPrestigeTabs(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	storeX, storeY, storeW, storeH := storeTabRect()
	drawPrestigeTab(dst, "Store", storeX, storeY, storeW, storeH, u.PrestigeTab != ui.PrestigeTabBonds)
	if canShowBondsTab(s) {
		bondX, bondY, bondW, bondH := bondsTabRect()
		drawPrestigeTab(dst, "Bonds", bondX, bondY, bondW, bondH, u.PrestigeTab == ui.PrestigeTabBonds)
	}
}

func drawPrestigeTab(dst *ebiten.Image, label string, x, y, w, h int, active bool) {
	bg := colorButton
	fg := colorTextMuted
	if active {
		bg = colorSelected
		fg = colorText
	}
	fillRect(dst, x, y, w, h, bg)
	strokeRect(dst, x, y, w, h, 1, colorTextMuted)
	drawTextCentered(dst, label, x, y, w, h, fg)
}

func storeTabRect() (x, y, w, h int) {
	w = (prestigePanelW - 24 - prestigeTabGap) / 2
	return prestigePanelX + 12, prestigePanelY + 42, w, prestigeTabH
}

func bondsTabRect() (x, y, w, h int) {
	storeX, storeY, storeW, storeH := storeTabRect()
	return storeX + storeW + prestigeTabGap, storeY, storeW, storeH
}

func drawBinderStorePanel(dst *ebiten.Image, s *sim.GameState, y, h int) {
	drawTextSmall(dst, "Reserve / capacity", prestigePanelX+12, y, colorTextMuted)
	y += 20
	for i, e := range sim.BinderStoreElementOrder {
		drawBinderStoreRow(dst, s, e, i, y)
	}
	_ = h
}

func drawBinderStoreRow(dst *ebiten.Image, s *sim.GameState, e sim.Element, i int, startY int) {
	x := prestigePanelX + 12
	y := startY + i*prestigeRowH
	w := prestigePanelW - 24
	fillRect(dst, x, y, w, prestigeRowH-4, color.RGBA{0x18, 0x18, 0x2a, 0xff})
	info := sim.ElementCatalog[e]
	drawTextSmall(dst, info.Symbol, x+8, y+7, elementAccentColor(e))
	drawTextSmall(dst, info.Name, x+34, y+7, colorText)
	reserve := s.BinderReserves[e]
	cap := s.EffectiveBinderStoreCapacity(e)
	count := itoa(reserve) + "/" + itoa(cap)
	cw, _ := measureTextSmall(count)
	drawTextSmall(dst, count, x+176-cw, y+7, reserveColor(reserve, cap))
	cost := sim.CrystallisationCost(e, s.TokenInventory[e])
	drawTextSmall(dst, "next "+itoa(cost), x+190, y+7, colorTextMuted)
	bx, by, bw, bh := crystalliseButtonRect(i)
	drawSmallActionButton(dst, bx, by, bw, bh, "Token", sim.CanCrystalliseToken(s, e))
}

func reserveColor(reserve, cap int) color.Color {
	if cap > 0 && reserve >= cap {
		return colorResetArmed
	}
	return colorText
}

func crystalliseButtonRect(i int) (x, y, w, h int) {
	return prestigePanelX + prestigePanelW - prestigeButtonW - 12,
		prestigePanelY + 96 + i*prestigeRowH,
		prestigeButtonW,
		prestigeButtonH
}

func drawBondsPanel(dst *ebiten.Image, s *sim.GameState, y, h int) {
	drawTokenSummary(dst, s, y)
	rowY := y + 24
	for i, id := range sim.BondCatalogOrder {
		drawBondRow(dst, s, id, i, rowY)
	}
	_ = h
}

func drawTokenSummary(dst *ebiten.Image, s *sim.GameState, y int) {
	parts := make([]string, 0, len(sim.BinderStoreElementOrder))
	for _, e := range sim.BinderStoreElementOrder {
		if n := s.TokenInventory[e]; n > 0 {
			parts = append(parts, sim.ElementCatalog[e].Symbol+":"+itoa(n))
		}
	}
	if len(parts) == 0 {
		parts = append(parts, "no Tokens")
	}
	drawTextSmall(dst, "Tokens  "+strings.Join(parts, "  "), prestigePanelX+12, y, colorTextMuted)
}

func drawBondRow(dst *ebiten.Image, s *sim.GameState, id sim.BondID, i int, startY int) {
	bond := sim.BondCatalog[id]
	x := prestigePanelX + 12
	y := startY + i*(prestigeRowH+2)
	w := prestigePanelW - 24
	owned := s.BondsState[id]
	bg := color.RGBA{0x18, 0x18, 0x2a, 0xff}
	if owned {
		bg = color.RGBA{0x16, 0x26, 0x1c, 0xff}
	}
	fillRect(dst, x, y, w, prestigeRowH-4, bg)
	drawTextSmall(dst, bond.Name+" "+bond.Formula, x+8, y+4, colorText)
	drawTextSmall(dst, bondTokenCostLabel(bond), x+8, y+18, colorTextMuted)
	bp := "+" + itoa(bond.BondPoints) + " BP"
	drawTextSmall(dst, bp, x+200, y+9, colorHeaderIncome)
	bx, by, bw, bh := synthesiseButtonRect(i)
	label := "Make"
	if owned {
		label = "Owned"
	}
	drawSmallActionButton(dst, bx, by, bw, bh, label, sim.CanSynthesiseBond(s, id))
}

func bondTokenCostLabel(b sim.Bond) string {
	parts := make([]string, 0, len(b.TokenCost))
	for _, e := range sim.BinderStoreElementOrder {
		if n := b.TokenCost[e]; n > 0 {
			parts = append(parts, itoa(n)+sim.ElementCatalog[e].Symbol)
		}
	}
	return strings.Join(parts, " + ")
}

func synthesiseButtonRect(i int) (x, y, w, h int) {
	return prestigePanelX + prestigePanelW - prestigeButtonW - 12,
		prestigePanelY + 100 + i*(prestigeRowH+1),
		prestigeButtonW,
		prestigeButtonH
}

func drawSmallActionButton(dst *ebiten.Image, x, y, w, h int, label string, enabled bool) {
	bg := color.RGBA{0x20, 0x20, 0x30, 0xff}
	fg := colorTextMuted
	if enabled {
		bg = colorPurchaseActive
		fg = colorText
	}
	fillRect(dst, x, y, w, h, bg)
	strokeRect(dst, x, y, w, h, 1, colorTextMuted)
	drawTextCentered(dst, label, x, y, w, h, fg)
}

func drawPrestigeNoticeAndButton(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	buttonVisible := sim.HasAnyBond(s)
	if u.PrestigeNotice != "" {
		drawTextSmall(dst, u.PrestigeNotice, prestigePanelX+12, prestigePanelY+72, colorTextMuted)
	}
	if !buttonVisible {
		return
	}
	x, y, w, h := prestigeButtonRect()
	fillRect(dst, x, y, w, h, colorBinder)
	strokeRect(dst, x, y, w, h, 1, colorTextMuted)
	drawTextCentered(dst, "Prestige", x, y, w, h, colorText)
}

func prestigeButtonRect() (x, y, w, h int) {
	return prestigePanelX + prestigePanelW - 124,
		prestigePanelY + 8,
		112,
		28
}

func canShowBondsTab(s *sim.GameState) bool {
	return sim.HasAnyToken(s) || sim.HasAnyBond(s)
}

func drawPrestigeConfirmModal(dst *ebiten.Image, s *sim.GameState) {
	fillRect(dst, 0, 0, screenW, screenH, colorOverlay)
	x, y := prestigeConfirmX(), prestigeConfirmY()
	fillRect(dst, x, y, prestigeConfirmW, prestigeConfirmH, colorModalBG)
	strokeRect(dst, x, y, prestigeConfirmW, prestigeConfirmH, 2, colorBinder)
	drawTextFaceCentered(dst, "Genesis Ascension", x, y+18, prestigeConfirmW, 28, fontTitle, colorText)
	drawTextWrapped(dst,
		"Prestige starts a fresh Genesis run. Grid layout, $USD, research, unlocked Elements, Binder reserves, Tokens, and in-flight Subjects will reset.",
		x+34, y+70, prestigeConfirmW-68, fontBody, colorText)
	drawTextWrapped(dst,
		"Bonds, Bond Points, Laboratory upgrades, Best Stats, and Auto-Inject preference persist. Bond effects are active immediately on the next run.",
		x+34, y+150, prestigeConfirmW-68, fontBody, colorHeaderIncome)
	if s.RunCount > 0 {
		drawTextCentered(dst, "Current run: #"+itoa(s.RunCount+1), x, y+222, prestigeConfirmW, 18, colorTextMuted)
	}
	cx, cy, cw, ch := prestigeCancelButtonRect()
	fillRect(dst, cx, cy, cw, ch, colorButton)
	strokeRect(dst, cx, cy, cw, ch, 1, colorTextMuted)
	drawTextCentered(dst, "Cancel", cx, cy, cw, ch, colorText)
	px, py, pw, ph := prestigeConfirmButtonRect()
	fillRect(dst, px, py, pw, ph, colorBinder)
	strokeRect(dst, px, py, pw, ph, 1, colorTextMuted)
	drawTextCentered(dst, "Prestige", px, py, pw, ph, colorText)
}

func prestigeCancelButtonRect() (x, y, w, h int) {
	return prestigeConfirmX() + prestigeConfirmW/2 - prestigeConfirmButtonW - 12,
		prestigeConfirmY() + prestigeConfirmH - 62,
		prestigeConfirmButtonW,
		prestigeConfirmButtonH
}

func prestigeConfirmButtonRect() (x, y, w, h int) {
	return prestigeConfirmX() + prestigeConfirmW/2 + 12,
		prestigeConfirmY() + prestigeConfirmH - 62,
		prestigeConfirmButtonW,
		prestigeConfirmButtonH
}
