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
	emptyTile   *ebiten.Image
	injector    *ebiten.Image
	accelerator *ebiten.Image
	meshGrid    *ebiten.Image
	magnetiser  *ebiten.Image
	elbow       *ebiten.Image
	collector   *ebiten.Image
}

var sprites = mustLoadTileSprites()

func mustLoadTileSprites() tileSprites {
	return tileSprites{
		emptyTile:   mustLoadTileSprite("images/tiles/empty_tile.png"),
		injector:    mustLoadTileSprite("images/tiles/injector.png"),
		accelerator: mustLoadTileSprite("images/tiles/accelerator.png"),
		meshGrid:    mustLoadTileSprite("images/tiles/mesh_grid.png"),
		magnetiser:  mustLoadTileSprite("images/tiles/magnetiser.png"),
		elbow:       mustLoadTileSprite("images/tiles/rotator_cw.png"),
		collector:   mustLoadTileSprite("images/tiles/collector.png"),
	}
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

func tileSpriteForComponent(c sim.Component) *ebiten.Image {
	switch v := c.(type) {
	case *components.Injector:
		return sprites.injector
	case *components.SimpleAccelerator:
		return sprites.accelerator
	case *components.MeshGrid:
		return sprites.meshGrid
	case *components.Magnetiser:
		return sprites.magnetiser
	case *components.Rotator:
		_ = v
		return sprites.elbow
	}
	return nil
}

func tileRotationForComponent(c sim.Component) float64 {
	switch v := c.(type) {
	case *components.Injector:
		return injectorRotation(v.Direction)
	case *components.SimpleAccelerator:
		return cardinalRotation(v.Orientation)
	case *components.Rotator:
		return cardinalRotation(v.Orientation)
	}
	return 0
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
		return sprites.accelerator
	case ui.ToolMeshGrid:
		return sprites.meshGrid
	case ui.ToolMagnetiser:
		return sprites.magnetiser
	case ui.ToolElbow:
		return sprites.elbow
	case ui.ToolCollector:
		return sprites.collector
	}
	return nil
}
