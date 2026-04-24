package render

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	hotkeysModalW = 600
	hotkeysModalH = 480
	hotkeysPad    = 24
	hotkeysHeadH  = 64
	hotkeysRowH   = 30
	hotkeysKeyColW = 200
)

func hotkeysModalX() int { return (screenW - hotkeysModalW) / 2 }
func hotkeysModalY() int { return (screenH - hotkeysModalH) / 2 }

func hotkeysCloseX() int { return hotkeysModalX() + hotkeysModalW - closeBtnW - 12 }
func hotkeysCloseY() int { return hotkeysModalY() + 12 }

func hotkeysInPanel(mx, my int) bool {
	return contains(mx, my, hotkeysModalX(), hotkeysModalY(), hotkeysModalW, hotkeysModalH)
}

type hotkeyEntry struct {
	key    string
	action string
}

var hotkeyEntries = []hotkeyEntry{
	{"Esc", "Close modal or overlay"},
	{"E", "Toggle Inventory"},
	{"C", "Toggle Codex"},
	{"L", "Toggle Collection Log"},
	{"T", "Toggle particle trails"},
	{"Space", "Manual injection"},
	{"/  or  ?", "Open this Hotkeys list"},
	{"H  (over cell)", "Component help"},
	{"Q  (over cell)", "Pick tile"},
	{"Scroll  (over cell)", "Rotate component"},
	{"Left-click", "Place component"},
	{"Right-click", "Erase"},
}

func drawHotkeysModal(dst *ebiten.Image) {
	fillRect(dst, 0, 0, screenW, screenH, colorOverlay)

	x, y := hotkeysModalX(), hotkeysModalY()
	fillRect(dst, x, y, hotkeysModalW, hotkeysModalH, colorModalBG)
	strokeRect(dst, x, y, hotkeysModalW, hotkeysModalH, 2, colorTextMuted)

	drawTextFaceCentered(dst, "Hotkeys", x, y+12, hotkeysModalW, 28, fontTitle, colorText)
	drawTextCentered(dst, "Press / or ? to toggle this list", x, y+40, hotkeysModalW, 16, colorTextMuted)

	cx, cy := hotkeysCloseX(), hotkeysCloseY()
	fillRect(dst, cx, cy, closeBtnW, closeBtnH, colorButton)
	strokeRect(dst, cx, cy, closeBtnW, closeBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Close", cx, cy, closeBtnW, closeBtnH, colorText)

	rowX := x + hotkeysPad
	rowY := y + hotkeysHeadH + 12
	for i, e := range hotkeyEntries {
		ry := rowY + i*hotkeysRowH
		drawTextFace(dst, e.key, rowX, ry, fontBody, colorText)
		drawTextFace(dst, e.action, rowX+hotkeysKeyColW, ry, fontBody, colorTextMuted)
	}
}
