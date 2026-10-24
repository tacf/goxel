package main

const (
	TileSize   = 16
	AtlasCols  = 4
	AtlasSize  = TileSize * AtlasCols
	AtlasBytes = AtlasSize * AtlasSize * 4
)

type Tile uint8

const (
	TileGrassTop Tile = iota
	TileGrassSide
	TileDirt
	TileStone
	TileCobble
	TileSand
	TileWoodSide
	TileWoodTop
	TilePlanks
	TileLeaves
	tileCount
)

type Rect struct{ U0, V0, U1, V1 float32 }

// Normalised atlas coordinates of a tile.
func tileRect(t Tile) Rect {
	const step = 1.0 / AtlasCols
	col := float32(int(t) % AtlasCols)
	row := float32(int(t) / AtlasCols)
	return Rect{col * step, row * step, (col + 1) * step, (row + 1) * step}
}

// Pixel coordinates of a tile (x, y, w, h), for 2D drawing.
func tilePixels(t Tile) [4]float32 {
	return [4]float32{
		float32(int(t) % AtlasCols * TileSize),
		float32(int(t) / AtlasCols * TileSize),
		TileSize, TileSize,
	}
}

type color3 [3]float32

// generateAtlas fills an RGBA atlas with every procedural tile.
func generateAtlas(pixels []byte) {
	clear(pixels)
	for t := range tileCount {
		ox := int(t) % AtlasCols * TileSize
		oy := int(t) / AtlasCols * TileSize
		for py := range TileSize {
			for px := range TileSize {
				c := tilePixel(t, px, py)
				at := ((oy+py)*AtlasSize + ox + px) * 4
				for ch := range 3 {
					pixels[at+ch] = byte(clamp(c[ch], 0, 255))
				}
				pixels[at+3] = 255
			}
		}
	}
}

func tilePixel(t Tile, px, py int) color3 {
	switch t {
	case TileGrassTop:
		return grassTop(px, py)
	case TileGrassSide:
		return grassSide(px, py)
	case TileDirt:
		return dirtTile(px, py)
	case TileStone:
		return stoneTile(px, py)
	case TileCobble:
		return cobbleTile(px, py)
	case TileSand:
		return sandTile(px, py)
	case TileWoodSide:
		return woodSide(px, py)
	case TileWoodTop:
		return woodTop(px, py)
	case TilePlanks:
		return planksTile(px, py)
	default:
		return leavesTile(px, py)
	}
}

// texel is the centre of a tile pixel in 0..1 tile coordinates.
func texel(px, py int) (float32, float32) {
	return (float32(px) + 0.5) / TileSize, (float32(py) + 0.5) / TileSize
}

func scale(c color3, f float32) color3 { return color3{c[0] * f, c[1] * f, c[2] * f} }

func mix(a, b color3, t float32) color3 {
	return color3{lerp(a[0], b[0], t), lerp(a[1], b[1], t), lerp(a[2], b[2], t)}
}

// Random bright and dark pixels on top of a base colour.
func speckle(c color3, px, py int, seed uint32, dark, light float32) color3 {
	s := rnd(px, py, seed)
	if s < 0.07 {
		return scale(c, dark)
	}
	if s > 0.94 {
		return scale(c, light)
	}
	return c
}

func grassTop(px, py int) color3 {
	u, v := texel(px, py)
	n := tilingFBM(u, v, 2, 3, 11)
	c := mix(color3{60, 112, 34}, color3{118, 178, 64}, n)
	// Short vertical streaks read as blades.
	blade := valueNoise(u*16, v*4, 16, 4, 12)
	c = scale(c, 0.86+0.28*blade)
	return speckle(c, px, py, 13, 0.78, 1.15)
}

func dirtTile(px, py int) color3 {
	u, v := texel(px, py)
	n := tilingFBM(u, v, 4, 2, 21)
	c := scale(color3{134, 96, 67}, 0.82+0.34*n)
	return speckle(c, px, py, 22, 0.72, 1.18)
}

func grassSide(px, py int) color3 {
	// Ragged grass fringe over dirt.
	fringe := 3 + int(hash(px, 0, 31)%3)
	if py < fringe {
		c := grassTop(px, py)
		if py == fringe-1 {
			return scale(c, 0.9)
		}
		return c
	}
	return dirtTile(px, py)
}

