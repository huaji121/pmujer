// Package camera provides a 2-D follow camera with framerate-independent lerp.
package camera

import (
	"math"

	"pmujer/src/tilemap"
)

// Camera tracks a target position and smoothly follows it.
type Camera struct {
	X, Y float32 // top-left corner in pixels

	// Follow speed (higher = snappier). Used in the exp decay formula.
	Speed float32

	// Viewport size in pixels.
	ViewW, ViewH float32

	// Level bounds in pixels (set once via SetBounds).
	maxX, maxY float32
}

// New creates a camera centred on (0, 0) with the given viewport size.
func New(viewW, viewH float32) *Camera {
	return &Camera{
		Speed:  10.0,
		ViewW:  viewW,
		ViewH:  viewH,
	}
}

// SetBounds configures the level dimensions so the camera can clamp.
// width/height are in tiles.
func (c *Camera) SetBounds(width, height int) {
	scaled := float32(tilemap.ScaledTile)
	c.maxX = float32(width)*scaled - c.ViewW
	c.maxY = float32(height)*scaled - c.ViewH
}

// Follow smoothly moves the camera toward the given world position (pixels)
// so that it sits at the centre of the viewport.  dt is the frame delta in
// seconds.
func (c *Camera) Follow(targetX, targetY, dt float32) {
	// Place target at viewport centre.
	destX := targetX - c.ViewW/2
	destY := targetY - c.ViewH/2

	// Framerate-independent exponential lerp.
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
	if c.X > c.maxX {
		c.X = c.maxX
	}
	if c.Y > c.maxY {
		c.Y = c.maxY
	}
}
