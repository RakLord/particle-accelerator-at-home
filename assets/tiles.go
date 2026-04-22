package assets

import "embed"

// TileFS embeds the current tile art so desktop and WASM builds use the same assets.
//
//go:embed images/tiles/*.png
var TileFS embed.FS
