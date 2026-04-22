package render

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

const (
	codexModalW = 640
	codexModalH = 440

	codexRowH  = 60
	codexRowGap = 8
	codexHeaderOffsetY = 52

	codexActionW = 180
	codexActionH = 36
)

func codexModalX() int { return (screenW - codexModalW) / 2 }
func codexModalY() int { return (screenH - codexModalH) / 2 }

func codexCloseX() int { return codexModalX() + codexModalW - closeBtnW - 12 }
func codexCloseY() int { return codexModalY() + codexModalH - closeBtnH - 12 }

func codexRowY(i int) int {
	return codexModalY() + codexHeaderOffsetY + 36 + i*(codexRowH+codexRowGap)
}

func codexActionRect(i int) (x, y, w, h int) {
	return codexModalX() + codexModalW - codexActionW - 20,
		codexRowY(i) + (codexRowH-codexActionH)/2,
		codexActionW, codexActionH
}

// codexActionAt returns the Element whose action button contains (mx, my), or
// ("", false) if none.
func codexActionAt(s *sim.GameState, mx, my int) (sim.Element, bool) {
	for i, e := range sim.CatalogOrder {
		if !sim.IsElementPurchasable(s, e) {
			continue
		}
		x, y, w, h := codexActionRect(i)
		if contains(mx, my, x, y, w, h) {
			return e, true
		}
	}
	return "", false
}

func drawPeriodicTable(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	fillRect(dst, 0, 0, screenW, screenH, colorOverlay)

	mx, my := codexModalX(), codexModalY()
	fillRect(dst, mx, my, codexModalW, codexModalH, colorModalBG)
	strokeRect(dst, mx, my, codexModalW, codexModalH, 2, colorTextMuted)

	drawTextCentered(dst, "Periodic Table", mx, my+12, codexModalW, 20, colorText)

	// Column headers.
	headerY := my + codexHeaderOffsetY + 12
	drawText(dst, "Element", mx+20, headerY, colorTextMuted)
	drawText(dst, "Research", mx+200, headerY, colorTextMuted)
	drawText(dst, "Multiplier", mx+320, headerY, colorTextMuted)
	drawText(dst, "Status", mx+440, headerY, colorTextMuted)

	for i, e := range sim.CatalogOrder {
		info := sim.ElementCatalog[e]
		y := codexRowY(i)

		rowBG := colorButton
		fillRect(dst, mx+12, y, codexModalW-24, codexRowH, rowBG)
		strokeRect(dst, mx+12, y, codexModalW-24, codexRowH, 1, colorTextMuted)

		// Element name + symbol.
		drawText(dst, info.Symbol+"  "+info.Name, mx+20, y+22, colorText)

		// Research.
		drawText(dst, fmt.Sprintf("%d", s.Research[e]), mx+200, y+22, colorText)

		// Effective multiplier (base × research bonus).
		research := s.Research[e]
		effective := info.Multiplier * (1 + float64(research)/sim.ResearchK)
		drawText(dst, fmt.Sprintf("×%.2f", effective), mx+320, y+22, colorText)

		// Status / action column.
		ax, ay, aw, ah := codexActionRect(i)
		switch {
		case sim.IsElementUnlocked(s, e):
			drawText(dst, "Unlocked", mx+440, y+22, colorPurchaseActive)
		case sim.IsElementPurchasable(s, e):
			fillRect(dst, ax, ay, aw, ah, colorPurchaseActive)
			strokeRect(dst, ax, ay, aw, ah, 1, colorTextMuted)
			drawTextCentered(dst, fmt.Sprintf("Unlock for $%.0f", info.UnlockCost), ax, ay, aw, ah, colorText)
		default:
			need := info.ResearchThreshold - s.Research[info.UnlocksFrom]
			drawText(dst, fmt.Sprintf("Locked · %d more %s research", need, sim.ElementCatalog[info.UnlocksFrom].Symbol), mx+440, y+22, colorTextMuted)
		}
	}

	if u.CodexNotice != "" {
		drawTextCentered(dst, u.CodexNotice, mx, codexCloseY()-24, codexModalW, 16, colorTextMuted)
	}

	// Close button.
	cx, cy := codexCloseX(), codexCloseY()
	fillRect(dst, cx, cy, closeBtnW, closeBtnH, colorButton)
	strokeRect(dst, cx, cy, closeBtnW, closeBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Close", cx, cy, closeBtnW, closeBtnH, colorText)
}

