package renderer

import (
	"bytes"
	"testing"
)

func TestAtlasTilesAreOpaqueAndNotFlat(t *testing.T) {
	pixels := make([]byte, AtlasBytes)
	generateAtlas(pixels)
	for tile := range tileCount {
		r := tilePixels(tile)
		lo, hi := byte(255), byte(0)
		for y := range TileSize {
			for x := range TileSize {
				at := ((int(r[1])+y)*AtlasSize + int(r[0]) + x) * 4
				if pixels[at+3] != 255 {
					t.Fatalf("tile %d has a transparent pixel", tile)
				}
				lo, hi = min(lo, pixels[at+1]), max(hi, pixels[at+1])
			}
		}
		if hi-lo <= 20 {
			t.Errorf("tile %d is nearly flat", tile)
		}
	}
}

func TestAtlasDeterministic(t *testing.T) {
	a, b := make([]byte, AtlasBytes), make([]byte, AtlasBytes)
	generateAtlas(a)
	generateAtlas(b)
	if !bytes.Equal(a, b) {
		t.Error("atlas differs between runs")
	}
}

func TestTileLayersMatchAtlasTiles(t *testing.T) {
	atlas := make([]byte, AtlasBytes)
	generateAtlas(atlas)
	layers := tileLayers(atlas)
	const layerBytes = TileSize * TileSize * 4
	if len(layers) != int(tileCount)*layerBytes {
		t.Fatalf("got %d bytes, want %d layers of %d", len(layers), tileCount, layerBytes)
	}
	for tile := range tileCount {
		ox, oy := tileOrigin(tile)
		layer := layers[int(tile)*layerBytes:][:layerBytes]
		for y := range TileSize {
			want := atlas[((oy+y)*AtlasSize+ox)*4:][:TileSize*4]
			if !bytes.Equal(layer[y*TileSize*4:][:TileSize*4], want) {
				t.Fatalf("tile %d row %d differs from the atlas", tile, y)
			}
		}
	}
}
