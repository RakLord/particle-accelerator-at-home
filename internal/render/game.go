package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"particleaccelerator/internal/sim"
)

const (
	cellSize    = 64
	gridPadding = 48
)

type Game struct {
	grid *sim.Grid
}

func New(g *sim.Grid) *Game { return &Game{grid: g} }

func (g *Game) Update() error {
	g.grid.Tick()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x0a, 0x0a, 0x14, 0xff})

	line := color.RGBA{0x33, 0x33, 0x5a, 0xff}
	for i := 0; i <= sim.GridSize; i++ {
		p := float32(gridPadding + i*cellSize)
		a := float32(gridPadding)
		b := float32(gridPadding + sim.GridSize*cellSize)
		vector.StrokeLine(screen, p, a, p, b, 1, line, false)
		vector.StrokeLine(screen, a, p, b, p, 1, line, false)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	side := gridPadding*2 + cellSize*sim.GridSize
	return side, side
}
