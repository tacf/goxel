// Command texview opens a window to browse the generated block textures.
package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/tacf/goxel/game"
	"github.com/tacf/goxel/renderer"

	goxel "github.com/tacf/goxel"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	windowWidth  = 1040
	windowHeight = 600
	bigSize      = 320
	tiledCell    = 96
	thumbSize    = 56
	margin       = 40
)

// faceNames follow the order of the voxel face table.
var faceNames = [6]string{"top", "bottom", "+x", "-x", "+z", "-z"}

func init() {
	// GLFW must run on the main thread.
	runtime.LockOSThread()
}

func main() {
	rl.SetTraceLogLevel(rl.LogWarning)
	goxel.QuietWaylandWarnings()
	rl.SetConfigFlags(rl.FlagVsyncHint | rl.FlagWindowHighdpi)
	rl.InitWindow(windowWidth, windowHeight, "goxel - Texture Viewer")
	defer rl.CloseWindow()

	pixels := make([]byte, renderer.AtlasBytes)
	renderer.GenerateAtlas(pixels)
	atlas := renderer.LoadAtlas(pixels)
	defer rl.UnloadTexture(atlas)

	uses := tileUses()
	current := renderer.Tile(0)
	for !rl.WindowShouldClose() {
		current = step(current)
		draw(atlas, current, uses[current])
	}
}

// step moves the selection with the arrow keys or A and D.
func step(t renderer.Tile) renderer.Tile {
	n := int(renderer.TileCount)
	next := func(key int32) bool { return rl.IsKeyPressed(key) || rl.IsKeyPressedRepeat(key) }
	switch {
	case next(rl.KeyRight):
		return renderer.Tile((int(t) + 1) % n)
	case next(rl.KeyLeft):
		return renderer.Tile((int(t) + n - 1) % n)
	}
	return t
}

// tileUses lists, for every tile, the blocks that show it and on which faces.
func tileUses() [renderer.TileCount][]string {
	var uses [renderer.TileCount][]string
	for b := game.Block(1); b < game.BlockCount; b++ {
		var faces [renderer.TileCount][]int
		for f := range faceNames {
			t := renderer.TileOf(b, f)
			faces[t] = append(faces[t], f)
		}
		for t, fs := range faces {
			if len(fs) > 0 {
				uses[t] = append(uses[t], fmt.Sprintf("%s: %s", b, faceLabel(fs)))
			}
		}
	}
	return uses
}

// faceLabel names a set of faces, using sides for the four vertical ones.
func faceLabel(faces []int) string {
	var mask int
	for _, f := range faces {
		mask |= 1 << f
	}
	switch mask {
	case 0b111111:
		return "all faces"
	case 0b111100:
		return "sides"
	}
	names := make([]string, len(faces))
	for i, f := range faces {
		names[i] = faceNames[f]
	}
	return strings.Join(names, ", ")
}

func draw(atlas rl.Texture2D, current renderer.Tile, uses []string) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.NewColor(32, 34, 40, 255))

	title := fmt.Sprintf("%s", current)
	rl.DrawText(title, margin, 24, 28, rl.White)

	top := int32(80)
	drawTile(atlas, current, margin, top, bigSize)
	rl.DrawText("1x", margin, top+bigSize+8, 16, rl.Gray)

	tx := int32(margin + bigSize + margin)
	for y := range 3 {
		for x := range 3 {
			drawTile(atlas, current, tx+int32(x)*tiledCell, top+int32(y)*tiledCell, tiledCell)
		}
	}
	rl.DrawText("3x3 tiling", tx, top+3*tiledCell+8, 16, rl.Gray)

	ux := tx + 3*tiledCell + margin
	rl.DrawText("Used by", ux, top, 22, rl.White)
	if len(uses) == 0 {
		rl.DrawText("no block", ux, top+34, 20, rl.Gray)
	}
	for i, line := range uses {
		rl.DrawText(line, ux, top+34+int32(i)*26, 20, rl.LightGray)
	}

	drawStrip(atlas, current)
	rl.DrawText("Left and Right | Esc to quit", margin, windowHeight-28, 16, rl.Gray)
	rl.EndDrawing()
}

// drawStrip shows every tile in a row and marks the current one.
func drawStrip(atlas rl.Texture2D, current renderer.Tile) {
	y := int32(windowHeight - 28 - 24 - thumbSize)
	for t := range renderer.TileCount {
		x := int32(margin) + int32(t)*(thumbSize+8)
		drawTile(atlas, t, x, y, thumbSize)
		if t == current {
			rl.DrawRectangleLinesEx(rl.NewRectangle(float32(x-3), float32(y-3), thumbSize+6, thumbSize+6), 3, rl.White)
		}
	}
}

// drawTile draws one atlas tile scaled to size pixels, without filtering.
func drawTile(atlas rl.Texture2D, t renderer.Tile, x, y, size int32) {
	src := renderer.TilePixels(t)
	rl.DrawTexturePro(
		atlas,
		rl.NewRectangle(src[0], src[1], src[2], src[3]),
		rl.NewRectangle(float32(x), float32(y), float32(size), float32(size)),
		rl.NewVector2(0, 0), 0, rl.White,
	)
}
