package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"particleaccelerator/internal/ui"
)

const (
	helperModalW = 400
	helperPad    = 18
	helperGap    = 18
	helperMinH   = 172
	helperMaxH   = 360
	helperOffset = 18
)

func helperModalRect(u *ui.UIState) (x, y, w, h int) {
	w = helperModalW
	innerW := w - 2*helperPad
	h = helperPad + 28 + 12 + wrappedHeight(u.HelperBody, innerW, fontBody) + 18 + closeBtnH + helperPad
	if h < helperMinH {
		h = helperMinH
	}
	if h > helperMaxH {
		h = helperMaxH
	}
	if u.HelperCentered {
		return (screenW - w) / 2, (screenH - h) / 2, w, h
	}
	x = u.HelperX + helperOffset
	if u.HelperX > screenW/2 {
		x = u.HelperX - w - helperOffset
	}
	y = u.HelperY + helperOffset
	if u.HelperY > screenH/2 {
		y = u.HelperY - h - helperOffset
	}
	if x < helperGap {
		x = helperGap
	}
	if y < helperGap {
		y = helperGap
	}
	if x+w > screenW-helperGap {
		x = screenW - helperGap - w
	}
	if y+h > screenH-helperGap {
		y = screenH - helperGap - h
	}
	return x, y, w, h
}

func helperCloseRect(u *ui.UIState) (x, y, w, h int) {
	mx, my, mw, mh := helperModalRect(u)
	return mx + mw - closeBtnW - 12, my + mh - closeBtnH - 12, closeBtnW, closeBtnH
}

func helperCloseHit(u *ui.UIState, mx, my int) bool {
	x, y, w, h := helperCloseRect(u)
	return contains(mx, my, x, y, w, h)
}

func drawHelperModal(dst *ebiten.Image, u *ui.UIState) {
	fillRect(dst, 0, 0, screenW, screenH, colorOverlay)

	x, y, w, h := helperModalRect(u)
	fillRect(dst, x, y, w, h, colorModalBG)
	strokeRect(dst, x, y, w, h, 2, colorHeaderIncome)

	innerX := x + helperPad
	innerW := w - 2*helperPad
	cur := y + helperPad
	drawTextFace(dst, u.HelperHeader, innerX, cur, fontTitle, colorText)
	cur += 40

	cx, cy, cw, ch := helperCloseRect(u)
	bodyH := cy - 12 - cur
	clipDrawWrapped(dst, u.HelperBody, innerX, cur, innerW, bodyH, fontBody, colorText)

	fillRect(dst, cx, cy, cw, ch, colorButton)
	strokeRect(dst, cx, cy, cw, ch, 1, colorTextMuted)
	drawTextCentered(dst, "Close", cx, cy, cw, ch, colorText)
}

func clipDrawWrapped(dst *ebiten.Image, body string, x, y, w, h int, face text.Face, col color.Color) {
	lines := wrapLines(body, w, face)
	lh := clippedLineHeight(face)
	for i, line := range lines {
		ly := y + i*lh
		if ly+lh > y+h {
			break
		}
		drawTextFace(dst, line, x, ly, face, col)
	}
}

func clippedLineHeight(face text.Face) int {
	lh := face.Metrics().HLineGap + face.Metrics().HAscent + face.Metrics().HDescent
	if lh == 0 {
		_, mh := text.Measure("Mg", face, 0)
		lh = mh
	}
	return int(math.Ceil(lh))
}
