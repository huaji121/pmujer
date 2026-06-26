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
	Speed          = 6.0   // horizontal movement speed (tiles/s)
	JumpVelocity   = -13.5 // initial vertical velocity on jump (tiles/s, up)
	Gravity        = 28.0  // downward acceleration (tiles/s²)
	GravityFallMul = 1.8   // extra gravity multiplier when falling
	MaxFallSpeed   = 20.0  // terminal fall speed (tiles/s)
	MaxJumps       = 2     // number of jumps allowed (1 = single, 2 = double)
)

// Collision box dimensions in tile-units.
// The sprite is 16×16 but the visible character is centered within it,
// so the hitbox is smaller than a full tile.
const (
	playerColW = 0.45 // collision box width
	playerColH = 0.7  // collision box height
)

// Offsets to position the collision box inside the 1×1 sprite area.
// X: collision box is horizontally centred → (1 - 0.5) / 2 = 0.25
// Y: collision box bottom aligns with sprite bottom → 1 - 0.7 = 0.30
const (
	colOffX = 0.27
	colOffY = 0.30
)

// Player is the controllable character.
// X, Y is the top-left of the collision box (not the sprite).
type Player struct {
	X, Y   float32 // collision box top-left in tile-units
	VX, VY float32 // velocity in tile-units/s

	Texture *sdl.Texture

	// State
	Grounded       bool
	Alive          bool
	FacingRight    bool
	JumpsLeft      int  // remaining jumps (resets to MaxJumps on landing)
	jumpWasPressed bool // previous frame's jump key state, for edge detection

	// Spawn point for respawning (collision box top-left)
	SpawnX, SpawnY float32
}

// New creates a player so that its collision box sits at (tileX, tileY).
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
		JumpsLeft:   MaxJumps,
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

	// Jump: W or J — edge-triggered, supports double jump
	jumpPressed := keys[sdl.SCANCODE_W] || keys[sdl.SCANCODE_J]
	if jumpPressed && !p.jumpWasPressed && p.JumpsLeft > 0 {
		p.VY = float32(JumpVelocity)
		p.Grounded = false
		p.JumpsLeft--
	}
	p.jumpWasPressed = jumpPressed

	// Variable jump height: releasing the jump key early cuts upward velocity
	if !jumpPressed && p.VY < float32(JumpVelocity)*0.4 {
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
	left := int(math.Floor(float64(p.X)))
	right := int(math.Floor(float64(p.X + playerColW - 0.001)))
	top := int(math.Floor(float64(p.Y)))
	bottom := int(math.Floor(float64(p.Y + playerColH - 0.001)))

	for ty := top; ty <= bottom; ty++ {
		for tx := left; tx <= right; tx++ {
			if !tm.IsSolid(tx, ty) {
				continue
			}
			tileL := float32(tx)
			tileR := float32(tx) + 1

			if p.VX > 0 { // moving right → push left
				p.X = tileL - playerColW
			} else if p.VX < 0 { // moving left → push right
				p.X = tileR
			}
			p.VX = 0
		}
	}
}

// resolveVertical pushes the player out of any solid tiles on the Y axis.
func (p *Player) resolveVertical(tm *tilemap.Tilemap) {
	left := int(math.Floor(float64(p.X)))
	right := int(math.Floor(float64(p.X + playerColW - 0.001)))
	top := int(math.Floor(float64(p.Y)))
	bottom := int(math.Floor(float64(p.Y + playerColH - 0.001)))

	p.Grounded = false

	for ty := top; ty <= bottom; ty++ {
		for tx := left; tx <= right; tx++ {
			if !tm.IsSolid(tx, ty) {
				continue
			}
			tileT := float32(ty)
			tileB := float32(ty) + 1

			if p.VY > 0 { // moving down → land on top
				p.Y = tileT - playerColH
				p.VY = 0
				p.Grounded = true
				p.JumpsLeft = MaxJumps
			} else if p.VY < 0 { // moving up → hit ceiling
				p.Y = tileB
				p.VY = 0
			}
		}
	}
}

// Respawn teleports the player back to their spawn point.
func (p *Player) Respawn() {
	p.X = p.SpawnX
	p.Y = p.SpawnY
	p.VX = 0
	p.VY = 0
	p.Grounded = false
	p.JumpsLeft = MaxJumps
	p.jumpWasPressed = false
	p.Alive = true
}

// Render draws the player sprite centered on the collision box.
func (p *Player) Render(renderer *sdl.Renderer, camX, camY float32) {
	if !p.Alive {
		return
	}
	scaled := float32(tilemap.ScaledTile)
	// Sprite is 1×1 tile-unit, offset so the collision box is centered inside it.
	offX, offY := float32(colOffX), float32(colOffY)
	dst := sdl.FRect{
		X: (p.X-offX)*scaled - camX,
		Y: (p.Y-offY)*scaled - camY,
		W: scaled,
		H: scaled,
	}
	renderer.RenderTexture(p.Texture, nil, &dst)
}

// CenterX returns the collision box centre X in tile-units (for camera).
func (p *Player) CenterX() float32 {
	return p.X + playerColW/2
}

// CenterY returns the collision box centre Y in tile-units.
func (p *Player) CenterY() float32 {
	return p.Y + playerColH/2
}
