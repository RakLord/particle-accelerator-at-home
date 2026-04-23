package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

const (
	codexTableCols    = 18
	codexTableRows    = 7
	codexTileSize     = 54
	codexTileGap      = 6
	codexGroupLabelH  = 16
	codexPeriodLabelW = 20

	codexCardW      = 380
	codexCardH      = 334
	codexUnlockBtnW = 220
	codexUnlockBtnH = 40
	codexCardStatsX = 28
	codexCardValueX = 230
	codexCardStatsY = 156
	codexCardRowGap = 26
	codexNoticeH    = 18
)

func codexPanelX() int { return 24 }
func codexPanelY() int { return 56 }
func codexPanelW() int { return screenW - 48 }
func codexPanelH() int { return screenH - codexPanelY() - 24 }

func codexCloseX() int { return codexPanelX() + codexPanelW() - closeBtnW - 16 }
func codexCloseY() int { return codexPanelY() + 16 }

func codexTableW() int {
	return codexPeriodLabelW + codexTableCols*codexTileSize + (codexTableCols-1)*codexTileGap
}

func codexTableH() int {
	return codexGroupLabelH + codexTableRows*codexTileSize + (codexTableRows-1)*codexTileGap
}

func codexTableX() int { return codexPanelX() + (codexPanelW()-codexTableW())/2 }
func codexTableY() int { return codexPanelY() + 84 }

func codexGridX() int { return codexTableX() + codexPeriodLabelW }
func codexGridY() int { return codexTableY() + codexGroupLabelH }

func codexSlotRect(period, group int) (x, y, w, h int) {
	return codexGridX() + (group-1)*(codexTileSize+codexTileGap),
		codexGridY() + (period-1)*(codexTileSize+codexTileGap),
		codexTileSize, codexTileSize
}

func codexTileRect(e sim.Element) (x, y, w, h int) {
	info := sim.ElementCatalog[e]
	return codexSlotRect(info.Period, info.Group)
}

func codexCardX() int { return codexPanelX() + (codexPanelW()-codexCardW)/2 }
func codexCardY() int { return codexPanelY() + (codexPanelH()-codexCardH)/2 + 28 }

func codexUnlockButtonRect() (x, y, w, h int) {
	return codexCardX() + (codexCardW-codexUnlockBtnW)/2,
		codexCardY() + codexCardH - codexUnlockBtnH - 24,
		codexUnlockBtnW, codexUnlockBtnH
}

func codexElementAt(mx, my int) (sim.Element, bool) {
	for _, e := range sim.CatalogOrder {
		x, y, w, h := codexTileRect(e)
		if contains(mx, my, x, y, w, h) {
			return e, true
		}
	}
	return "", false
}

func codexFocusedElement(hovered, pinned sim.Element) sim.Element {
	if pinned != "" {
		return pinned
	}
	return hovered
}

