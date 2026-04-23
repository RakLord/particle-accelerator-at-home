package render

import (
	"testing"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
)

func TestSplitTileSpritesForComponent(t *testing.T) {
	t.Run("accelerator uses split sprites", func(t *testing.T) {
		split := splitTileSpritesForComponent(&components.SimpleAccelerator{Orientation: sim.DirEast})
		if split.bottom == nil || split.top == nil {
			t.Fatalf("accelerator split sprites missing: %+v", split)
		}
	})

	t.Run("rotator uses split sprites", func(t *testing.T) {
		split := splitTileSpritesForComponent(&components.Rotator{Orientation: sim.DirNorth})
		if split.bottom == nil || split.top == nil {
			t.Fatalf("rotator split sprites missing: %+v", split)
		}
	})

	t.Run("other components stay single-layer", func(t *testing.T) {
		split := splitTileSpritesForComponent(&components.MeshGrid{})
		if split.bottom != nil || split.top != nil {
			t.Fatalf("unexpected split sprites for mesh grid: %+v", split)
		}
	})
}
