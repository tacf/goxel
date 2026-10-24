package main

import "goxel/internal/fastnoise"

// hash mixes a 2D integer coordinate and a seed into 32 random bits.
// It drives per pixel speckles, tree placement and per block brightness.
func hash(x, y int, seed uint32) uint32 {
	h := seed ^ (uint32(int32(x)) * 0x27d4eb2d) ^ (uint32(int32(y)) * 0x165667b1)
	h ^= h >> 15
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

// rnd is a deterministic random value in 0..1 for a coordinate.
func rnd(x, y int, seed uint32) float32 {
	return float32(hash(x, y, seed)>>8) / (1 << 24)
}

func smoothstep(t float32) float32 { return t * t * (3 - 2*t) }

// tileNoise is FastNoiseLite value noise that repeats every unit tile.
//
// FastNoiseLite has no tiling mode, so four copies of the noise, each shifted
// by one tile, are cross faded with smooth weights. The weights are zero or
// one on opposite edges, which makes the result seamless. The sum is rescaled
// so the contrast does not drop towards the middle of the tile.
type tileNoise struct {
	state *fastnoise.State[float32]
	// cells is the number of lattice cells across the tile in x and y.
	cellsX, cellsY float32
}

func newTileNoise(seed int, cellsX, cellsY float32, octaves int) *tileNoise {
	s := fastnoise.New[float32]()
	s.Seed = seed
	s.Frequency = 1
	s.NoiseType(fastnoise.Value)
	if octaves > 1 {
		s.FractalType(fastnoise.FractalFBm)
		s.Octaves = octaves
	}
	return &tileNoise{s, cellsX, cellsY}
}

// at samples the noise at a position in 0..1 tile coordinates and returns 0..1.
func (t *tileNoise) at(u, v float32) float32 {
	su, sv := smoothstep(u), smoothstep(v)
	wu := [2]float32{1 - su, su}
	wv := [2]float32{1 - sv, sv}
	x, y := u*t.cellsX, v*t.cellsY

	var sum float32
	for j := range 2 {
		for i := range 2 {
			sum += wu[i] * wv[j] * t.state.GetNoise2D(x-float32(i)*t.cellsX, y-float32(j)*t.cellsY)
		}
	}
	// The copies are independent, so their variances add up to the sum of squared weights.
	energy := (wu[0]*wu[0] + wu[1]*wu[1]) * (wv[0]*wv[0] + wv[1]*wv[1])
	return clamp(sum/sqrt(energy)*0.5+0.5, 0, 1)
}
