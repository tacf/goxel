package main

const sandLevel = 22

// terrainFBM is non wrapping fractal noise for the height map.
func terrainFBM(x, z float32, seed uint32) float32 {
	var sum, norm float32
	amp, freq := float32(0.5), float32(1)
	for o := range 4 {
		sum += valueNoise(x*freq, z*freq, 0, 0, seed+uint32(o)) * amp
		norm += amp
		amp *= 0.5
		freq *= 2
	}
	return sum / norm
}

func columnHeight(x, z int, seed uint32) int {
	n := terrainFBM(float32(x)*0.008, float32(z)*0.008, seed)
	h := int(8 + n*44)
	return min(max(h, 2), Height-14)
}

func (w *World) generate(seed uint32) {
	for z := range SizeXZ {
		for x := range SizeXZ {
			h := columnHeight(x, z, seed)
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
