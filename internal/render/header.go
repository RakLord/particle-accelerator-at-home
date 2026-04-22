package render

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

const (
	settingsBtnW = 112
	settingsBtnH = 32
	settingsBtnX = screenW - settingsBtnW - 12
	settingsBtnY = (headerH - settingsBtnH) / 2
)

func drawHeader(dst *ebiten.Image, s *sim.GameState, u *ui.UIState) {
	fillRect(dst, 0, 0, screenW, headerH, colorHeaderBG)

	drawText(dst, formatUSD(s.USD), 16, (headerH-13)/2, colorText)

	if u.AutosaveError != "" {
		msg := "Save error: " + u.AutosaveError
		drawText(dst, msg, 160, (headerH-13)/2, colorResetArmed)
	}

	fillRect(dst, settingsBtnX, settingsBtnY, settingsBtnW, settingsBtnH, colorButton)
	strokeRect(dst, settingsBtnX, settingsBtnY, settingsBtnW, settingsBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Settings", settingsBtnX, settingsBtnY, settingsBtnW, settingsBtnH, colorText)
}

func formatUSD(v float64) string {
	// Simple comma-grouped integer dollars for MVP.
	whole := int64(v)
	neg := whole < 0
	if neg {
		whole = -whole
	}
	digits := fmt.Sprintf("%d", whole)
	out := make([]byte, 0, len(digits)+len(digits)/3+2)
	out = append(out, '$')
	if neg {
		out = append(out, '-')
	}
	for i, r := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(r))
	}
	return string(out)
}
