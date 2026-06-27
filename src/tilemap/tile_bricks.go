package tilemap

import (
	"github.com/Zyko0/go-sdl3/img"
	"github.com/Zyko0/go-sdl3/sdl"
)

func registerBricks(renderer *sdl.Renderer) {
	tex, err := img.LoadTexture(renderer, "assets/textures/bricks.png")
	if err != nil {
		panic("failed to load bricks texture: " + err.Error())
	}
	Register(&TileDef{
		ID:      TileBricks,
		Name:    "bricks",
		Texture: tex,
		Solid:   true,
	})
}
