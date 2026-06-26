// Package tilemap provides an extensible tile-based map system for the game.
package tilemap

import (
	"fmt"

	"github.com/Zyko0/go-sdl3/img"
	"github.com/Zyko0/go-sdl3/sdl"
)

// TileSize is the native pixel size of a tile sprite (before scaling).
const TileSize = 16

// Scale is the pixel-art upscale factor.
const Scale = 3

// ScaledTile is the on-screen pixel size of one tile (TileSize * Scale).
const ScaledTile = TileSize * Scale

// TileID identifies a type of tile. 0 is always empty (air).
type TileID uint8

const (
	TileEmpty  TileID = 0
	TileBricks TileID = 1
)

// TileDef describes the properties of a registered tile type.
type TileDef struct {
	ID       TileID
	Name     string
	Texture  *sdl.Texture
	Solid    bool // whether the tile blocks movement
	Deadly   bool // whether the tile kills the player (e.g. spikes)
}

// Registry maps TileIDs to their definitions.
var Registry = map[TileID]*TileDef{}

// Register adds a new tile definition to the global registry.
func Register(def *TileDef) {
	if _, exists := Registry[def.ID]; exists {
		panic(fmt.Sprintf("tilemap: duplicate registration for TileID %d", def.ID))
	}
	Registry[def.ID] = def
}

// LoadAssets loads tile textures and registers built-in tile types.
// Call this after SDL and the renderer are initialised.
func LoadAssets(renderer *sdl.Renderer) {
	// Bricks
	bricksTex, err := img.LoadTexture(renderer, "assets/textures/bricks.png")
	if err != nil {
		panic("failed to load bricks texture: " + err.Error())
	}
	Register(&TileDef{
		ID:      TileBricks,
		Name:    "bricks",
		Texture: bricksTex,
		Solid:   true,
	})
}

// Tilemap holds a 2-D grid of TileIDs and provides rendering.
type Tilemap struct {
	Width  int // number of tiles horizontally
	Height int // number of tiles vertically
	Tiles  [][]TileID
}

// New creates a Tilemap of the given dimensions, filled with TileEmpty.
func New(width, height int) *Tilemap {
	tiles := make([][]TileID, height)
	for y := range tiles {
		tiles[y] = make([]TileID, width)
	}
	return &Tilemap{Width: width, Height: height, Tiles: tiles}
}

// Set sets a single tile. Out-of-range calls are silently ignored.
func (tm *Tilemap) Set(x, y int, id TileID) {
	if x < 0 || x >= tm.Width || y < 0 || y >= tm.Height {
		return
	}
	tm.Tiles[y][x] = id
}

// Get returns the TileID at (x, y). Out-of-range returns TileEmpty.
func (tm *Tilemap) Get(x, y int) TileID {
	if x < 0 || x >= tm.Width || y < 0 || y >= tm.Height {
		return TileEmpty
	}
	return tm.Tiles[y][x]
}

// IsSolid reports whether the tile at (x, y) blocks movement.
func (tm *Tilemap) IsSolid(x, y int) bool {
	id := tm.Get(x, y)
	if id == TileEmpty {
		return false
	}
	def, ok := Registry[id]
	if !ok {
		return false
	}
	return def.Solid
}

// IsDeadly reports whether the tile at (x, y) kills the player.
func (tm *Tilemap) IsDeadly(x, y int) bool {
	id := tm.Get(x, y)
	if id == TileEmpty {
		return false
	}
	def, ok := Registry[id]
	if !ok {
		return false
	}
	return def.Deadly
}

// Render draws the visible portion of the tilemap.
// camX, camY are the camera offset in pixels.
func (tm *Tilemap) Render(renderer *sdl.Renderer, camX, camY float32) {
	screenW := float32(960) // window width
	screenH := float32(720) // window height

	// Compute which tile columns/rows are visible.
	startX := int(camX / float32(ScaledTile))
	startY := int(camY / float32(ScaledTile))
	endX := int((camX + screenW) / float32(ScaledTile)) + 1
	endY := int((camY + screenH) / float32(ScaledTile)) + 1

	// Clamp to map bounds.
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if endX > tm.Width {
		endX = tm.Width
	}
	if endY > tm.Height {
		endY = tm.Height
	}

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			id := tm.Tiles[y][x]
			if id == TileEmpty {
				continue
			}
			def, ok := Registry[id]
			if !ok || def.Texture == nil {
				continue
			}
			dst := sdl.FRect{
				X: float32(x)*float32(ScaledTile) - camX,
				Y: float32(y)*float32(ScaledTile) - camY,
				W: float32(ScaledTile),
				H: float32(ScaledTile),
			}
			renderer.RenderTexture(def.Texture, nil, &dst)
		}
	}
}
