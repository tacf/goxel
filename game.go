// Package goxel wires the game logic to the window, the input and the renderer.
package goxel

import (
	"github.com/tacf/goxel/game"
	"github.com/tacf/goxel/renderer"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const mouseSensitivity = 0.0025

// Game is the running application: a game session, its renderer and the window.
type Game struct {
	session  *game.Session
	renderer *renderer.Renderer
	winSync  windowSync
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
func (g *Game) Update(dt float32) {
	g.winSync.Update()
	g.session.Update(readControls(), dt)
	g.renderer.Update(g.session, dt)
}

func (g *Game) Draw() { g.renderer.Draw(g.session) }

// readControls samples the keyboard and the mouse.
func readControls() game.Controls {
	mouse := rl.GetMouseDelta()
	c := game.Controls{
		Move:        readMovement(),
		Yaw:         mouse.X * mouseSensitivity,
		Pitch:       -mouse.Y * mouseSensitivity,
		ToggleFly:   rl.IsKeyPressed(rl.KeyF),
		ToggleDebug: rl.IsKeyPressed(rl.KeyF3),
		Slot:        -1,
		Mine:        rl.IsMouseButtonDown(rl.MouseButtonLeft),
		Place:       rl.IsMouseButtonDown(rl.MouseButtonRight),
	}
	for i := range game.Hotbar {
		if rl.IsKeyPressed(int32(rl.KeyOne + i)) {
			c.Slot = i
		}
	}
	if wheel := rl.GetMouseWheelMove(); wheel > 0 {
		c.Scroll = -1
	} else if wheel < 0 {
		c.Scroll = 1
	}
	return c
}

func readMovement() game.Input {
	var in game.Input
	if rl.IsKeyDown(rl.KeyW) {
		in.Forward++
	}
	if rl.IsKeyDown(rl.KeyS) {
		in.Forward--
	}
	if rl.IsKeyDown(rl.KeyD) {
		in.Strafe++
	}
	if rl.IsKeyDown(rl.KeyA) {
		in.Strafe--
	}
	in.Jump = rl.IsKeyDown(rl.KeySpace)
	in.Descend = rl.IsKeyDown(rl.KeyLeftShift)
	in.Sprint = rl.IsKeyDown(rl.KeyLeftControl)
	return in
}
