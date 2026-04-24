package assets

import "embed"

// TileFS embeds the current tile art so desktop and WASM builds use the same assets.
//
//go:embed images/tiles/empty_tile.png
//go:embed images/tiles/pipe_hori.png
//go:embed images/tiles/pipe_vert.png
//go:embed images/tiles/turn_ne.png
//go:embed images/tiles/turn_nw.png
//go:embed images/tiles/turn_se.png
//go:embed images/tiles/turn_sw.png
//go:embed images/tiles/accelerator_bottom.png
//go:embed images/tiles/accelerator_top.png
//go:embed images/tiles/mesh_grid_top.png
//go:embed images/tiles/collector.png
//go:embed images/tiles/injector.png
//go:embed images/tiles/magnetiser_top.png
//go:embed images/tiles/magnetiser_bottom.png
//go:embed images/tiles/accelerator_logo.png
//go:embed images/tiles/mesh_grid_logo.png
//go:embed images/tiles/magnetiser_logo.png
//go:embed images/tiles/pipe_logo.png
var TileFS embed.FS
