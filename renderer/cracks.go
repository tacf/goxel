package renderer

import (
	"image/color"

	. "github.com/tacf/goxel/math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const crackStages = 10

// Cracks holds the crack overlay texture of every damage stage.
type Cracks struct {
	textures [crackStages]rl.Texture2D
}

// LoadCracks builds the overlay textures. Each stage extends the cracks of the previous one.
func LoadCracks() *Cracks {
	var birth Birth
	generateCracks(&birth, 0x2545f491)

	c := &Cracks{}
	for s := range crackStages {
		reach := float32(s+1) / crackStages
		img := rl.GenImageColor(crackTile, crackTile, color.RGBA{})
		for y := range crackTile {
			for x := range crackTile {
				born := birth[y][x]
				if born > reach {
					continue
				}
				// Older parts of a crack read darker and deeper.
				age := Clamp((reach-born)/0.5, 0, 1)
				shade := uint8(30 - 24*age)
				alpha := uint8(175 + 75*age)
				rl.ImageDrawPixel(img, int32(x), int32(y), color.RGBA{shade, shade, shade, alpha})
			}
		}
		c.textures[s] = rl.LoadTextureFromImage(img)
		rl.UnloadImage(img)
		rl.SetTextureFilter(c.textures[s], rl.FilterPoint)
	}
	return c
}

func (c *Cracks) Unload() {
	for _, t := range c.textures {
		rl.UnloadTexture(t)
	}
}

// Draw renders the crack overlay for a mining progress in 0..1 around the block.
func (c *Cracks) Draw(pos [3]int, progress float32) {
	stage := min(crackStages-1, int(max(progress, 0)*crackStages))
	o := [3]float32{float32(pos[0]), float32(pos[1]), float32(pos[2])}
	uv := [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}}
	const grow = 0.002
	rl.SetTexture(c.textures[stage].ID)
	rl.Begin(rl.Quads)
	rl.Color4ub(255, 255, 255, 255)
	for fi := range VoxelFaces {
		for k, v := range VoxelFaces[fi].V {
			rl.TexCoord2f(uv[k][0], uv[k][1])
			rl.Vertex3f(
				o[0]+float32(v[0])*(1+2*grow)-grow,
				o[1]+float32(v[1])*(1+2*grow)-grow,
				o[2]+float32(v[2])*(1+2*grow)-grow,
			)
		}
	}
	rl.End()
	rl.SetTexture(0)
}
