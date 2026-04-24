package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

const (
	notifModalW = 900
	notifModalH = 560
	notifPad    = 24
	notifRowH   = 68
	notifHeadH  = 64
)

func notifModalX() int { return (screenW - notifModalW) / 2 }
func notifModalY() int { return (screenH - notifModalH) / 2 }

func notifCloseX() int { return notifModalX() + notifModalW - closeBtnW - 12 }
func notifCloseY() int { return notifModalY() + 12 }

func notifInPanel(mx, my int) bool {
	return contains(mx, my, notifModalX(), notifModalY(), notifModalW, notifModalH)
}

func notifVisibleRows() int {
	return (notifModalH - notifHeadH - notifPad*2) / notifRowH
}

func drawNotificationHistory(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	fillRect(dst, 0, 0, screenW, screenH, colorOverlay)

	x, y := notifModalX(), notifModalY()
	fillRect(dst, x, y, notifModalW, notifModalH, colorModalBG)
	strokeRect(dst, x, y, notifModalW, notifModalH, 2, colorTextMuted)

	drawTextFaceCentered(dst, "Notification History", x, y+12, notifModalW, 28, fontTitle, colorText)
	drawTextCentered(dst, "Logged helper notifications · newest first", x, y+40, notifModalW, 16, colorTextMuted)

	cx, cy := notifCloseX(), notifCloseY()
	fillRect(dst, cx, cy, closeBtnW, closeBtnH, colorButton)
	strokeRect(dst, cx, cy, closeBtnW, closeBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Close", cx, cy, closeBtnW, closeBtnH, colorText)

	if len(s.NotificationLog) == 0 {
		drawTextCentered(dst, "No notifications yet.", x, y+notifModalH/2-10, notifModalW, 20, colorTextMuted)
		return
	}

	maxScroll := maxNotificationScroll(s)
	scroll := u.NotificationHistoryScroll
	if scroll > maxScroll {
		scroll = maxScroll
	}
	if scroll < 0 {
		scroll = 0
	}

	listX := x + notifPad
	listY := y + notifHeadH + 18
	listW := notifModalW - 2*notifPad
	visible := notifVisibleRows()
	for row := 0; row < visible; row++ {
		idx := scroll + row
		if idx >= len(s.NotificationLog) {
			break
		}
		drawNotificationRow(dst, s.NotificationLog[idx], row, listX, listY+row*notifRowH, listW)
	}

	if maxScroll > 0 {
		label := "Scroll " + itoa(scroll+1) + "/" + itoa(maxScroll+1)
		lw, _ := measureTextSmall(label)
		drawTextSmall(dst, label, x+notifModalW-notifPad-lw, y+notifModalH-18, colorTextMuted)
	}
}

func drawNotificationRow(dst *ebiten.Image, entry sim.NotificationEntry, index, x, y, w int) {
	bg := color.RGBA{0x12, 0x12, 0x20, 0xf0}
	if index%2 == 1 {
		bg = color.RGBA{0x18, 0x18, 0x28, 0xf0}
	}
	fillRect(dst, x, y, w, notifRowH-6, bg)
	strokeRect(dst, x, y, w, notifRowH-6, 1, colorGridLine)

	timeText := entry.TimeHHMM
	if timeText == "" {
		timeText = "--:--"
	}
	drawTextSmall(dst, timeText, x+12, y+12, colorTextMuted)
	drawTextFace(dst, entry.Header, x+76, y+10, fontBody, colorText)
	clipDrawWrapped(dst, entry.Body, x+76, y+32, w-96, 26, fontSmall, colorTextMuted)
}

func maxNotificationScroll(s *sim.GameState) int {
	maxScroll := len(s.NotificationLog) - notifVisibleRows()
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}
