package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/bignum"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

const (
	settingsBtnW = 112
	settingsBtnH = 32
	settingsBtnX = screenW - settingsBtnW - 12
	settingsBtnY = (headerH - settingsBtnH) / 2

	codexBtnW = 112
	codexBtnH = 32
	codexBtnX = settingsBtnX - codexBtnW - 8
	codexBtnY = settingsBtnY
)

func drawHeader(dst *ebiten.Image, s *sim.GameState, u *ui.UIState, incomePerSecond bignum.Decimal) {
	fillRect(dst, 0, 0, screenW, headerH, colorHeaderBG)

	usdText := formatUSD(s.USD)
	usdX := 16
	_, usdH := measureText(usdText)
	usdY := (headerH - usdH) / 2
	drawText(dst, usdText, usdX, usdY, colorText)
	usdW, _ := measureText(usdText)

	rateText := formatIncomeRate(incomePerSecond)
	rateX := usdX + usdW + 12
	_, rateH := measureTextSmall(rateText)
	rateY := (headerH - rateH) / 2
	drawTextSmall(dst, rateText, rateX, rateY, colorHeaderIncome)
	rateW, _ := measureTextSmall(rateText)

	if u.AutosaveError != "" {
		msg := "Save error: " + u.AutosaveError
		drawText(dst, msg, rateX+rateW+24, usdY, colorResetArmed)
	}

	fillRect(dst, codexBtnX, codexBtnY, codexBtnW, codexBtnH, colorButton)
	strokeRect(dst, codexBtnX, codexBtnY, codexBtnW, codexBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Codex", codexBtnX, codexBtnY, codexBtnW, codexBtnH, colorText)

	fillRect(dst, settingsBtnX, settingsBtnY, settingsBtnW, settingsBtnH, colorButton)
	strokeRect(dst, settingsBtnX, settingsBtnY, settingsBtnW, settingsBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Settings", settingsBtnX, settingsBtnY, settingsBtnW, settingsBtnH, colorText)
}
