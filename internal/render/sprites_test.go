package render

import (
	"testing"

	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
)

func TestSpriteLayersForComponent(t *testing.T) {
	t.Run("accelerator wraps around foreground pipe", func(t *testing.T) {
		layers := spriteLayersForComponent(&components.SimpleAccelerator{Orientation: sim.DirEast})
		if len(layers.base) != 1 || layers.base[0].image != sprites.acceleratorBottom {
			t.Fatalf("accelerator base = %+v, want acceleratorBottom", layers.base)
		}
		if len(layers.top) != 2 || layers.top[0].image != sprites.pipeHori || layers.top[1].image != sprites.acceleratorTop {
			t.Fatalf("accelerator top = %+v, want pipeHori then acceleratorTop", layers.top)
		}
	})

	t.Run("vertical accelerator uses vertical foreground pipe", func(t *testing.T) {
		layers := spriteLayersForComponent(&components.SimpleAccelerator{Orientation: sim.DirNorth})
		if len(layers.top) != 2 || layers.top[0].image != sprites.pipeVert {
			t.Fatalf("vertical accelerator top = %+v, want pipeVert first", layers.top)
		}
	})

	t.Run("horizontal mesh uses horizontal pipe and horizontal overlay", func(t *testing.T) {
		layers := spriteLayersForComponent(&components.MeshGrid{Orientation: sim.DirEast})
		if len(layers.base) != 0 {
			t.Fatalf("mesh base = %+v, want none", layers.base)
		}
		if len(layers.top) != 2 || layers.top[0].image != sprites.pipeHori || layers.top[1].image != sprites.meshGridHori {
			t.Fatalf("mesh top = %+v, want pipeHori then meshGridHori", layers.top)
		}
		if layers.top[1].rotation != 0 {
			t.Fatalf("mesh rotation = %v, want 0 (orientation baked into sprite)", layers.top[1].rotation)
		}
	})

	t.Run("vertical mesh uses vertical pipe and vertical overlay", func(t *testing.T) {
		layers := spriteLayersForComponent(&components.MeshGrid{Orientation: sim.DirNorth})
		if len(layers.top) != 2 || layers.top[0].image != sprites.pipeVert || layers.top[1].image != sprites.meshGridVert {
			t.Fatalf("vertical mesh top = %+v, want pipeVert then meshGridVert", layers.top)
		}
	})

	t.Run("rotators select directional turn assets", func(t *testing.T) {
		cases := []struct {
			name string
			dir  sim.Direction
			want any
		}{
			{name: "north", dir: sim.DirNorth, want: sprites.turnNW},
			{name: "east", dir: sim.DirEast, want: sprites.turnNE},
			{name: "south", dir: sim.DirSouth, want: sprites.turnSE},
			{name: "west", dir: sim.DirWest, want: sprites.turnSW},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				layers := spriteLayersForComponent(&components.Rotator{Orientation: tc.dir})
				if len(layers.base) != 0 {
					t.Fatalf("rotator base = %+v, want none", layers.base)
				}
				if len(layers.top) != 1 || layers.top[0].image != tc.want {
					t.Fatalf("rotator top = %+v, want %v", layers.top, tc.want)
				}
			})
		}
	})

	t.Run("wip-missing components keep fallback base sprites", func(t *testing.T) {
		for _, component := range []sim.Component{
			&components.Injector{},
		} {
			layers := spriteLayersForComponent(component)
			if len(layers.base) != 1 || layers.base[0].image == nil {
				t.Fatalf("%T fallback base missing: %+v", component, layers.base)
			}
			if len(layers.top) != 0 {
				t.Fatalf("%T top = %+v, want none", component, layers.top)
			}
		}
	})

	t.Run("horizontal magnetiser stacks top/bottom around a horizontal pipe", func(t *testing.T) {
		layers := spriteLayersForComponent(&components.Magnetiser{Orientation: sim.DirEast})
		if len(layers.base) != 1 || layers.base[0].image != sprites.magnetiserBottom {
			t.Fatalf("magnetiser base = %+v, want magnetiserBottom", layers.base)
		}
		if len(layers.top) != 2 || layers.top[0].image != sprites.pipeHori || layers.top[1].image != sprites.magnetiserTop {
			t.Fatalf("magnetiser top = %+v, want pipeHori then magnetiserTop", layers.top)
		}
		wantRotation := cardinalRotation(sim.DirEast)
		if layers.base[0].rotation != wantRotation || layers.top[1].rotation != wantRotation {
			t.Fatalf("magnetiser rotation = base:%v top:%v, want %v", layers.base[0].rotation, layers.top[1].rotation, wantRotation)
		}
	})

	t.Run("vertical magnetiser uses vertical pipe", func(t *testing.T) {
		layers := spriteLayersForComponent(&components.Magnetiser{Orientation: sim.DirNorth})
		if len(layers.top) != 2 || layers.top[0].image != sprites.pipeVert {
			t.Fatalf("vertical magnetiser top = %+v, want pipeVert first", layers.top)
		}
	})
}
