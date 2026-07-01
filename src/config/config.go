// Package config holds shared game-wide constants.
package config

import "github.com/Zyko0/go-sdl3/sdl"

const (
	// WindowWidth is the actual window width in screen pixels.
	WindowWidth = 1280
	// WindowHeight is the actual window height in screen pixels.
	WindowHeight = 960

	// LogicalWidth is the renderer's logical resolution width.
	LogicalWidth = 1920
	// LogicalHeight is the renderer's logical resolution height.
	LogicalHeight = 1080

	// PresentationMode controls how the logical resolution is mapped to the
	// window. LETTERBOX preserves aspect ratio and avoids pixel jitter.
	PresentationMode = sdl.LOGICAL_PRESENTATION_LETTERBOX

	// TileSize is the native pixel size of a tile sprite (before scaling).
	TileSize = 16
	// Scale is the pixel-art upscale factor.
	Scale = 3
	// ScaledTile is the logical pixel size of one tile (TileSize * Scale).
	ScaledTile = TileSize * Scale

	// CameraZoom is the default camera zoom factor (< 1 = zoomed out).
	CameraZoom = 1.0
)
