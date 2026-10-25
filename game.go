// Package goxel wires the game logic to the window, the input and the renderer.
package goxel

import (
	"github.com/tacf/goxel/game"
	"github.com/tacf/goxel/platform"
	"github.com/tacf/goxel/renderer"
)

const mouseSensitivity = 0.0025

// Game is the running application: a game session, its renderer and the window.
type Game struct {
	session  *game.Session
	renderer *renderer.Renderer
}

// NewGame loads the world and the GPU resources. The window must already exist.
func NewGame() *Game {
	return &Game{
		session:  game.NewSession(1337),
		renderer: renderer.New(),
	}
}

// Close releases the GPU resources.
func (g *Game) Close() { g.renderer.Close() }

// Update reads input and advances the simulation by dt seconds.
func (g *Game) Update(input platform.Input, dt float32) {
	g.session.Update(readControls(input), dt)
	g.renderer.Update(g.session, dt)
}

func (g *Game) Draw() { g.renderer.Draw(g.session) }

// readControls samples the keyboard and the mouse.
func readControls(input platform.Input) game.Controls {
	c := game.Controls{
		Move:        readMovement(input),
		Yaw:         input.MouseDX * mouseSensitivity,
		Pitch:       -input.MouseDY * mouseSensitivity,
		ToggleFly:   input.Pressed(platform.KeyF),
		ToggleDebug: input.Pressed(platform.KeyF3),
		Slot:        -1,
		Mine:        input.MouseDown(platform.MouseLeft),
		Place:       input.MouseDown(platform.MouseRight),
	}
	for i := range game.Hotbar {
		if input.Pressed(platform.KeyOne + platform.Key(i)) {
			c.Slot = i
		}
	}
	if input.WheelY > 0 {
		c.Scroll = -1
	} else if input.WheelY < 0 {
		c.Scroll = 1
	}
	return c
}

func readMovement(input platform.Input) game.Input {
	var in game.Input
	if input.Down(platform.KeyW) {
		in.Forward++
	}
	if input.Down(platform.KeyS) {
		in.Forward--
	}
	if input.Down(platform.KeyD) {
		in.Strafe++
	}
	if input.Down(platform.KeyA) {
		in.Strafe--
	}
	in.Jump = input.Down(platform.KeySpace)
	in.Descend = input.Down(platform.KeyLeftShift)
	in.Sprint = input.Down(platform.KeyLeftControl)
	return in
}
