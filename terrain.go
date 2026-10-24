package main

import "goxel/internal/fastnoise"

const sandLevel = 23

// terrain is the height map, fractal OpenSimplex2 noise from FastNoiseLite.
type terrain struct{ noise *fastnoise.State[float32] }

func newTerrain(seed uint32) terrain {
	n := fastnoise.New[float32]()
	n.Seed = int(seed)
	n.Frequency = 0.008
	n.NoiseType(fastnoise.OpenSimplex2)
	n.FractalType(fastnoise.FractalFBm)
	n.Octaves = 4
	return terrain{n}
}

// height is the y of the top block of a column.
func (t terrain) height(x, z int) int {
	n := t.noise.GetNoise2D(float32(x), float32(z))*0.5 + 0.5
	h := int(14 + n*36)
	return min(max(h, 2), Height-14)
}

func (w *World) generate(seed uint32) {
	height := newTerrain(seed)
	for z := range SizeXZ {
		for x := range SizeXZ {
			h := height.height(x, z)
			beach := h <= sandLevel
			for y := 0; y <= h; y++ {
				var b Block
				switch {
				case y < h-3:
					b = Stone
				case beach:
					b = Sand
				case y == h:
					b = Grass
				default:
					b = Dirt
				}
				w.blocks[index(x, y, z)] = b
			}
		}
	}

	// Sparse trees on grass, kept away from the world border.
	for z := 3; z < SizeXZ-3; z++ {
		for x := 3; x < SizeXZ-3; x++ {
			h := w.SurfaceHeight(x, z)
			if w.Get(x, h, z) != Grass {
				continue
			}
			r := hash(x, z, seed+999)
			if r%90 != 0 {
				continue
			}
			w.placeTree(x, h, z, 4+int((r>>8)%2))
		}
	}
}

func (w *World) placeTree(x, y, z, trunk int) {
	for ty := 1; ty <= trunk; ty++ {
		w.Set(x, y+ty, z, Wood)
	}
	top := y + trunk
	for ly := top - 2; ly <= top+1; ly++ {
		r := 2
		if ly >= top {
			r = 1
		}
		for dz := -r; dz <= r; dz++ {
			for dx := -r; dx <= r; dx++ {
				// Round off the corners of each leaf layer.
				if dx*dx == r*r && dz*dz == r*r {
					continue
				}
				if w.Get(x+dx, ly, z+dz) == Air {
					w.Set(x+dx, ly, z+dz, Leaves)
				}
			}
		}
	}
}
