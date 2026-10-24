package main

import (
	"fmt"
	"os"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	windowWidth  = 1280
	windowHeight = 720
	// Frames to wait for the compositor to report a size before asking for one.
	syncFrames = 3
)

// Wayland compositors choose the window size after creation and clients may not
// position windows, so raylib reports a warning for the centring it attempts.
func quietWaylandWarnings() {
	rl.SetTraceLogCallback(func(level int, text string) {
		if strings.Contains(text, "window position") {
			return
		}
		fmt.Fprintln(os.Stderr, text)
	})
}

// windowSync makes sure the viewport follows the real window and framebuffer size.
type windowSync struct {
	frames int
	synced bool
}

// Update runs once per frame. A resize event from the compositor already
// refreshes raylib's viewport and scale. If none arrived shortly after start,
// requesting the size again triggers a fresh configure and framebuffer update.
func (s *windowSync) Update() {
	if s.synced {
		return
	}
	s.frames++
	if rl.IsWindowResized() {
		s.synced = true
		return
	}
	if s.frames >= syncFrames {
		rl.SetWindowSize(windowWidth, windowHeight)
		s.synced = true
	}
}

// clearDepth clears the depth buffer and leaves the picture untouched, so a
// following pass is not hidden by, or clipped into, what was drawn before.
func clearDepth() {
	rl.DrawRenderBatchActive()
	rl.ColorMask(false, false, false, false)
	rl.ClearScreenBuffers()
	rl.ColorMask(true, true, true, true)
}
