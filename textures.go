package goxel

import . "github.com/tacf/goxel/math"

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
func TileRect(t Tile) Rect {
	const step = 1.0 / AtlasCols
	col := float32(int(t) % AtlasCols)
	row := float32(int(t) / AtlasCols)
	return Rect{col * step, row * step, (col + 1) * step, (row + 1) * step}
}

// tileOrigin is the top left pixel of a tile in the atlas.
func tileOrigin(t Tile) (x, y int) {
	return int(t) % AtlasCols * TileSize, int(t) / AtlasCols * TileSize
}

// tilePixels is the pixel rectangle of a tile (x, y, w, h), for 2D drawing.
func tilePixels(t Tile) [4]float32 {
	x, y := tileOrigin(t)
	return [4]float32{float32(x), float32(y), TileSize, TileSize}
}

// tileLayers cuts the atlas into one RGBA image per tile, in tile order.
// Layer i belongs to Tile(i) and is what the tile array texture stores.
func tileLayers(atlas []byte) []byte {
	const row = TileSize * 4
	layers := make([]byte, 0, int(tileCount)*TileSize*row)
	for t := range tileCount {
		ox, oy := tileOrigin(t)
		for y := range TileSize {
			start := ((oy+y)*AtlasSize + ox) * 4
			layers = append(layers, atlas[start:start+row]...)
		}
	}
	return layers
}

type color3 [3]float32

// texGen holds the noise generators the tile painters draw from.
type texGen struct {
	grass, blade, dirt, stone, vein        *TileNoise
	warpX, warpY, grain                    *TileNoise
	sand, streak, bark, rings, plank, leaf *TileNoise
}

func newTexGen() *texGen {
	return &texGen{
		grass: NewTileNoise(11, 2, 2, 3),
		// Stretched vertically so it reads as blades of grass.
		blade: NewTileNoise(12, 16, 4, 1),
		dirt:  NewTileNoise(21, 4, 4, 2),
		stone: NewTileNoise(41, 4, 4, 3),
		vein:  NewTileNoise(42, 3, 3, 2),
		warpX: NewTileNoise(51, 3, 3, 2),
		warpY: NewTileNoise(52, 3, 3, 2),
		grain: NewTileNoise(55, 8, 8, 2),
		sand:  NewTileNoise(61, 8, 8, 2),
		// Stretched vertically for the streaks of bark.
		streak: NewTileNoise(71, 8, 2, 1),
		bark:   NewTileNoise(72, 4, 4, 2),
		rings:  NewTileNoise(74, 4, 4, 2),
		// Stretched horizontally for the grain along a board.
		plank: NewTileNoise(82, 2, 16, 1),
		leaf:  NewTileNoise(91, 4, 4, 2),
	}
}

// generateAtlas fills an RGBA atlas with every procedural tile.
func generateAtlas(pixels []byte) {
	clear(pixels)
	g := newTexGen()
	for t := range tileCount {
		ox, oy := tileOrigin(t)
		for py := range TileSize {
			for px := range TileSize {
				c := g.tilePixel(t, px, py)
				at := ((oy+py)*AtlasSize + ox + px) * 4
				for ch := range 3 {
					pixels[at+ch] = byte(Clamp(c[ch], 0, 255))
				}
				pixels[at+3] = 255
			}
		}
	}
}

func (g *texGen) tilePixel(t Tile, px, py int) color3 {
	switch t {
	case TileGrassTop:
		return g.grassTop(px, py)
	case TileGrassSide:
		return g.grassSide(px, py)
	case TileDirt:
		return g.dirtTile(px, py)
	case TileStone:
		return g.stoneTile(px, py)
	case TileCobble:
		return g.cobbleTile(px, py)
	case TileSand:
		return g.sandTile(px, py)
	case TileWoodSide:
		return g.woodSide(px, py)
	case TileWoodTop:
		return g.woodTop(px, py)
	case TilePlanks:
		return g.planksTile(px, py)
	default:
		return g.leavesTile(px, py)
	}
}

// texel is the centre of a tile pixel in 0..1 tile coordinates.
func texel(px, py int) (float32, float32) {
	return (float32(px) + 0.5) / TileSize, (float32(py) + 0.5) / TileSize
}

func scale(c color3, f float32) color3 { return color3{c[0] * f, c[1] * f, c[2] * f} }

func mix(a, b color3, t float32) color3 {
	return color3{Lerp(a[0], b[0], t), Lerp(a[1], b[1], t), Lerp(a[2], b[2], t)}
}

// Random bright and dark pixels on top of a base colour.
func speckle(c color3, px, py int, seed uint32, dark, light float32) color3 {
	s := Rnd(px, py, seed)
	if s < 0.07 {
		return scale(c, dark)
	}
	if s > 0.94 {
		return scale(c, light)
	}
	return c
}

func (g *texGen) grassTop(px, py int) color3 {
	u, v := texel(px, py)
	n := g.grass.At(u, v)
	c := mix(color3{60, 112, 34}, color3{118, 178, 64}, n)
	// Short vertical streaks read as blades.
	blade := g.blade.At(u, v)
	c = scale(c, 0.86+0.28*blade)
	return speckle(c, px, py, 13, 0.78, 1.15)
}

