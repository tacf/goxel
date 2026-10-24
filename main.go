// Command zube is a small Minecraft style voxel game built on raylib.
package main

import (
	"runtime"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() {
	// GLFW must run on the main thread.
	runtime.LockOSThread()
}

func main() {
	rl.SetTraceLogLevel(rl.LogWarning)
	quietWaylandWarnings()
	// High DPI makes raylib size its viewport from the real framebuffer on scaled displays.
	rl.SetConfigFlags(rl.FlagMsaa4xHint | rl.FlagVsyncHint | rl.FlagWindowResizable | rl.FlagWindowHighdpi)
	rl.InitWindow(windowWidth, windowHeight, "zube")
	defer rl.CloseWindow()
	rl.DisableCursor()

	game := NewGame()
	defer game.Close()

	for !rl.WindowShouldClose() {
		game.Update(min(rl.GetFrameTime(), 0.05))
		game.Draw()
	}
}
