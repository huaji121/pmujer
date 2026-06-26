# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build ./src/...          # compile check (no binary)
go build -o pmujer.exe ./src/  # build executable
go run ./src/               # build and run
go vet ./src/...            # static analysis
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

**`src/tilemap`** — Extensible tile system.
- `TileID` (uint8) identifies tile types. `TileDef` in the global `Registry` map holds per-type data (texture, solid, deadly).
- To add a new tile: define a `TileID` constant, load its texture, call `Register()`.
- `Tilemap` is a 2D grid of `TileID`. `IsSolid(x,y)` / `IsDeadly(x,y)` are the collision queries.

**`src/player`** — Player character.
- Position/velocity in tile-units. Physics: gravity, variable jump height, faster-fall multiplier.
- Collision: separates horizontal and vertical axes to avoid corner-catching.
- Input: `sdl.GetKeyboardState()` indexed by `sdl.SCANCODE_*` (polled each frame).

**`src/main`** — Glue: SDL init, level layout (`buildLevel()`), game loop, camera follow.

### Key conventions
- Level data is a 2D grid in `buildLevel()`. Row 0 = top; increasing Y = downward.
- SDL3 event access uses typed methods on `sdl.Event` (e.g. `event.KeyboardEvent()`), not direct field access.
- `binsdl.Load()` / `binimg.Load()` embed SDL3/SDL3_image DLLs at compile time — no external DLL install needed.
- `defer` order matters: `binsdl` and `binimg` must load before `sdl.Init()`, and quit/close in reverse.
