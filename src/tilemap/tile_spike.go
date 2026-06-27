package tilemap

import (
	"github.com/Zyko0/go-sdl3/img"
	"github.com/Zyko0/go-sdl3/sdl"
)

func registerSpike(renderer *sdl.Renderer) {
	tex, err := img.LoadTexture(renderer, "assets/textures/spike.png")
	if err != nil {
		panic("failed to load spike texture: " + err.Error())
	}
	Register(&TileDef{
		ID:      TileSpike,
		Name:    "spike",
		Texture: tex,
		Deadly:  true,
		// Upward-pointing triangle: bottom-left → top-centre → bottom-right (CCW)
		Hitbox: ConvexHitbox{
			{X: 0, Y: 1},
			{X: 0.5, Y: 0},
			{X: 1, Y: 1},
		},
	})
}
