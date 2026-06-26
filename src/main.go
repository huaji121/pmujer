package main

import (
	"github.com/Zyko0/go-sdl3/bin/binimg"
	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"

	"pmujer/src/camera"
	"pmujer/src/particle"
	"pmujer/src/player"
	"pmujer/src/tilemap"
)

const (
	WindowWidth  = 960
	WindowHeight = 720
)

// buildLevel creates the default level layout.
//
//	Map size: 40 × 15 tiles (each tile = 16 px × scale 3 = 48 px on screen).
//	Index 0 = top row, index 14 = bottom row.
func buildLevel() *tilemap.Tilemap {
	tm := tilemap.New(40, 15)

	// Ground (row 13-14) with a few gaps
	for x := 0; x < 40; x++ {
		// Gap at x=8-9, x=18-19, x=30-31
		if (x >= 8 && x <= 9) || (x >= 18 && x <= 19) || (x >= 30 && x <= 31) {
			continue
		}
		tm.Set(x, 13, tilemap.TileBricks)
		tm.Set(x, 14, tilemap.TileBricks)
	}

	// Floating platforms
	platforms := []struct {
		x, y, w int
	}{
		{5, 10, 3},   // low platform near start
		{12, 9, 4},   // mid platform
		{17, 7, 3},   // high platform
		{22, 10, 5},  // long low platform
		{25, 6, 3},   // high platform
		{30, 9, 4},   // mid platform
		{35, 11, 3},  // low platform near end
	}
	for _, p := range platforms {
		for dx := 0; dx < p.w; dx++ {
			tm.Set(p.x+dx, p.y, tilemap.TileBricks)
		}
	}

	// Step blocks
	for i := 0; i < 3; i++ {
		tm.Set(37-i, 12-i, tilemap.TileBricks)
	}

	return tm
}

func main() {
	// Load SDL3 and SDL3_image from embedded binaries.
	defer binsdl.Load().Unload()
	defer binimg.Load().Unload()
	defer sdl.Quit()

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}

	window, renderer, err := sdl.CreateWindowAndRenderer(
		"Pmujer", WindowWidth, WindowHeight, 0,
	)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()
	defer window.Destroy()

	// Disable texture filtering – keep pixels crisp.
	renderer.SetDefaultTextureScaleMode(sdl.SCALEMODE_NEAREST)

	// Load assets and build level.
	tilemap.LoadAssets(renderer)
	tm := buildLevel()

	// Spawn the player on the first safe ground tile (row 12, column 1).
	p := player.New(renderer, 1, 12)

	// Particle system for blood effects.
	bloodPS := particle.NewSystem(renderer, "assets/textures/blood.png", 20.0)
	var wasAlive = true // track previous frame's alive state

	// Camera.
	cam := camera.New(WindowWidth, WindowHeight)
	cam.SetBounds(tm.Width, tm.Height)

	// Debug mode – toggle with F3.
	debug := false

	// Track real time for a fixed-step physics loop.
	var lastTicks uint64

	// ----- Main loop -----
	sdl.RunLoop(func() error {
		// -- Delta time (seconds) --
		now := sdl.Ticks()
		if lastTicks == 0 {
			lastTicks = now
		}
		dt := float32(now-lastTicks) / 1000.0
		lastTicks = now

		// Cap dt to avoid spiral-of-death on long hitches.
		if dt > 0.05 {
			dt = 0.05
		}

		// -- Events --
		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type {
			case sdl.EVENT_QUIT:
				return sdl.EndLoop
			case sdl.EVENT_KEY_DOWN:
				key := event.KeyboardEvent()
				if key.Scancode == sdl.SCANCODE_ESCAPE {
					return sdl.EndLoop
				}
				// F3 to toggle debug mode
				if key.Scancode == sdl.SCANCODE_F3 {
					debug = !debug
				}
				// R to die and respawn (triggers blood particles when alive)
				if key.Scancode == sdl.SCANCODE_R {
					if p.Alive {
						bloodPS.Emit(p.CenterX(), p.CenterY(), 30, 8.0, 1.0, 0.3)
					}
					p.Respawn()
				}
			}
		}

		// -- Update --
		p.Update(dt, tm)

		// Detect death from falling off map → emit blood particles
		if wasAlive && !p.Alive {
			bloodPS.Emit(p.LastVisX, p.LastVisY, 30, 8.0, 1.0, 0.3)
		}
		wasAlive = p.Alive

		bloodPS.Update(dt)

		// -- Camera --
		scaled := float32(tilemap.ScaledTile)
		cam.Follow(p.CenterX()*scaled, p.CenterY()*scaled, dt)

		// -- Render --
		renderer.SetDrawColor(135, 206, 235, 255) // sky blue
		renderer.Clear()

		tm.Render(renderer, cam.X, cam.Y)
		bloodPS.Render(renderer, cam.X, cam.Y)
		p.Render(renderer, cam.X, cam.Y, debug)

		renderer.Present()

		return nil
	})
}
