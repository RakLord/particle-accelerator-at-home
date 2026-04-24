package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/sim"
)

const (
	logModalW = 960
	logModalH = 520
	logPad    = 24
	logRowH   = 34
	logHeadH  = 28
)

func logModalX() int { return (screenW - logModalW) / 2 }
func logModalY() int { return (screenH - logModalH) / 2 }

func logCloseX() int { return logModalX() + logModalW - closeBtnW - 12 }
func logCloseY() int { return logModalY() + 12 }

func logInPanel(mx, my int) bool {
	return contains(mx, my, logModalX(), logModalY(), logModalW, logModalH)
}

func drawCollectionLog(dst *ebiten.Image, s *sim.GameState) {
	fillRect(dst, 0, 0, screenW, screenH, colorOverlay)

	x, y := logModalX(), logModalY()
	fillRect(dst, x, y, logModalW, logModalH, colorModalBG)
	strokeRect(dst, x, y, logModalW, logModalH, 2, colorTextMuted)

	drawTextFaceCentered(dst, "Collection Log", x, y+12, logModalW, 28, fontTitle, colorText)
	drawTextCentered(dst, "Recent 10 collected Subjects · newest first", x, y+40, logModalW, 16, colorTextMuted)

	cx, cy := logCloseX(), logCloseY()
	fillRect(dst, cx, cy, closeBtnW, closeBtnH, colorButton)
	strokeRect(dst, cx, cy, closeBtnW, closeBtnH, 1, colorTextMuted)
	drawTextCentered(dst, "Close", cx, cy, closeBtnW, closeBtnH, colorText)

	if len(s.CollectionLog) == 0 {
		drawTextCentered(dst, "No Subjects collected yet.", x, y+logModalH/2-10, logModalW, 20, colorTextMuted)
		return
	}

	listX := x + logPad
	listY := y + 86
	listW := logModalW - 2*logPad
	drawLogHeader(dst, listX, listY, listW)
	for i, entry := range s.CollectionLog {
		if i >= sim.MaxCollectionLogEntries {
			break
		}
		drawLogRow(dst, entry, i, listX, listY+logHeadH+i*logRowH, listW)
	}
}

func drawLogHeader(dst *ebiten.Image, x, y, w int) {
	fillRect(dst, x, y, w, logHeadH, colorButton)
	strokeRect(dst, x, y, w, logHeadH, 1, colorTextMuted)
	drawTextSmall(dst, "Tick", x+10, y+8, colorTextMuted)
	drawTextSmall(dst, "Element", x+100, y+8, colorTextMuted)
	drawTextSmall(dst, "Mass", x+230, y+8, colorTextMuted)
	drawTextSmall(dst, "Speed", x+350, y+8, colorTextMuted)
	drawTextSmall(dst, "Magnetism", x+450, y+8, colorTextMuted)
	drawTextSmall(dst, "Research", x+590, y+8, colorTextMuted)
	drawTextSmall(dst, "Value", x+720, y+8, colorTextMuted)
}

func drawLogRow(dst *ebiten.Image, entry sim.CollectionLogEntry, index, x, y, w int) {
	bg := color.RGBA{0x12, 0x12, 0x20, 0xf0}
	if index%2 == 1 {
		bg = color.RGBA{0x18, 0x18, 0x28, 0xf0}
	}
	fillRect(dst, x, y, w, logRowH, bg)
	strokeRect(dst, x, y, w, logRowH, 1, colorGridLine)

	info := sim.ElementCatalog[entry.Element]
	elementName := info.Symbol + " " + info.Name
	drawTextSmall(dst, itoa(int(entry.Tick)), x+10, y+10, colorTextMuted)
	drawTextSmall(dst, elementName, x+100, y+10, elementAccentColor(entry.Element))
	drawTextSmall(dst, formatNumber(entry.Mass), x+230, y+10, colorText)
	drawTextSmall(dst, itoa(entry.Speed), x+350, y+10, colorText)
	drawTextSmall(dst, formatNumber(entry.Magnetism), x+450, y+10, colorText)
	drawTextSmall(dst, itoa(entry.ResearchLevel), x+590, y+10, colorText)
	drawTextSmall(dst, formatUSD(entry.Value), x+720, y+10, colorText)
}
