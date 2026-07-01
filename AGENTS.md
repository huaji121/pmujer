 # AGENTS.md

 Codex guidance for working on **Pmujer**, a pixel-art 2D platformer written in Go using SDL3 via `github.com/Zyko0/go-sdl3`.

 ## Objective

 Keep the game buildable, runnable, and easy to extend. When modifying code, preserve the existing architecture, tile-unit conventions, and SDL3 usage patterns already established in the repository.

 ## Build & Run

 Prefer the existing commands when validating changes:

 ```bash
 go build ./src/...
 go build -o pmujer.exe ./src/
 go run ./src/
 go vet ./src/...
 bash start.sh
 ```

 If tests are added later, run them with `go test ./src/...` or a single package such as `go test ./src/tilemap/`.

 ## Repository Structure

 - `src/main.go` initializes SDL, builds the level, runs the game loop, handles input, and orchestrates rendering/audio/particles.
 - `src/tilemap/` owns the tile registry, collision system, and per-tile-type definitions.
 - `src/player/` handles movement, physics, collision response, and death/respawn logic.
 - `src/camera/` provides framerate-independent follow behavior.
 - `src/particle/` runs a small particle system for visual effects.
 - `src/audio/` plays WAV sounds through the SDL3 core audio API.
 - `assets/textures/` contains 16x16 pixel art textures.
 - `assets/sounds/` contains WAV audio assets.
 - `docs/go-sdl3-reference.md` contains the current go-sdl3 API cheatsheet for this project.

 ## Working Rules

 1. Read relevant source files before changing behavior; do not assume module boundaries from file names alone.
 2. Keep game logic in **tile units**. Convert to pixels only when rendering.
 3. Preserve the project convention that `TileID`, `TileDef`, and the tile registry are the extensibility points for tile types.
 4. When adding a new tile type, follow the existing pattern: add a new `tile_xxx.go` file, define a `TileID` constant, implement `registerXxx(renderer)`, and call it from `LoadAssets()`.
 5. Keep collision behavior explicit: solid tiles use full-tile AABB, deadly tiles may use `ConvexHitbox` SAT testing.
 6. Preserve the SDL3 usage rules already documented in `CLAUDE.md` and `docs/go-sdl3-reference.md`.

 ## SDL3 API Notes

 This project uses `go-sdl3` without CGO. Key conventions from the repository reference:

 - Load embedded SDL libraries with `binsdl.Load()` and `binimg.Load()` before `sdl.Init()`, and unload them after quitting in reverse order.
 - Use `sdl.RunLoop(...)` with `sdl.EndLoop` for the main loop.
 - Access SDL event data through typed methods, e.g. `event.KeyboardEvent()`, not direct field access.
 - Use `sdl.GetKeyboardState()` for real-time movement/input polling.
 - Render textures with `renderer.RenderTexture(...)`.
 - Draw outlines with `renderer.RenderRect(...)`, filled rectangles with `renderer.RenderFillRect(...)`, and lines with `renderer.RenderLine(...)`.
 - Audio is handled with SDL3 core audio streams: `sdl.LoadWAV(...)` plus `stream.PutData(...)`.

 ## Coding Style

 - Keep changes minimal and consistent with nearby code.
 - Prefer clear, direct naming over clever abstractions.
 - Add a new package only when it clearly separates a distinct responsibility.
 - Avoid removing existing debug aids unless the change is intentional and coordinated.

 ## Safe Change Checklist

 - If touching `tilemap`, verify collision logic still matches the documented tile-unit model.
 - If touching `player`, verify movement, jump behavior, and death/respawn semantics.
 - If touching `camera`, verify follow behavior remains framerate-independent and level-bounded.
 - If touching `audio`, verify WAV loading and playback still use the SDL3 stream model.
 - If changing assets, verify textures remain pixel-art friendly and aligned to the 16x16 source convention.

 ## References

 - `CLAUDE.md`
 - `docs/go-sdl3-reference.md`
