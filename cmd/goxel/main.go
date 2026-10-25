// Command goxel is a small Minecraft style voxel game built on raylib.
package main

import (
	goxel "github.com/tacf/goxel"
	"github.com/tacf/goxel/platform"
)

const (
	windowWidth  = 1280
	windowHeight = 720
)

func main() {
	runtime, window := platform.Init(platform.Options{
		Title: "goxel", Width: windowWidth, Height: windowHeight,
		Flags: platform.WindowMSAA4x | platform.WindowVSync | platform.WindowResizable | platform.WindowHighDPI,
	})
	defer runtime.Close()

	game := goxel.NewGame()
	defer game.Close()

	for !window.ShouldClose() {
		window.SyncSize()
		game.Update(window.PollInput(), window.FrameTime(0.05))
		game.Draw()
	}
}
