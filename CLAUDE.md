# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build ./src/...          # compile check (no binary)
go build -o pmujer.exe ./src/  # build executable
go run ./src/               # build and run
go vet ./src/...            # static analysis
bash start.sh               # build and run (shorthand)
```

No tests exist yet. When added: `go test ./src/...` / `go test ./src/tilemap/` for a single package.

## Architecture

Pmujer is a pixel-art 2D platformer using **Go + SDL3** via `github.com/Zyko0/go-sdl3`.

### Rendering pipeline
- Native sprite size: **16×16 px**, upscaled **3×** to **48 px** on screen (`tilemap.ScaledTile`)
- All game logic uses **tile-units** (1 unit = 1 tile = 48 screen pixels); only convert to pixels at render time
- Camera is pixel-based; subtract `camX`/`camY` when rendering world objects
- Textures loaded via `img.LoadTexture()` (SDL3_image); set `SCALEMODE_NEAREST` for crisp pixels

### Packages

**`src/tilemap`** — Extensible tile system with per-tile-type files.
- `TileID` (uint8) identifies tile types. `TileDef` in the global `Registry` map holds per-type data (texture, solid, deadly, hitbox).
- Each tile type lives in its own file (`tile_bricks.go`, `tile_spike.go`): exports a `registerXxx(renderer)` function called from `LoadAssets()`.
- `ConvexHitbox` (`[]Vec2`) defines precise collision shapes for deadly tiles; nil means full-tile AABB.
- `HitDeadly(rx, ry, rw, rh)` uses SAT (separating axis theorem) to test AABB vs convex polygon overlap.
- To add a new tile: create `tile_xxx.go`, define `TileID` constant, write `registerXxx()`, call it from `LoadAssets()`.

**`src/player`** — Player character.
- Position/velocity in tile-units. Physics: gravity, variable jump height, faster-fall multiplier, double jump.
- Collision box (6×11 px centred in 16×16 sprite) defined by `playerColW`/`playerColH`/`colOffX`/`colOffY`.
- `JustJumped` flag (true for one frame) lets main.go trigger audio.
- `Die()` / `Respawn()` / `LastVisX/Y` for death/respawn lifecycle.

**`src/camera`** — Framerate-independent follow camera.
- `cam.Follow(targetX, targetY, dt)` uses exponential lerp (`1 - e^(-speed*dt)`) + level-bounds clamping.

**`src/particle`** — Simple particle system.
- `System.Emit(cx, cy, count, speed, life, size)` spawns a burst; `Update(dt)` applies gravity + culls dead particles; `Render()` fades alpha over lifetime.

**`src/audio`** — WAV playback via SDL3 core audio API.
- `NewSystem()` opens default playback device + audio stream.
- `LoadWAV()` / `Play()` — fire-and-forget, multiple calls overlap.

**`src/main`** — Glue: SDL init, level layout (`buildLevel()`), game loop, input dispatch, death detection.

### Key conventions
- Level data is a 2D grid in `buildLevel()`. Row 0 = top; increasing Y = downward.
- SDL3 event access uses typed methods on `sdl.Event` (e.g. `event.KeyboardEvent()`), not direct field access.
- `binsdl.Load()` / `binimg.Load()` embed SDL3/SDL3_image DLLs at compile time — no external DLL install needed.
- `defer` order matters: `binsdl` and `binimg` must load before `sdl.Init()`, and quit/close in reverse.
- Debug mode (F3 toggle) draws collision boxes: green for player AABB, red for tile hitboxes.
- See `docs/go-sdl3-reference.md` for the go-sdl3 API cheatsheet.
