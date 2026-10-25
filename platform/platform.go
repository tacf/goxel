// Package platform owns the window, frame timing and input device access.
package platform

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	// WindowResizable lets the window manager resize the game window.
	WindowResizable = 1 << iota
	// WindowHighDPI requests a high-density framebuffer where available.
	WindowHighDPI
	// WindowVSync synchronizes presentation to the display refresh rate.
	WindowVSync
	// WindowMSAA4x requests four-sample multisampling.
	WindowMSAA4x
)

// Options configures a platform window.
type Options struct {
	Title         string
	Width, Height int
	Flags         uint
}

// Runtime owns raylib initialization and the locked main thread.
type Runtime struct{ closed bool }

// Init initializes the native runtime and creates its window on the calling OS thread.
func Init(opts Options) (*Runtime, *Window) {
	runtime.LockOSThread()
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.SetTraceLogCallback(func(_ int, text string) {
		if strings.Contains(text, "window position") {
			return
		}
		fmt.Fprintln(os.Stderr, text)
	})
	var flags uint32
	if opts.Flags&WindowResizable != 0 {
		flags |= rl.FlagWindowResizable
	}
	if opts.Flags&WindowHighDPI != 0 {
		flags |= rl.FlagWindowHighdpi
	}
	if opts.Flags&WindowVSync != 0 {
		flags |= rl.FlagVsyncHint
	}
	if opts.Flags&WindowMSAA4x != 0 {
		flags |= rl.FlagMsaa4xHint
	}
	rl.SetConfigFlags(flags)
	rl.InitWindow(int32(opts.Width), int32(opts.Height), opts.Title)
	rl.DisableCursor()
	return &Runtime{}, &Window{width: opts.Width, height: opts.Height, syncFrames: 3}
}

// Close releases the window and returns the calling goroutine to the scheduler.
func (r *Runtime) Close() {
	if r == nil || r.closed {
		return
	}
	rl.CloseWindow()
	r.closed = true
	runtime.UnlockOSThread()
}

// Window is the platform-neutral window handle used by the application loop.
type Window struct {
	width, height int
	syncFrames    int
	synced        bool
}

// ShouldClose reports whether the user requested the window to close.
func (w *Window) ShouldClose() bool { return rl.WindowShouldClose() }

// SyncSize refreshes the window size if the compositor did not report it at startup.
func (w *Window) SyncSize() {
	if w == nil || w.synced {
		return
	}
	w.syncFrames--
	if rl.IsWindowResized() {
		w.synced = true
		return
	}
	if w.syncFrames <= 0 {
		rl.SetWindowSize(int(w.width), int(w.height))
		w.synced = true
	}
}

// FrameTime returns the elapsed seconds since the last frame, capped at max.
func (w *Window) FrameTime(max float32) float32 { return min(rl.GetFrameTime(), max) }

// Key names physical or logical game actions without exposing raylib key codes.
type Key uint8

const (
	KeyW Key = iota
	KeyA
	KeyS
	KeyD
	KeySpace
	KeyLeftShift
	KeyLeftControl
	KeyF
	KeyF3
	KeyOne
	KeyTwo
	KeyThree
	KeyFour
	KeyFive
	KeySix
	KeySeven
	KeyEight
	KeyNine
	KeyLeft
	KeyRight
	KeyEscape
	keyCount
)

var raylibKeys = [keyCount]int32{
	rl.KeyW, rl.KeyA, rl.KeyS, rl.KeyD, rl.KeySpace, rl.KeyLeftShift,
	rl.KeyLeftControl, rl.KeyF, rl.KeyF3, rl.KeyOne, rl.KeyTwo, rl.KeyThree,
	rl.KeyFour, rl.KeyFive, rl.KeySix, rl.KeySeven, rl.KeyEight, rl.KeyNine,
	rl.KeyLeft, rl.KeyRight, rl.KeyEscape,
}

// MouseButton identifies a portable mouse button.
type MouseButton uint8

const (
	MouseLeft MouseButton = iota
	MouseRight
	mouseButtonCount
)

// Input is a snapshot of input state for one frame.
type Input struct {
	down, pressed, repeated [keyCount]bool
	mouseDown               [mouseButtonCount]bool
	MouseDX, MouseDY        float32
	WheelY                  float32
}

// PollInput samples keyboard and mouse state once for the current frame.
func (w *Window) PollInput() Input {
	var in Input
	for key, code := range raylibKeys {
		in.down[key] = rl.IsKeyDown(code)
		in.pressed[key] = rl.IsKeyPressed(code)
		in.repeated[key] = rl.IsKeyPressedRepeat(code)
	}
	in.mouseDown[MouseLeft] = rl.IsMouseButtonDown(rl.MouseButtonLeft)
	in.mouseDown[MouseRight] = rl.IsMouseButtonDown(rl.MouseButtonRight)
	delta := rl.GetMouseDelta()
	in.MouseDX, in.MouseDY = delta.X, delta.Y
	in.WheelY = rl.GetMouseWheelMove()
	return in
}

// Down reports whether a key is held in this frame.
func (i Input) Down(key Key) bool { return key < keyCount && i.down[key] }

// Pressed reports whether a key was first pressed in this frame.
func (i Input) Pressed(key Key) bool { return key < keyCount && i.pressed[key] }

// Repeated reports a new press or an OS key-repeat event.
func (i Input) Repeated(key Key) bool { return key < keyCount && (i.pressed[key] || i.repeated[key]) }

// MouseDown reports whether a mouse button is held.
func (i Input) MouseDown(button MouseButton) bool {
	return button < mouseButtonCount && i.mouseDown[button]
}

// GLProcAddress resolves a graphics procedure using the active rendering backend.
func GLProcAddress(name string) unsafe.Pointer { return rl.GetProcAddress(name) }
