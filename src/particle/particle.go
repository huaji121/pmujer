// Package particle provides a simple 2D particle system for visual effects.
package particle

import (
	"math"
	"math/rand/v2"

	"github.com/Zyko0/go-sdl3/img"
	"github.com/Zyko0/go-sdl3/sdl"

	"pmujer/src/tilemap"
)

// Particle is a single particle in the system.
type Particle struct {
	X, Y   float32 // position in tile-units
	VX, VY float32 // velocity in tile-units/s
	Life   float32 // remaining lifetime in seconds
	MaxLife float32 // initial lifetime (for alpha fade calculation)
	Size    float32 // render size in tile-units
}

// System manages a pool of particles sharing one texture.
type System struct {
	Texture    *sdl.Texture
	Particles  []Particle
	Gravity    float32 // tiles/s²
}

// NewSystem loads the particle texture and returns a ready-to-use system.
func NewSystem(renderer *sdl.Renderer, texturePath string, gravity float32) *System {
	tex, err := img.LoadTexture(renderer, texturePath)
	if err != nil {
		panic("failed to load particle texture: " + err.Error())
	}
	return &System{
		Texture:   tex,
		Particles: make([]Particle, 0, 256),
		Gravity:   gravity,
	}
}

// Emit spawns a burst of particles at (cx, cy) in tile-units.
// count: number of particles; speed: initial velocity range; life: lifetime range (seconds).
func (s *System) Emit(cx, cy float32, count int, speed, life, size float32) {
	for i := 0; i < count; i++ {
		// Random direction (full 360°)
		angle := rand.Float32() * 2 * math.Pi
		sp := speed * (0.4 + 0.6*rand.Float32()) // 40%–100% of speed
		s.Particles = append(s.Particles, Particle{
			X:       cx,
			Y:       cy,
			VX:      float32(math.Cos(float64(angle))) * sp,
			VY:      float32(math.Sin(float64(angle))) * sp,
			Life:    life * (0.5 + 0.5*rand.Float32()),
			MaxLife: life,
			Size:    size * (0.6 + 0.4*rand.Float32()),
		})
	}
}

// Update advances all particles by dt seconds. Dead particles are removed.
func (s *System) Update(dt float32) {
	alive := s.Particles[:0]
	for i := range s.Particles {
		p := &s.Particles[i]
		p.Life -= dt
		if p.Life <= 0 {
			continue
		}
		p.VY += s.Gravity * dt
		p.X += p.VX * dt
		p.Y += p.VY * dt
		alive = append(alive, *p)
	}
	s.Particles = alive
}

// Render draws all particles. camX/camY are the camera offset in pixels.
func (s *System) Render(renderer *sdl.Renderer, camX, camY float32) {
	scaled := float32(tilemap.ScaledTile)
	for i := range s.Particles {
		p := &s.Particles[i]
		// Alpha fades from 255 → 0 over the particle's lifetime.
		alpha := uint8(float32(255) * (p.Life / p.MaxLife))
		s.Texture.SetAlphaMod(alpha)

		sizePx := p.Size * scaled
		dst := sdl.FRect{
			X: p.X*scaled - camX - sizePx/2,
			Y: p.Y*scaled - camY - sizePx/2,
			W: sizePx,
			H: sizePx,
		}
		renderer.RenderTexture(s.Texture, nil, &dst)
	}
}

// Clear removes all particles.
func (s *System) Clear() {
	s.Particles = s.Particles[:0]
}
