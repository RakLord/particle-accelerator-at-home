package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/render"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/ui"
)

func main() {
	state := loadOrNew()
	uiState := ui.NewUIState()

	save := func() error { return state.Save() }
	g := render.New(state, uiState, save, save)

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Particle Accelerator @ Home")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(state.TickRate)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func loadOrNew() *sim.GameState {
	if s, ok, err := sim.Load(); err == nil && ok {
		return s
	} else if err != nil {
		log.Printf("save load failed, starting fresh: %v", err)
	}
	return sim.NewGameState()
}
