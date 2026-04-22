package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/internal/render"
	"particleaccelerator/internal/sim"
)

func main() {
	g := render.New(sim.NewGrid())
	ebiten.SetWindowSize(880, 880)
	ebiten.SetWindowTitle("Particle Accelerator @ Home")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
