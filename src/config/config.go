// Package config holds shared game-wide constants.
package config

const (
	// WindowWidth is the window width in screen pixels.
	WindowWidth = 1280
	// WindowHeight is the window height in screen pixels.
	WindowHeight = 720

	// TileSize is the native pixel size of a tile sprite (before scaling).
	TileSize = 16
	// Scale is the pixel-art upscale factor.
	Scale = 3
	// ScaledTile is the on-screen pixel size of one tile (TileSize * Scale).
	ScaledTile = TileSize * Scale

	// CameraZoom is the default camera zoom factor (< 1 = zoomed out).
	CameraZoom = 0.6
)
