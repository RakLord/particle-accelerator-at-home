package render

import (
	"bytes"
	"fmt"
	"image/png"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"particleaccelerator/assets"
	"particleaccelerator/internal/sim"
	"particleaccelerator/internal/sim/components"
	"particleaccelerator/internal/ui"
)

type tileSprites struct {
	emptyTile *ebiten.Image

	pipeHori *ebiten.Image
	pipeVert *ebiten.Image
	turnNE   *ebiten.Image
	turnNW   *ebiten.Image
	turnSE   *ebiten.Image
	turnSW   *ebiten.Image

	acceleratorBottom *ebiten.Image
	acceleratorTop    *ebiten.Image
	meshGridTop       *ebiten.Image

	injector   *ebiten.Image
	magnetiser *ebiten.Image
	collector  *ebiten.Image
}

var sprites = mustLoadTileSprites()

func mustLoadTileSprites() tileSprites {
	return tileSprites{
		emptyTile: mustLoadTileSprite("images/tiles/empty_tile.png"),

		pipeHori: mustLoadTileSprite("images/tiles/pipe_hori.png"),
		pipeVert: mustLoadTileSprite("images/tiles/pipe_vert.png"),
		turnNE:   mustLoadTileSprite("images/tiles/turn_ne.png"),
		turnNW:   mustLoadTileSprite("images/tiles/turn_nw.png"),
		turnSE:   mustLoadTileSprite("images/tiles/turn_se.png"),
		turnSW:   mustLoadTileSprite("images/tiles/turn_sw.png"),

		acceleratorBottom: mustLoadTileSprite("images/tiles/accelerator_bottom.png"),
		acceleratorTop:    mustLoadTileSprite("images/tiles/accelerator_top.png"),
		meshGridTop:       mustLoadTileSprite("images/tiles/mesh_grid_top.png"),

		injector:   mustLoadTileSprite("images/tiles/injector.png"),
		magnetiser: mustLoadTileSprite("images/tiles/magnetiser.png"),
		collector:  mustLoadTileSprite("images/tiles/collector.png"),
	}
}

type spriteLayer struct {
	image    *ebiten.Image
	rotation float64
}

type componentSpriteLayers struct {
	base []spriteLayer
	top  []spriteLayer
}

func spriteLayersForComponent(c sim.Component) componentSpriteLayers {
	switch v := c.(type) {
	case *components.Injector:
		return componentSpriteLayers{base: []spriteLayer{{image: sprites.injector, rotation: injectorRotation(v.Direction)}}}
	case *components.SimpleAccelerator:
		rotation := cardinalRotation(v.Orientation)
		return componentSpriteLayers{
			base: []spriteLayer{
				{image: sprites.acceleratorBottom, rotation: rotation},
			},
			top: []spriteLayer{
				{image: pipeSpriteForOrientation(v.Orientation)},
				{image: sprites.acceleratorTop, rotation: rotation},
			},
		}
	case *components.MeshGrid:
		rotation := cardinalRotation(v.Orientation)
		return componentSpriteLayers{
			top: []spriteLayer{
				{image: pipeSpriteForOrientation(v.Orientation)},
				{image: sprites.meshGridTop, rotation: rotation},
			},
		}
	case *components.Magnetiser:
		return componentSpriteLayers{base: []spriteLayer{{image: sprites.magnetiser}}}
	case *components.Rotator:
		return componentSpriteLayers{top: []spriteLayer{{image: turnSpriteForOrientation(v.Orientation)}}}
	}
	return componentSpriteLayers{}
}

func pipeSpriteForOrientation(d sim.Direction) *ebiten.Image {
	if d == sim.DirNorth || d == sim.DirSouth {
		return sprites.pipeVert
	}
	return sprites.pipeHori
}

func turnSpriteForOrientation(d sim.Direction) *ebiten.Image {
	switch d {
	case sim.DirNorth:
		return sprites.turnNW
	case sim.DirEast:
		return sprites.turnNE
	case sim.DirSouth:
		return sprites.turnSE
	case sim.DirWest:
		return sprites.turnSW
	}
	return sprites.turnNW
}

func mustLoadTileSprite(path string) *ebiten.Image {
	b, err := assets.TileFS.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("render: read tile sprite %q: %v", path, err))
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		panic(fmt.Sprintf("render: decode tile sprite %q: %v", path, err))
	}
	return ebiten.NewImageFromImage(img)
}

func drawSpriteFitted(dst, img *ebiten.Image, x, y, w, h int) {
	if img == nil {
		return
	}
	b := img.Bounds()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/float64(b.Dx()), float64(h)/float64(b.Dy()))
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(img, op)
}

func drawSpriteCenteredRotated(dst, img *ebiten.Image, x, y, w, h int, angle float64) {
	if img == nil {
		return
	}
	b := img.Bounds()
	scaleX := float64(w) / float64(b.Dx())
	scaleY := float64(h) / float64(b.Dy())
	cx := float64(b.Dx()) / 2
	cy := float64(b.Dy()) / 2
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-cx, -cy)
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Rotate(angle)
	op.GeoM.Translate(float64(x)+float64(w)/2, float64(y)+float64(h)/2)
	dst.DrawImage(img, op)
}

func injectorRotation(d sim.Direction) float64 {
	switch d {
	case sim.DirNorth:
		return -math.Pi / 2
	case sim.DirSouth:
		return math.Pi / 2
	case sim.DirWest:
		return math.Pi
	default:
		return 0
	}
}

func cardinalRotation(d sim.Direction) float64 {
	switch d {
	case sim.DirNorth:
		return math.Pi / 2
	case sim.DirEast:
		return math.Pi
	case sim.DirSouth:
		return 3 * math.Pi / 2
	case sim.DirWest:
		return 0
	default:
		return math.Pi / 2
	}
}

func tileSpriteForTool(t ui.Tool) *ebiten.Image {
	switch t {
	case ui.ToolInjectorHydrogen, ui.ToolInjectorHelium:
		return sprites.injector
	case ui.ToolAccelerator:
		return sprites.acceleratorTop
	case ui.ToolMeshGrid:
		return sprites.meshGridTop
	case ui.ToolMagnetiser:
		return sprites.magnetiser
	case ui.ToolElbow:
		return sprites.turnNE
	case ui.ToolCollector:
		return sprites.collector
	}
	return nil
}
