// Package camera provides a 2-D follow camera with framerate-independent lerp.
package camera

import (
	"math"

	"pmujer/src/config"
)

// Camera tracks a target position and smoothly follows it.
type Camera struct {
	X, Y float32 // top-left corner in world-pixels

	// Follow speed (higher = snappier). Used in the exp decay formula.
	Speed float32

	// Zoom factor (< 1 = zoomed out, shows more; > 1 = zoomed in).
	Zoom float32

	// Viewport size in logical pixels.
	ViewW, ViewH float32

	// Level bounds in world-pixels (set once via SetBounds).
	maxX, maxY float32
}

// New creates a camera centred on (0, 0) with the given viewport size.
func New(viewW, viewH float32) *Camera {
	return &Camera{
		Speed: config.CameraSpeed,
		Zoom:  1.0,
		ViewW: viewW,
		ViewH: viewH,
	}
}

// SetBounds configures the level dimensions so the camera can clamp.
// width/height are in tiles.
func (c *Camera) SetBounds(width, height int) {
	scaled := float32(config.ScaledTile)
	visW := c.ViewW / c.Zoom
	visH := c.ViewH / c.Zoom
	c.maxX = float32(width)*scaled - visW
	c.maxY = float32(height)*scaled - visH
}

// Follow smoothly moves the camera toward the given world position (pixels)
// so that it sits at the centre of the visible area.  dt is the frame delta
// in seconds.
func (c *Camera) Follow(targetX, targetY, dt float32) {
	visW := c.ViewW / c.Zoom
	visH := c.ViewH / c.Zoom

	// Place target at visible-area centre.
	destX := targetX - visW/2
	destY := targetY - visH/2

	// Framerate-independent exponential lerp.
	// At Speed=5.0, one second closes ~99.3% of the gap; the trailing
	// effect is clearly visible frame-to-frame.
	factor := float32(1.0 - math.Exp(float64(-c.Speed*dt)))
	c.X += (destX - c.X) * factor
	c.Y += (destY - c.Y) * factor

	// Clamp to level bounds.
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	if c.maxX > 0 && c.X > c.maxX {
		c.X = c.maxX
	}
	if c.maxY > 0 && c.Y > c.maxY {
		c.Y = c.maxY
	}

	// Snap to integer logical pixels to prevent sub-pixel jitter.
	c.X = float32(math.Round(float64(c.X)))
	c.Y = float32(math.Round(float64(c.Y)))
}