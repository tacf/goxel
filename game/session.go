package game

import (
	. "github.com/tacf/goxel/math"
)

const (
	reach         = 6
	placeCooldown = 0.2
)

// Hotbar lists the selectable items, matching the number keys 1 to 9.
var Hotbar = [...]Item{
	{Kind: ItemPickaxe},
	{Kind: ItemBlock, Block: Grass},
	{Kind: ItemBlock, Block: Dirt},
	{Kind: ItemBlock, Block: Stone},
	{Kind: ItemBlock, Block: Cobble},
	{Kind: ItemBlock, Block: Planks},
	{Kind: ItemBlock, Block: Wood},
	{Kind: ItemBlock, Block: Sand},
	{Kind: ItemBlock, Block: Leaves},
}

// Mining tracks the damage dealt to the block under the crosshair.
type Mining struct {
	Active bool
	Pos    [3]int
	// Progress is the damage dealt so far, 0 to 1.
	Progress float32
	// Shown is an eased copy of Progress used for the crack overlay.
	Shown float32
}

// aimAt starts over when the crosshair moves to another block.
func (m *Mining) aimAt(pos [3]int) {
	if !m.Active || m.Pos != pos {
		*m = Mining{Active: true, Pos: pos}
	}
}

// hit adds damage and reports whether the block is destroyed.
func (m *Mining) hit(damage float32) bool {
	m.Progress += damage
	return m.Progress >= 1
}

// ease moves the shown progress towards the real one so cracks pop in quickly.
func (m *Mining) ease(dt float32) {
	m.Shown += (m.Progress - m.Shown) * min(1, dt*25)
}

// Controls is the player intent for one frame, already read from the input devices.
type Controls struct {
	Move Input
	// Yaw and Pitch are the view turn in radians.
	Yaw, Pitch  float32
	ToggleFly   bool
	ToggleDebug bool
	// Slot selects a hotbar slot directly, negative for none.
	Slot int
	// Scroll moves the selection by one slot per unit, in either direction.
	Scroll      int
	Mine, Place bool
}

// Session owns everything that lives for the duration of a play session.
type Session struct {
	World  *World
	Player Player
	Mining Mining

	Selected  int
	ShowDebug bool
	cooldown  float32

	// Target is the block under the crosshair, valid when HasTarget is set.
	Target    Hit
	HasTarget bool
}

// NewSession generates the world and puts the player on its surface.
func NewSession(seed uint32) *Session {
	s := &Session{
		World:     NewWorld(seed),
		ShowDebug: true,
	}
	spawnX, spawnZ := SizeXZ/2, SizeXZ/2
	s.Player.Pos = Vec3{
		float32(spawnX) + 0.5,
		float32(s.World.SurfaceHeight(spawnX, spawnZ) + 1),
		float32(spawnZ) + 0.5,
	}
	return s
}

func (s *Session) Holding() Item { return Hotbar[s.Selected] }

// Update applies the controls and advances the simulation by dt seconds.
func (s *Session) Update(c Controls, dt float32) {
	s.Player.Look(c.Yaw, c.Pitch)
	s.applyToggles(c)
	s.Player.Update(s.World, c.Move, dt)
	s.interact(c, dt)
	s.Player.Swing.Update(dt)
}

// applyToggles covers the toggles and the hotbar selection.
func (s *Session) applyToggles(c Controls) {
	if c.ToggleFly {
		s.Player.Flying = !s.Player.Flying
	}
	if c.ToggleDebug {
		s.ShowDebug = !s.ShowDebug
	}
	if c.Slot >= 0 && c.Slot < len(Hotbar) {
		s.Selected = c.Slot
	}
	if c.Scroll != 0 {
		s.Selected = Wrap(s.Selected+c.Scroll, len(Hotbar))
	}
}

// interact handles aiming, mining and placing blocks.
func (s *Session) interact(c Controls, dt float32) {
	s.Target, s.HasTarget = s.World.Raycast(s.Player.Eye(), s.Player.Forward(), reach)
	s.cooldown -= dt

	// Damage is only dealt when the swing lands, so cracks advance with each hit.
	struck := s.Player.Swing.TakeImpact()
	if c.Mine {
		s.Player.Swing.Start()
	}
	if c.Mine && s.HasTarget {
		s.mine(struck)
	} else {
		s.Mining = Mining{}
	}
	s.Mining.ease(dt)

	if c.Place && !c.Mine && s.HasTarget && s.cooldown <= 0 {
		s.place()
	}
}

// mine keeps aiming at the target and deals damage when the swing strikes.
func (s *Session) mine(struck bool) {
	s.Mining.aimAt(s.Target.Pos)
	if !struck {
		return
	}
	b := s.World.GetCell(s.Target.Pos)
	damage := SwingTime * s.Holding().MineSpeed(b) / b.Hardness()
	if s.Mining.hit(damage) {
		s.World.SetCell(s.Target.Pos, Air)
		s.Mining = Mining{}
	}
}

// place puts the held block on the face under the crosshair.
func (s *Session) place() {
	held := s.Holding()
	if held.Kind != ItemBlock || s.Target.Normal == [3]int{} {
		return
	}
	var pos [3]int
	for a := range pos {
		pos[a] = s.Target.Pos[a] + s.Target.Normal[a]
	}
	if s.World.GetCell(pos) != Air || s.Player.OverlapsCell(pos[0], pos[1], pos[2]) {
		return
	}
	s.World.SetCell(pos, held.Block)
	s.cooldown = placeCooldown
	s.Player.Swing.Start()
}
