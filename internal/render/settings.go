package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/ui"
)

const (
	modalW = 360
	modalH = 440

	saveBtnW     = 160
	saveBtnH     = 40
	resetBtnW    = 160
	resetBtnH    = 40
	historyBtnW  = 200
	historyBtnH  = 40
	hotkeysBtnW  = 200
	hotkeysBtnH  = 40
	closeBtnW    = 80
	closeBtnH    = 32

	trailsRowW = 240
	trailsRowH = 32
	trailsBoxS = 20 // checkbox square side
)

func modalX() int { return (screenW - modalW) / 2 }
func modalY() int { return (screenH - modalH) / 2 }

func saveBtnX() int { return modalX() + (modalW-saveBtnW)/2 }
func saveBtnY() int { return modalY() + 56 }

func resetBtnX() int { return modalX() + (modalW-resetBtnW)/2 }
func resetBtnY() int { return modalY() + 120 }

func historyBtnX() int { return modalX() + (modalW-historyBtnW)/2 }
func historyBtnY() int { return modalY() + 184 }

func hotkeysBtnX() int { return modalX() + (modalW-hotkeysBtnW)/2 }
func hotkeysBtnY() int { return modalY() + 248 }

func closeBtnX() int { return modalX() + modalW - closeBtnW - 12 }
func closeBtnY() int { return modalY() + modalH - closeBtnH - 12 }

func trailsRowX() int { return modalX() + (modalW-trailsRowW)/2 }
func trailsRowY() int { return modalY() + 312 }

func drawSettings(dst *ebiten.Image, u *ui.UIState) {
	fillRect(dst, 0, 0, screenW, screenH, colorOverlay)

	x, y := modalX(), modalY()
	fillRect(dst, x, y, modalW, modalH, colorModalBG)
	strokeRect(dst, x, y, modalW, modalH, 2, colorTextMuted)

	drawTextCentered(dst, "Settings", x, y+12, modalW, 20, colorText)

	// Save now
	bx, by := saveBtnX(), saveBtnY()
	fillRect(dst, bx, by, saveBtnW, saveBtnH, colorButton)
	strokeRect(dst, bx, by, saveBtnW, saveBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Save now", bx, by, saveBtnW, saveBtnH, colorText)

	// Hard reset
	rx, ry := resetBtnX(), resetBtnY()
	bg := colorCollector
	label := "Hard reset"
	if u.ResetArmed {
		bg = colorResetArmed
		label = "Confirm reset?"
	}
	fillRect(dst, rx, ry, resetBtnW, resetBtnH, bg)
	strokeRect(dst, rx, ry, resetBtnW, resetBtnH, 1, colorTextMuted)
	drawTextCentered(dst, label, rx, ry, resetBtnW, resetBtnH, colorText)

	if u.LastSaveNotice != "" {
		drawTextCentered(dst, u.LastSaveNotice, x, ry+resetBtnH+8, modalW, 16, colorTextMuted)
	} else if u.ResetArmed {
		drawTextCentered(dst, "This wipes your save.", x, ry+resetBtnH+8, modalW, 16, colorTextMuted)
	}

	// Notification history
	hx, hy := historyBtnX(), historyBtnY()
	fillRect(dst, hx, hy, historyBtnW, historyBtnH, colorButton)
	strokeRect(dst, hx, hy, historyBtnW, historyBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Notification History", hx, hy, historyBtnW, historyBtnH, colorText)

	// Hotkeys
	kx, ky := hotkeysBtnX(), hotkeysBtnY()
	fillRect(dst, kx, ky, hotkeysBtnW, hotkeysBtnH, colorButton)
	strokeRect(dst, kx, ky, hotkeysBtnW, hotkeysBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Hotkeys  (/)", kx, ky, hotkeysBtnW, hotkeysBtnH, colorText)

	// Particle trails toggle (checkbox + label, toggled by click anywhere on the row).
	trx, try_ := trailsRowX(), trailsRowY()
	fillRect(dst, trx, try_, trailsRowW, trailsRowH, colorButton)
	strokeRect(dst, trx, try_, trailsRowW, trailsRowH, 1, colorTextMuted)
	boxX := trx + 8
	boxY := try_ + (trailsRowH-trailsBoxS)/2
	strokeRect(dst, boxX, boxY, trailsBoxS, trailsBoxS, 1, colorText)
	if u.TrailsEnabled {
		fillRect(dst, boxX+4, boxY+4, trailsBoxS-8, trailsBoxS-8, colorText)
	}
	drawText(dst, "Particle trails (T)", boxX+trailsBoxS+10, try_+10, colorText)

	// Close
	cx, cy := closeBtnX(), closeBtnY()
	fillRect(dst, cx, cy, closeBtnW, closeBtnH, colorButton)
	strokeRect(dst, cx, cy, closeBtnW, closeBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Close", cx, cy, closeBtnW, closeBtnH, colorText)
}