func drawPeriodicTable(dst *ebiten.Image, s *sim.GameState, u *ui.UIState, focused sim.Element) {
	fillRect(dst, 0, 0, screenW, screenH, colorOverlay)

	px, py, pw, ph := codexPanelX(), codexPanelY(), codexPanelW(), codexPanelH()
	fillRect(dst, px, py, pw, ph, colorModalBG)
	strokeRect(dst, px, py, pw, ph, 2, colorTextMuted)

	drawTextFaceCentered(dst, "Periodic Table", px, py+14, pw, 28, fontTitle, colorText)
	if u.CodexNotice != "" {
		drawTextCentered(dst, u.CodexNotice, px, py+46, pw, codexNoticeH, colorTextMuted)
	}

	drawCodexTableFrame(dst)
	drawCodexTiles(dst, s, focused)

	if focused != "" {
		drawCodexCard(dst, s, focused)
	} else {
		drawTextCentered(dst, "Hover or select an element to inspect it.", px, codexCardY()+codexCardH/2-10, pw, 20, colorTextMuted)
	}

	cx, cy := codexCloseX(), codexCloseY()
	fillRect(dst, cx, cy, closeBtnW, closeBtnH, colorButton)
	strokeRect(dst, cx, cy, closeBtnW, closeBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Close", cx, cy, closeBtnW, closeBtnH, colorText)
}

func drawCodexTableFrame(dst *ebiten.Image) {
	labelY := codexTableY()
	for group := 1; group <= codexTableCols; group++ {
		x, _, w, _ := codexSlotRect(1, group)
		drawTextSmall(dst, itoa(group), x+(w/2)-4, labelY, colorTextMuted)
	}
	for period := 1; period <= codexTableRows; period++ {
		_, y, _, h := codexSlotRect(period, 1)
		drawTextSmall(dst, itoa(period), codexTableX(), y+(h/2)-6, colorTextMuted)
	}

	emptyBG := color.RGBA{0x12, 0x12, 0x20, 0xd0}
	for period := 1; period <= codexTableRows; period++ {
		for group := 1; group <= codexTableCols; group++ {
			x, y, w, h := codexSlotRect(period, group)
			fillRect(dst, x, y, w, h, emptyBG)
			strokeRect(dst, x, y, w, h, 1, colorGridLine)
		}
	}
}

func drawCodexTiles(dst *ebiten.Image, s *sim.GameState, focused sim.Element) {
	for _, e := range sim.CatalogOrder {
		info := sim.ElementCatalog[e]
		x, y, w, h := codexTileRect(e)
		bg, border := codexTileColors(s, e, focused == e)
		fillRect(dst, x, y, w, h, bg)
		strokeRect(dst, x, y, w, h, 2, border)

		drawTextSmall(dst, itoa(info.AtomicNumber), x+6, y+6, colorTextMuted)
		drawTextFaceCentered(dst, info.Symbol, x, y+10, w, 28, fontTitle, elementAccentColor(e))
		if sim.IsElementUnlocked(s, e) {
			drawTextSmall(dst, "LIVE", x+6, y+h-16, colorPurchaseActive)
		} else if sim.IsElementPurchasable(s, e) {
			drawTextSmall(dst, "READY", x+6, y+h-16, colorText)
		} else {
			drawTextSmall(dst, "LOCKED", x+6, y+h-16, colorTextMuted)
		}
	}
}

func codexTileColors(s *sim.GameState, e sim.Element, focused bool) (bg, border color.Color) {
	switch {
	case sim.IsElementUnlocked(s, e):
		bg = color.RGBA{0x1f, 0x3b, 0x28, 0xff}
	case sim.IsElementPurchasable(s, e):
		bg = color.RGBA{0x22, 0x2d, 0x54, 0xff}
	default:
		bg = colorButton
	}
	border = colorTextMuted
	if focused {
		border = colorText
	}
	return bg, border
}

func drawCodexCard(dst *ebiten.Image, s *sim.GameState, e sim.Element) {
	info := sim.ElementCatalog[e]
	x, y := codexCardX(), codexCardY()
	cardBG := color.RGBA{0x10, 0x10, 0x20, 0xf8}
	fillRect(dst, x, y, codexCardW, codexCardH, cardBG)
	strokeRect(dst, x, y, codexCardW, codexCardH, 2, colorText)

	statusText, statusColor := codexStatusLabel(s, e)
	drawTextSmall(dst, "#"+itoa(info.AtomicNumber), x+18, y+18, colorTextMuted)
	stw, _ := measureTextSmall(statusText)
	drawTextSmall(dst, statusText, x+codexCardW-stw-18, y+18, statusColor)

	drawTextFaceCentered(dst, info.Symbol, x, y+24, codexCardW, 56, fontDisplay, elementAccentColor(e))
	drawTextFaceCentered(dst, info.Name, x, y+88, codexCardW, 24, fontTitle, colorText)
	drawTextFaceCentered(dst, "Period "+itoa(info.Period)+" · Group "+itoa(info.Group), x, y+118, codexCardW, 18, fontSmall, colorTextMuted)

	drawCodexStatRow(dst, x+codexCardStatsX, y+codexCardStatsY+0*codexCardRowGap, "Research", itoa(s.Research[e]))
	drawCodexStatRow(dst, x+codexCardStatsX, y+codexCardStatsY+1*codexCardRowGap, "Multiplier", formatMultiplier(codexEffectiveMultiplier(s, e)))

	stats := s.BestStats[e]
	if stats.MaxSpeed == 0 && stats.MaxMass.IsZero() && stats.MaxCollectedValue.IsZero() {
		drawTextCentered(dst, "No codex records yet.", x, y+codexCardStatsY+2*codexCardRowGap+8, codexCardW, 18, colorTextMuted)
	} else {
		drawCodexStatRow(dst, x+codexCardStatsX, y+codexCardStatsY+2*codexCardRowGap, "Max Speed", itoa(stats.MaxSpeed))
		drawCodexStatRow(dst, x+codexCardStatsX, y+codexCardStatsY+3*codexCardRowGap, "Max Mass", formatNumber(stats.MaxMass))
		drawCodexStatRow(dst, x+codexCardStatsX, y+codexCardStatsY+4*codexCardRowGap, "Best Value", formatUSD(stats.MaxCollectedValue))
	}

	if sim.IsElementUnlocked(s, e) {
		drawTextCentered(dst, "Unlocked in palette", x, y+codexCardH-60, codexCardW, 18, colorPurchaseActive)
		return
	}
	if sim.IsElementPurchasable(s, e) {
		bx, by, bw, bh := codexUnlockButtonRect()
		fillRect(dst, bx, by, bw, bh, colorPurchaseActive)
		strokeRect(dst, bx, by, bw, bh, 1, colorTextMuted)
		drawTextCentered(dst, "Unlock for "+formatUSD(info.UnlockCost), bx, by, bw, bh, colorText)
		return
	}
	need := info.ResearchThreshold - s.Research[info.UnlocksFrom]
	if need < 0 {
		need = 0
	}
	drawTextCentered(dst, "Locked · need "+itoa(need)+" more "+sim.ElementCatalog[info.UnlocksFrom].Symbol+" research", x+20, y+codexCardH-64, codexCardW-40, 18, colorTextMuted)
}

func drawCodexStatRow(dst *ebiten.Image, x, y int, label, value string) {
	drawText(dst, label, x, y, colorTextMuted)
	drawText(dst, value, x+codexCardValueX, y, colorText)
}

func codexEffectiveMultiplier(s *sim.GameState, e sim.Element) bignum.Decimal {
	research := s.Research[e]
	info := sim.ElementCatalog[e]
	return info.Multiplier.Mul(bignum.One().Add(bignum.FromInt(research).Div(sim.ResearchK)))
}

func codexStatusLabel(s *sim.GameState, e sim.Element) (string, color.Color) {
	switch {
	case sim.IsElementUnlocked(s, e):
		return "Unlocked", colorPurchaseActive
	case sim.IsElementPurchasable(s, e):
		return "Ready to unlock", colorText
	default:
		return "Research-locked", colorTextMuted
	}
}

func elementAccentColor(e sim.Element) color.Color {
	return subjectColor(e)
}
