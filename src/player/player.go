// Package player provides the playable character with movement, jumping,
// physics and tilemap collision.
package player

import (
	"math"

	"github.com/Zyko0/go-sdl3/img"
	"github.com/Zyko0/go-sdl3/sdl"

	"pmujer/src/tilemap"
)

// Physics constants (units: tile-units and seconds).
const (
	Speed          = 6.0  // horizontal movement speed (tiles/s)
	JumpVelocity   = -13.5 // initial vertical velocity on jump (tiles/s, up)
	Gravity        = 28.0  // downward acceleration (tiles/s²)
	GravityFallMul = 1.8  // extra gravity multiplier when falling
	MaxFallSpeed   = 20.0  // terminal fall speed (tiles/s)
)

// Player is the controllable character.
type Player struct {
	X, Y   float32 // position in tile-units (top-left of the 1×1 bounding box)
	VX, VY float32 // velocity in tile-units/s

	Texture *sdl.Texture

	// State
	Grounded  bool
	Alive     bool
	FacingRight bool

	// Spawn point for respawning
	SpawnX, SpawnY float32
}

// New creates a player at the given tile position and loads its texture.
func New(renderer *sdl.Renderer, tileX, tileY float32) *Player {
	tex, err := img.LoadTexture(renderer, "assets/textures/player.png")
	if err != nil {
		panic("failed to load player texture: " + err.Error())
	}
	return &Player{
		X:           tileX,
		Y:           tileY,
		Texture:     tex,
		Alive:       true,
		FacingRight: true,
		SpawnX:      tileX,
		SpawnY:      tileY,
	}
}

// Update runs one physics step: reads keyboard, applies gravity, moves,
// and resolves collisions against the tilemap.
func (p *Player) Update(dt float32, tm *tilemap.Tilemap) {
	if !p.Alive {
		return
	}

	// --- Input via SDL keyboard state (scancodes) ---
	keys := sdl.GetKeyboardState()

	// Horizontal movement: A = left, D = right
	p.VX = 0
	if keys[sdl.SCANCODE_A] {
		p.VX = -Speed
		p.FacingRight = false
	}
	if keys[sdl.SCANCODE_D] {
		p.VX = Speed
		p.FacingRight = true
	}

	// Jump: W or J (only when grounded)
	if (keys[sdl.SCANCODE_W] || keys[sdl.SCANCODE_J]) && p.Grounded {
		p.VY = JumpVelocity
		p.Grounded = false
	}

	// Variable jump height: releasing the jump key early cuts upward velocity
	if !keys[sdl.SCANCODE_W] && !keys[sdl.SCANCODE_J] && p.VY < float32(JumpVelocity)*0.4 {
		p.VY = float32(JumpVelocity) * 0.4
	}

	// --- Gravity ---
	grav := float32(Gravity)
	if p.VY > 0 {
		grav *= float32(GravityFallMul) // fall faster than we rise
	}
	p.VY += grav * dt
	if p.VY > MaxFallSpeed {
		p.VY = MaxFallSpeed
	}

	// --- Move & collide (separate axes) ---
	p.moveAndCollide(dt, tm)

	// --- Respawn if fallen off the map ---
	if p.Y > float32(tm.Height+5) {
		p.Respawn()
	}
}

// moveAndCollide handles movement on each axis independently so that
// sliding along walls/floors works naturally.
func (p *Player) moveAndCollide(dt float32, tm *tilemap.Tilemap) {
	// --- Horizontal ---
	p.X += p.VX * dt
	p.resolveHorizontal(tm)

	// --- Vertical ---
	p.Y += p.VY * dt
	p.resolveVertical(tm)
}

// resolveHorizontal pushes the player out of any solid tiles on the X axis.
func (p *Player) resolveHorizontal(tm *tilemap.Tilemap) {
	pW, pH := playerSize()

	left := int(math.Floor(float64(p.X)))
	right := int(math.Floor(float64(p.X + pW - 0.001)))
	top := int(math.Floor(float64(p.Y)))
	bottom := int(math.Floor(float64(p.Y + pH - 0.001)))

	for ty := top; ty <= bottom; ty++ {
		for tx := left; tx <= right; tx++ {
			if !tm.IsSolid(tx, ty) {
				continue
			}
			tileL := float32(tx)
			tileR := float32(tx) + 1

			if p.VX > 0 { // moving right → push left
				p.X = tileL - pW
			} else if p.VX < 0 { // moving left → push right
				p.X = tileR
			}
			p.VX = 0
		}
	}
}

// resolveVertical pushes the player out of any solid tiles on the Y axis.
func (p *Player) resolveVertical(tm *tilemap.Tilemap) {
	pW, pH := playerSize()

	left := int(math.Floor(float64(p.X)))
	right := int(math.Floor(float64(p.X + pW - 0.001)))
	top := int(math.Floor(float64(p.Y)))
	bottom := int(math.Floor(float64(p.Y + pH - 0.001)))

	p.Grounded = false

	for ty := top; ty <= bottom; ty++ {
		for tx := left; tx <= right; tx++ {
			if !tm.IsSolid(tx, ty) {
				continue
			}
			tileT := float32(ty)
			tileB := float32(ty) + 1

			if p.VY > 0 { // moving down → land on top
				p.Y = tileT - pH
				p.VY = 0
				p.Grounded = true
			} else if p.VY < 0 { // moving up → hit ceiling
				p.Y = tileB
				p.VY = 0
			}
		}
	}
}

// playerSize returns the player's bounding-box dimensions in tile-units.
// The sprite is 16×16 at scale 3, same as one tile → 1×1 tile-units.
func playerSize() (float32, float32) {
	return 1.0, 1.0
}

// Respawn teleports the player back to their spawn point.
func (p *Player) Respawn() {
	p.X = p.SpawnX
	p.Y = p.SpawnY
	p.VX = 0
	p.VY = 0
	p.Grounded = false
	p.Alive = true
}

// Render draws the player sprite at the current screen position.
func (p *Player) Render(renderer *sdl.Renderer, camX, camY float32) {
	if !p.Alive {
		return
	}
	scaled := float32(tilemap.ScaledTile)
	dst := sdl.FRect{
		X: p.X*scaled - camX,
		Y: p.Y*scaled - camY,
		W: scaled,
		H: scaled,
	}
	renderer.RenderTexture(p.Texture, nil, &dst)
}

// CenterX returns the player's centre X in tile-units (for camera tracking).
func (p *Player) CenterX() float32 {
	return p.X + 0.5
}

// CenterY returns the player's centre Y in tile-units.
func (p *Player) CenterY() float32 {
	return p.Y + 0.5
}