func (g *texGen) dirtTile(px, py int) color3 {
	u, v := texel(px, py)
	n := g.dirt.At(u, v)
	c := scale(color3{134, 96, 67}, 0.82+0.34*n)
	return speckle(c, px, py, 22, 0.72, 1.18)
}

func (g *texGen) grassSide(px, py int) color3 {
	// Ragged grass fringe over dirt.
	fringe := 3 + int(Hash(px, 0, 31)%3)
	if py < fringe {
		c := g.grassTop(px, py)
		if py == fringe-1 {
			return scale(c, 0.9)
		}
		return c
	}
	return g.dirtTile(px, py)
}

func (g *texGen) stoneTile(px, py int) color3 {
	u, v := texel(px, py)
	n := g.stone.At(u, v)
	c := scale(color3{126, 126, 130}, 0.8+0.4*n)
	// Thin dark veins where a second noise field crosses its midpoint.
	ridge := Abs(g.vein.At(u, v) - 0.5)
	if ridge < 0.03 {
		c = scale(c, 0.76)
	}
	return speckle(c, px, py, 43, 0.85, 1.12)
}

const cobbleCells = 3

func (g *texGen) cobbleTile(px, py int) color3 {
	u, v := texel(px, py)
	// Warp the sample point so stone outlines are irregular.
	wx := (g.warpX.At(u, v) - 0.5) * 0.35
	wy := (g.warpY.At(u, v) - 0.5) * 0.35
	fx := u*cobbleCells + wx
	fy := v*cobbleCells + wy
	cx, cy := int(Floor(fx)), int(Floor(fy))

	// Worley noise: distance to the nearest and second nearest feature point.
	f1, f2 := float32(99), float32(99)
	id := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			gx, gy := cx+dx, cy+dy
			wrapx, wrapy := Wrap(gx, cobbleCells), Wrap(gy, cobbleCells)
			jx := 0.15 + 0.7*Rnd(wrapx, wrapy, 53)
			jy := 0.15 + 0.7*Rnd(wrapx, wrapy, 54)
			ddx := float32(gx) + jx - fx
			ddy := float32(gy) + jy - fy
			d := Sqrt(ddx*ddx + ddy*ddy)
			if d < f1 {
				f2 = f1
				f1 = d
				id = wrapy*cobbleCells + wrapx
			} else if d < f2 {
				f2 = d
			}
		}
	}

	grain := 0.88 + 0.24*g.grain.At(u, v)
	edge := f2 - f1
	if edge < 0.2 {
		return scale(color3{58, 58, 62}, grain)
	}
	// Every stone gets its own tone, lit brighter towards its middle.
	tone := 92 + 62*Rnd(id, 7, 56)
	bevel := 0.8 + 0.3*min(1, edge*1.6)
	return speckle(scale(color3{tone, tone, tone + 4}, bevel*grain), px, py, 57, 0.85, 1.1)
}

func (g *texGen) sandTile(px, py int) color3 {
	u, v := texel(px, py)
	n := g.sand.At(u, v)
	c := scale(color3{219, 207, 142}, 0.93+0.12*n)
	return speckle(c, px, py, 62, 0.9, 1.05)
}

func (g *texGen) woodSide(px, py int) color3 {
	u, v := texel(px, py)
	streak := g.streak.At(u, v)
	n := g.bark.At(u, v)
	c := scale(color3{104, 82, 50}, 0.72+0.34*streak+0.18*n)
	// Dark bark grooves.
	if Hash(px, 0, 73)%5 == 0 {
		c = scale(c, 0.78)
	}
	return c
}

func (g *texGen) woodTop(px, py int) color3 {
	u, v := texel(px, py)
	dx, dy := u-0.5, v-0.5
	d := Sqrt(dx*dx+dy*dy) * TileSize
	n := g.rings.At(u, v)
	if d > 6.4 {
		return scale(color3{104, 82, 50}, 0.8+0.3*n)
	}
	// Growth rings, bent a little by noise.
	ring := 0.5 + 0.5*Sin(d*1.8+n*3.5)
	return mix(color3{140, 108, 64}, color3{184, 150, 96}, ring)
}

func (g *texGen) planksTile(px, py int) color3 {
	u, v := texel(px, py)
	row := py / 4
	inside := py % 4
	c := scale(color3{178, 144, 90}, 0.86+0.22*Rnd(row, 0, 81))
	// Grain runs along each board.
	grain := g.plank.At(u, v)
	c = scale(c, 0.9+0.2*grain)
	if inside == 3 {
		c = scale(c, 0.68)
	}
	// One vertical joint per board at a random column.
	if px == int(Hash(row, 0, 83)%TileSize) {
		c = scale(c, 0.78)
	}
	return c
}

func (g *texGen) leavesTile(px, py int) color3 {
	u, v := texel(px, py)
	n := g.leaf.At(u, v)
	c := mix(color3{34, 92, 26}, color3{82, 164, 54}, n)
	s := Rnd(px, py, 92)
	if s < 0.18 {
		c = color3{22, 64, 20}
	}
	if s > 0.9 {
		c = color3{100, 184, 66}
	}
	return c
}
