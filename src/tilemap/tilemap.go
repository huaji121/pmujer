// Package tilemap provides an extensible tile-based map system for the game.
package tilemap

import (
	"fmt"
	"math"

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
	TileSpike  TileID = 2
)

// ConvexHitbox is a convex polygon defined by vertices in counter-clockwise
// order, in tile-local coordinates ([0,1] range).  Used for precise collision
// with deadly tiles.
type ConvexHitbox []Vec2

// Vec2 is a 2D point/vector.
type Vec2 struct {
	X, Y float32
}

// TileDef describes the properties of a registered tile type.
type TileDef struct {
	ID       TileID
	Name     string
	Texture  *sdl.Texture
	Solid    bool          // whether the tile blocks movement
	Deadly   bool          // whether the tile kills the player (e.g. spikes)
	Hitbox   ConvexHitbox  // convex polygon hitbox (nil = use full tile AABB)
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
// Each tile type is defined in its own file (tile_*.go).
func LoadAssets(renderer *sdl.Renderer) {
	registerBricks(renderer)
	registerSpike(renderer)
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

// HitDeadly reports whether an AABB (rx, ry, rw, rh) in tile-units overlaps
// any deadly tile's convex hitbox.
func (tm *Tilemap) HitDeadly(rx, ry, rw, rh float32) bool {
	left := int(math.Floor(float64(rx)))
	right := int(math.Floor(float64(rx + rw - 0.001)))
	top := int(math.Floor(float64(ry)))
	bottom := int(math.Floor(float64(ry + rh - 0.001)))

	for ty := top; ty <= bottom; ty++ {
		for tx := left; tx <= right; tx++ {
			id := tm.Get(tx, ty)
			if id == TileEmpty {
				continue
			}
			def, ok := Registry[id]
			if !ok || !def.Deadly {
				continue
			}
			// Transform AABB into tile-local coords [0,1].
			lx := rx - float32(tx)
			ly := ry - float32(ty)
			if def.Hitbox != nil {
				if aabbConvexOverlap(lx, ly, rw, rh, def.Hitbox) {
					return true
				}
			} else {
				// No hitbox → full tile AABB.
				if lx+rw > 0 && lx < 1 && ly+rh > 0 && ly < 1 {
					return true
				}
			}
		}
	}
	return false
}

// --- AABB vs convex polygon (SAT) ---

// aabbConvexOverlap tests whether an AABB (rx, ry)→(rx+rw, ry+rh) overlaps a
// convex polygon defined by CCW vertices.
func aabbConvexOverlap(rx, ry, rw, rh float32, poly ConvexHitbox) bool {
	rx2 := rx + rw
	ry2 := ry + rh
	n := len(poly)

	// Test each polygon edge normal as a separating axis.
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		// Edge vector
		ex := poly[j].X - poly[i].X
		ey := poly[j].Y - poly[i].Y
		// Outward normal (CCW winding → right-hand normal)
		nx := ey
		ny := -ex

		// Project polygon onto normal
		pMin := nx*poly[0].X + ny*poly[0].Y
		pMax := pMin
		for k := 1; k < n; k++ {
			p := nx*poly[k].X + ny*poly[k].Y
			if p < pMin {
				pMin = p
			}
			if p > pMax {
				pMax = p
			}
		}

		// Project AABB onto normal (only the two extreme corners matter)
		corners := [4][2]float32{{rx, ry}, {rx2, ry}, {rx2, ry2}, {rx, ry2}}
		bMin := nx*corners[0][0] + ny*corners[0][1]
		bMax := bMin
		for k := 1; k < 4; k++ {
			b := nx*corners[k][0] + ny*corners[k][1]
			if b < bMin {
				bMin = b
			}
			if b > bMax {
				bMax = b
			}
		}

		// Separating axis found?
		if bMax <= pMin || bMin >= pMax {
			return false
		}
	}

	// Test AABB axes (X and Y)
	// X axis
	pMin := poly[0].X
	pMax := poly[0].X
	for k := 1; k < n; k++ {
		if poly[k].X < pMin {
			pMin = poly[k].X
		}
		if poly[k].X > pMax {
			pMax = poly[k].X
		}
	}
	if rx2 <= pMin || rx >= pMax {
		return false
	}
	// Y axis
	pMin = poly[0].Y
	pMax = poly[0].Y
	for k := 1; k < n; k++ {
		if poly[k].Y < pMin {
			pMin = poly[k].Y
		}
		if poly[k].Y > pMax {
			pMax = poly[k].Y
		}
	}
	if ry2 <= pMin || ry >= pMax {
		return false
	}

	return true
}

// Render draws the visible portion of the tilemap.
// camX, camY are the camera offset in world-pixels; zoom is the scale factor.
func (tm *Tilemap) Render(renderer *sdl.Renderer, camX, camY, zoom float32, debug bool) {
	screenW := float32(960) // window width
	screenH := float32(720) // window height
	s := float32(ScaledTile)

	// Visible area in world-pixels (larger when zoomed out).
	visW := screenW / zoom
	visH := screenH / zoom

	// Compute which tile columns/rows are visible.
	startX := int(camX / s)
	startY := int(camY / s)
	endX := int((camX+visW)/s) + 1
	endY := int((camY+visH)/s) + 1

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

	tilePx := s * zoom // rendered tile size in screen pixels

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
				X: (float32(x)*s - camX) * zoom,
				Y: (float32(y)*s - camY) * zoom,
				W: tilePx,
				H: tilePx,
			}
			renderer.RenderTexture(def.Texture, nil, &dst)

			// Debug: draw hitbox outline (red)
			if debug && def.Hitbox != nil {
				ox := (float32(x)*s - camX) * zoom
				oy := (float32(y)*s - camY) * zoom
				renderer.SetDrawColor(255, 0, 0, 255)
				for i := 0; i < len(def.Hitbox); i++ {
					j := (i + 1) % len(def.Hitbox)
					renderer.RenderLine(
						ox+def.Hitbox[i].X*tilePx, oy+def.Hitbox[i].Y*tilePx,
						ox+def.Hitbox[j].X*tilePx, oy+def.Hitbox[j].Y*tilePx,
					)
				}
			}
		}
	}
}
