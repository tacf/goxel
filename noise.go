package main

// hash mixes a 2D integer coordinate and a seed into 32 random bits.
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

// valueNoise is smooth noise in 0..1. When nx and ny are positive the lattice
// wraps around after that many cells, so the result tiles seamlessly.
// Zero means no wrapping.
func valueNoise(x, y float32, nx, ny int, seed uint32) float32 {
	x0, y0 := floor(x), floor(y)
	ix, iy := int(x0), int(y0)
	fx, fy := x-x0, y-y0
	u := fx * fx * (3 - 2*fx)
	v := fy * fy * (3 - 2*fy)
	xa, xb, ya, yb := ix, ix+1, iy, iy+1
	if nx > 0 {
		xa, xb = wrap(xa, nx), wrap(xb, nx)
	}
	if ny > 0 {
		ya, yb = wrap(ya, ny), wrap(yb, ny)
	}
	top := lerp(rnd(xa, ya, seed), rnd(xb, ya, seed), u)
	bottom := lerp(rnd(xa, yb, seed), rnd(xb, yb, seed), u)
	return lerp(top, bottom, v)
}

// tilingFBM sums octaves of wrapping noise over the unit square.
// Each octave doubles the lattice, starting from the given number of cells.
func tilingFBM(u, v float32, cells, octaves int, seed uint32) float32 {
	var sum, norm float32
	amp := float32(0.5)
	n := cells
	for o := range octaves {
		f := float32(n)
		sum += valueNoise(u*f, v*f, n, n, seed+uint32(o)) * amp
		norm += amp
		amp *= 0.5
		n *= 2
	}
	return sum / norm
}
