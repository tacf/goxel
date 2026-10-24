// Command goxel is a small Minecraft style voxel game built on raylib.
package main

import (
	"runtime"

	goxel "github.com/tacf/goxel"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	windowWidth  = 1280
	windowHeight = 720
)

func init() {
	// GLFW must run on the main thread.
	runtime.LockOSThread()
}

func main() {
	rl.SetTraceLogLevel(rl.LogWarning)
	goxel.QuietWaylandWarnings()
	// High DPI makes raylib size its viewport from the real framebuffer on scaled displays.
	rl.SetConfigFlags(rl.FlagMsaa4xHint | rl.FlagVsyncHint | rl.FlagWindowResizable | rl.FlagWindowHighdpi)
	rl.InitWindow(windowWidth, windowHeight, "goxel")
	defer rl.CloseWindow()
	rl.DisableCursor()

	game := goxel.NewGame()
	defer game.Close()

	for !rl.WindowShouldClose() {
		game.Update(min(rl.GetFrameTime(), 0.05))
		game.Draw()
	}
}
