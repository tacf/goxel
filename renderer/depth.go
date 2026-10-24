package renderer

import rl "github.com/gen2brain/raylib-go/raylib"

// ClearDepth clears the depth buffer and leaves the picture untouched, so a
// following pass is not hidden by, or clipped into, what was drawn before.
func ClearDepth() {
	rl.DrawRenderBatchActive()
	rl.ColorMask(false, false, false, false)
	rl.ClearScreenBuffers()
	rl.ColorMask(true, true, true, true)
}
