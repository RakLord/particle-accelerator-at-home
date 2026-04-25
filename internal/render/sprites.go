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
	meshGridHori      *ebiten.Image
	meshGridVert      *ebiten.Image

	injector         *ebiten.Image
	magnetiserTop    *ebiten.Image
	magnetiserBottom *ebiten.Image
	collector        *ebiten.Image
	compressorHori   *ebiten.Image
	compressorVert   *ebiten.Image
	binder           *ebiten.Image

	accelLogo    *ebiten.Image
	meshLogo     *ebiten.Image
	magnetLogo   *ebiten.Image
	pipeLogo     *ebiten.Image
	turnLogo     *ebiten.Image
	injectorLogo *ebiten.Image
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
		meshGridHori:      mustLoadTileSprite("images/tiles/mesh_grid_hori.png"),
		meshGridVert:      mustLoadTileSprite("images/tiles/mesh_grid_vert.png"),

		injector:         mustLoadTileSprite("images/tiles/injector.png"),
		magnetiserTop:    mustLoadTileSprite("images/tiles/magnetiser_top.png"),
		magnetiserBottom: mustLoadTileSprite("images/tiles/magnetiser_bottom.png"),
		collector:        mustLoadTileSprite("images/tiles/collector.png"),
		compressorHori:   mustLoadTileSprite("images/tiles/compressor_hori.png"),
		compressorVert:   mustLoadTileSprite("images/tiles/compressor_vert.png"),
		binder:           mustLoadTileSprite("images/tiles/binder.png"),

		accelLogo:    mustLoadTileSprite("images/tiles/accelerator_logo.png"),
		meshLogo:     mustLoadTileSprite("images/tiles/mesh_grid_logo.png"),
		magnetLogo:   mustLoadTileSprite("images/tiles/magnetiser_logo.png"),
		pipeLogo:     mustLoadTileSprite("images/tiles/pipe_logo.png"),
		turnLogo:     mustLoadTileSprite("images/tiles/turn_logo.png"),
		injectorLogo: mustLoadTileSprite("images/tiles/injector_logo.png"),
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
		return componentSpriteLayers{
			top: []spriteLayer{
				{image: pipeSpriteForOrientation(v.Orientation)},
				{image: meshGridSpriteForOrientation(v.Orientation)},
			},
		}
	case *components.Magnetiser:
		rotation := cardinalRotation(v.Orientation)
		return componentSpriteLayers{
			base: []spriteLayer{
				{image: sprites.magnetiserBottom, rotation: rotation},
			},
			top: []spriteLayer{
				{image: pipeSpriteForOrientation(v.Orientation)},
				{image: sprites.magnetiserTop, rotation: rotation},
			},
		}
	case *components.Rotator:
		return componentSpriteLayers{top: []spriteLayer{{image: turnSpriteForOrientation(v.Orientation)}}}
	case *components.Pipe:
		return componentSpriteLayers{top: []spriteLayer{{image: pipeSpriteForOrientation(v.Orientation)}}}
	case *components.Resonator:
		// Placeholder: reuses the magnetiser bottom sprite until dedicated art lands.
		return componentSpriteLayers{base: []spriteLayer{{image: sprites.magnetiserBottom}}}
	case *components.Catalyst:
		// Placeholder: reuses injector sprite centered (no rotation).
		return componentSpriteLayers{base: []spriteLayer{{image: sprites.injector}}}
	case *components.Duplicator:
		// Placeholder: a pipe orthogonal to the input plus a turn tile on the
		// input side to hint the T-junction shape. Real art pending.
		return componentSpriteLayers{
			top: []spriteLayer{
				{image: pipeSpriteForOrientation(perpendicular(v.Orientation))},
				{image: turnSpriteForOrientation(v.Orientation)},
			},
		}
	case *components.Compressor:
		return componentSpriteLayers{top: []spriteLayer{{image: compressorSpriteForOrientation(v.Orientation)}}}
	case *components.Binder:
		return componentSpriteLayers{top: []spriteLayer{{image: sprites.binder, rotation: binderRotation(v.Orientation)}}}
	}
	return componentSpriteLayers{}
}

// perpendicular returns the horizontal axis if d is vertical, and vice versa.
// Used by Duplicator placeholder rendering to draw its output pipe.
func perpendicular(d sim.Direction) sim.Direction {
	if d == sim.DirNorth || d == sim.DirSouth {
		return sim.DirEast
	}
	return sim.DirNorth
}

func pipeSpriteForOrientation(d sim.Direction) *ebiten.Image {
	if d == sim.DirNorth || d == sim.DirSouth {
		return sprites.pipeVert
	}
	return sprites.pipeHori
}

func compressorSpriteForOrientation(d sim.Direction) *ebiten.Image {
	if d == sim.DirNorth || d == sim.DirSouth {
		return sprites.compressorVert
	}
	return sprites.compressorHori
}

func meshGridSpriteForOrientation(d sim.Direction) *ebiten.Image {
	if d == sim.DirNorth || d == sim.DirSouth {
		return sprites.meshGridVert
	}
	return sprites.meshGridHori
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

// collectorRotation maps a placed collector's direction to a rotation angle.
// The source PNG is drawn facing South (down) — DirSouth is the unrotated case.
func collectorRotation(d sim.Direction) float64 {
	switch d {
	case sim.DirNorth:
		return math.Pi
	case sim.DirEast:
		return -math.Pi / 2
	case sim.DirWest:
		return math.Pi / 2
	default:
		return 0
	}
}

// binderRotation maps a Binder's Orientation to a rotation angle.
// The source PNG is drawn facing South — same convention as the Collector.
func binderRotation(d sim.Direction) float64 {
	return collectorRotation(d)
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
	case ui.ToolInjector:
		return sprites.injector
	case ui.ToolAccelerator:
		return sprites.acceleratorTop
	case ui.ToolMeshGrid:
		return sprites.meshGridTop
	case ui.ToolMagnetiser:
		return sprites.magnetiserTop
	case ui.ToolElbow:
		return sprites.turnNE
	case ui.ToolPipe:
		return sprites.pipeHori
	case ui.ToolCollector:
		return sprites.collector
	case ui.ToolResonator:
		return sprites.magnetiserBottom
	case ui.ToolCatalyst:
		return sprites.injector
	case ui.ToolDuplicator:
		return sprites.turnNE
	case ui.ToolCompressor:
		return sprites.compressorHori
	case ui.ToolBinder:
		return sprites.binder
	}
	return nil
}

// logoSpriteForTool returns the inventory-facing logo icon for a Tool.
// Dedicated logos exist for Injector, Accelerator, Mesh Grid, Magnetiser,
// and Elbow. Every other Tool falls back to the generic pipe logo.
func logoSpriteForTool(t ui.Tool) *ebiten.Image {
	switch t {
	case ui.ToolInjector:
		return sprites.injectorLogo
	case ui.ToolAccelerator:
		return sprites.accelLogo
	case ui.ToolMeshGrid:
		return sprites.meshLogo
	case ui.ToolMagnetiser:
		return sprites.magnetLogo
	case ui.ToolElbow:
		return sprites.turnLogo
	}
	return sprites.pipeLogo
}