func stoneTile(px, py int) color3 {
	u, v := texel(px, py)
	n := tilingFBM(u, v, 4, 3, 41)
	c := scale(color3{126, 126, 130}, 0.8+0.4*n)
	// Thin dark veins where a second noise field crosses its midpoint.
	ridge := abs(tilingFBM(u, v, 3, 2, 42) - 0.5)
	if ridge < 0.03 {
		c = scale(c, 0.76)
	}
	return speckle(c, px, py, 43, 0.85, 1.12)
}

const cobbleCells = 3

func cobbleTile(px, py int) color3 {
	u, v := texel(px, py)
	// Warp the sample point so stone outlines are irregular.
	wx := (tilingFBM(u, v, 3, 2, 51) - 0.5) * 0.35
	wy := (tilingFBM(u, v, 3, 2, 52) - 0.5) * 0.35
	fx := u*cobbleCells + wx
	fy := v*cobbleCells + wy
	cx, cy := int(floor(fx)), int(floor(fy))

	// Worley noise: distance to the nearest and second nearest feature point.
	f1, f2 := float32(99), float32(99)
	id := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			gx, gy := cx+dx, cy+dy
			wrapx, wrapy := wrap(gx, cobbleCells), wrap(gy, cobbleCells)
			jx := 0.15 + 0.7*rnd(wrapx, wrapy, 53)
			jy := 0.15 + 0.7*rnd(wrapx, wrapy, 54)
			ddx := float32(gx) + jx - fx
			ddy := float32(gy) + jy - fy
			d := sqrt(ddx*ddx + ddy*ddy)
			if d < f1 {
				f2 = f1
				f1 = d
				id = wrapy*cobbleCells + wrapx
			} else if d < f2 {
				f2 = d
			}
		}
	}

	grain := 0.88 + 0.24*tilingFBM(u, v, 8, 2, 55)
	edge := f2 - f1
	if edge < 0.2 {
		return scale(color3{58, 58, 62}, grain)
	}
	// Every stone gets its own tone, lit brighter towards its middle.
	tone := 92 + 62*rnd(id, 7, 56)
	bevel := 0.8 + 0.3*min(1, edge*1.6)
	return speckle(scale(color3{tone, tone, tone + 4}, bevel*grain), px, py, 57, 0.85, 1.1)
}

func sandTile(px, py int) color3 {
	u, v := texel(px, py)
	n := tilingFBM(u, v, 8, 2, 61)
	c := scale(color3{219, 207, 142}, 0.93+0.12*n)
	return speckle(c, px, py, 62, 0.9, 1.05)
}

func woodSide(px, py int) color3 {
	u, v := texel(px, py)
	streak := valueNoise(u*8, v*2, 8, 2, 71)
	n := tilingFBM(u, v, 4, 2, 72)
	c := scale(color3{104, 82, 50}, 0.72+0.34*streak+0.18*n)
	// Dark bark grooves.
	if hash(px, 0, 73)%5 == 0 {
		c = scale(c, 0.78)
	}
	return c
}

func woodTop(px, py int) color3 {
	u, v := texel(px, py)
	dx, dy := u-0.5, v-0.5
	d := sqrt(dx*dx+dy*dy) * TileSize
	n := tilingFBM(u, v, 4, 2, 74)
	if d > 6.4 {
		return scale(color3{104, 82, 50}, 0.8+0.3*n)
	}
	// Growth rings, bent a little by noise.
	ring := 0.5 + 0.5*sin(d*1.8+n*3.5)
	return mix(color3{140, 108, 64}, color3{184, 150, 96}, ring)
}

func planksTile(px, py int) color3 {
	u, v := texel(px, py)
	row := py / 4
	inside := py % 4
	c := scale(color3{178, 144, 90}, 0.86+0.22*rnd(row, 0, 81))
	// Grain runs along each board.
	grain := valueNoise(u*2, v*16, 2, 16, 82)
	c = scale(c, 0.9+0.2*grain)
	if inside == 3 {
		c = scale(c, 0.68)
	}
	// One vertical joint per board at a random column.
	if px == int(hash(row, 0, 83)%TileSize) {
		c = scale(c, 0.78)
	}
	return c
}

func leavesTile(px, py int) color3 {
	u, v := texel(px, py)
	n := tilingFBM(u, v, 4, 2, 91)
	c := mix(color3{34, 92, 26}, color3{82, 164, 54}, n)
	s := rnd(px, py, 92)
	if s < 0.18 {
		c = color3{22, 64, 20}
	}
	if s > 0.9 {
		c = color3{100, 184, 66}
	}
	return c
}
