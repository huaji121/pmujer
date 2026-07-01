// Package config holds shared game-wide constants.
package config

import "github.com/Zyko0/go-sdl3/sdl"

const (
	// WindowWidth is the actual window width in screen pixels.
	WindowWidth = 1280
	// WindowHeight is the actual window height in screen pixels.
	WindowHeight = 720

	// LogicalWidth is the renderer's logical resolution width.
	LogicalWidth = 1920
	// LogicalHeight is the renderer's logical resolution height.
	LogicalHeight = 1080

	// PresentationMode controls how the logical resolution is mapped to the
	// window. STRETCH fills the entire window without letterbox bars.
	PresentationMode = sdl.LOGICAL_PRESENTATION_STRETCH

	// TileSize is the native pixel size of a tile sprite (before scaling).
	TileSize = 16
	// Scale is the pixel-art upscale factor.
	Scale = 3
	// ScaledTile is the logical pixel size of one tile (TileSize * Scale).
	ScaledTile = TileSize * Scale

	// CameraZoom is the default camera zoom factor (< 1 = zoomed out).
	CameraZoom = 1.5

	// CameraSpeed controls how fast the camera follows the player (higher = snappier).
	CameraSpeed float32 = 5.0
)
