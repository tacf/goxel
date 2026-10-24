package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	viewDistance     = 8 * ChunkSize
	reach            = 6
	mouseSensitivity = 0.0025
	placeCooldown    = 0.2
)

// hotbar lists the selectable items, matching the number keys 1 to 9.
var hotbar = [...]Item{
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

var skyColor = color.RGBA{135, 206, 235, 255}

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

// Game owns everything that lives for the duration of a play session.
type Game struct {
	world    *World
	player   Player
	arm      *ViewModel
	cracks   *Cracks
	material rl.Material
	atlas    rl.Texture2D

	selected  int
	cooldown  float32
	showDebug bool
	mining    Mining
	winSync   windowSync

	// target is the block under the crosshair, valid when hasTarget is set.
	target    Hit
	hasTarget bool
}

// NewGame loads the world and the GPU resources. The window must already exist.
func NewGame() *Game {
	atlas := loadAtlas()
	// Set up on a local, cgo rejects a pointer into a heap struct that holds Go pointers.
	material := rl.LoadMaterialDefault()
	rl.SetMaterialTexture(&material, rl.MapDiffuse, atlas)

	g := &Game{
		world:     NewWorld(1337),
		arm:       NewViewModel(),
		cracks:    LoadCracks(),
		material:  material,
		atlas:     atlas,
		showDebug: true,
	}

	spawnX, spawnZ := SizeXZ/2, SizeXZ/2
	g.player.Pos = Vec3{
		float32(spawnX) + 0.5,
		float32(g.world.SurfaceHeight(spawnX, spawnZ) + 1),
		float32(spawnZ) + 0.5,
	}
	return g
}

// Close releases the GPU resources.
func (g *Game) Close() {
	g.world.Close()
	g.cracks.Unload()
	rl.UnloadTexture(g.atlas)
}

func (g *Game) holding() Item { return hotbar[g.selected] }

// Update reads input and advances the simulation by dt seconds.
func (g *Game) Update(dt float32) {
	g.winSync.Update()

	mouse := rl.GetMouseDelta()
	g.player.Look(mouse.X*mouseSensitivity, -mouse.Y*mouseSensitivity)
	g.handleKeys()
	g.player.Update(g.world, readMovement(), dt)
	g.interact(dt)

	walk := float32(0)
	if g.player.OnGround || g.player.Flying {
		walk = sqrt(g.player.Vel[0]*g.player.Vel[0] + g.player.Vel[2]*g.player.Vel[2])
	}
	g.arm.Update(dt, walk)

	g.world.UpdateMeshes(g.player.Pos[0], g.player.Pos[2], 0.006)
}

func readMovement() Input {
	var in Input
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

// handleKeys covers the toggles and the hotbar selection.
func (g *Game) handleKeys() {
	if rl.IsKeyPressed(rl.KeyF) {
		g.player.Flying = !g.player.Flying
	}
	if rl.IsKeyPressed(rl.KeyF3) {
		g.showDebug = !g.showDebug
	}
	for i := range hotbar {
		if rl.IsKeyPressed(int32(rl.KeyOne + i)) {
			g.selectSlot(i)
		}
	}
	if wheel := rl.GetMouseWheelMove(); wheel > 0 {
		g.selectSlot(wrap(g.selected-1, len(hotbar)))
	} else if wheel < 0 {
		g.selectSlot(wrap(g.selected+1, len(hotbar)))
	}
}

func (g *Game) selectSlot(i int) {
	if i != g.selected {
		g.selected = i
		g.arm.StartEquip()
	}
}

// interact handles aiming, mining and placing blocks.
func (g *Game) interact(dt float32) {
	g.target, g.hasTarget = g.world.Raycast(g.player.Eye(), g.player.Forward(), reach)
	left := rl.IsMouseButtonDown(rl.MouseButtonLeft)
	right := rl.IsMouseButtonDown(rl.MouseButtonRight)
	g.cooldown -= dt

	// Damage is only dealt when the swing lands, so cracks advance with each hit.
	struck := g.arm.TakeImpact()
	if left {
		g.arm.StartSwing()
	}
	if left && g.hasTarget {
		g.mine(struck)
	} else {
		g.mining = Mining{}
	}
	g.mining.ease(dt)

	if right && !left && g.hasTarget && g.cooldown <= 0 {
		g.place()
	}
}

// mine keeps aiming at the target and deals damage when the swing strikes.
func (g *Game) mine(struck bool) {
	g.mining.aimAt(g.target.Pos)
	if !struck {
		return
	}
	b := g.world.GetCell(g.target.Pos)
	damage := swingTime * g.holding().MineSpeed(b) / b.Hardness()
	if g.mining.hit(damage) {
		g.world.SetCell(g.target.Pos, Air)
		g.mining = Mining{}
	}
}

// place puts the held block on the face under the crosshair.
func (g *Game) place() {
	held := g.holding()
	if held.Kind != ItemBlock || g.target.Normal == [3]int{} {
		return
	}
	var pos [3]int
	for a := range pos {
		pos[a] = g.target.Pos[a] + g.target.Normal[a]
	}
	if g.world.GetCell(pos) != Air || g.player.OverlapsCell(pos[0], pos[1], pos[2]) {
		return
	}
	g.world.SetCell(pos, held.Block)
	g.cooldown = placeCooldown
	g.arm.StartSwing()
}

// Draw renders the world, the arm and the HUD.
func (g *Game) Draw() {
	eye, dir := g.player.Eye(), g.player.Forward()
	camera := rl.Camera3D{
		Position:   rl.NewVector3(eye[0], eye[1], eye[2]),
		Target:     rl.NewVector3(eye[0]+dir[0], eye[1]+dir[1], eye[2]+dir[2]),
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       70,
		Projection: rl.CameraPerspective,
	}

	rl.BeginDrawing()
	rl.ClearBackground(skyColor)
	rl.BeginMode3D(camera)
	g.world.Draw(g.material, g.player.Pos[0], g.player.Pos[2], viewDistance)
	if g.hasTarget {
		rl.DrawCubeWiresV(
			rl.NewVector3(float32(g.target.Pos[0])+0.5, float32(g.target.Pos[1])+0.5, float32(g.target.Pos[2])+0.5),
			rl.NewVector3(1.003, 1.003, 1.003),
			black,
		)
	}
	if g.mining.Active && g.mining.Shown > 0.03 {
		g.cracks.Draw(g.mining.Pos, g.mining.Shown)
	}
	rl.EndMode3D()
	g.arm.Draw(g.holding(), g.atlas)
	g.drawHUD()
	rl.EndDrawing()
}

// loadAtlas builds the procedural texture atlas and uploads it with mipmaps.
func loadAtlas() rl.Texture2D {
	pixels := make([]byte, AtlasBytes)
	generateAtlas(pixels)
	img := rl.NewImage(pixels, AtlasSize, AtlasSize, 1, rl.UncompressedR8g8b8a8)
	atlas := rl.LoadTextureFromImage(img)
	rl.GenTextureMipmaps(&atlas)
	// Crisp pixels up close, smooth mip blending at a distance.
	rl.TextureParameters(atlas.ID, rl.TextureMagFilter, rl.TextureFilterNearest)
	rl.TextureParameters(atlas.ID, rl.TextureMinFilter, rl.TextureFilterNearestMipLinear)
	return atlas
}
