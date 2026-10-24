// Package renderer draws the game state with raylib. It reads the game and never changes it.
package renderer

import (
	"image/color"
	"log"

	"github.com/tacf/goxel/game"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	viewDistance = 8 * game.ChunkSize
	// meshBudget is the time per frame spent rebuilding chunk meshes, in seconds.
	meshBudget = 0.006
)

var skyColor = color.RGBA{135, 206, 235, 255}

// Renderer owns the GPU resources and draws a session.
type Renderer struct {
	world    WorldMesh
	cracks   *Cracks
	view     *ViewModel
	material rl.Material
	shader   rl.Shader
	// tiles feeds the block shader, atlas serves the 2D drawing of the hotbar and held block.
	tiles TileArray
	atlas rl.Texture2D
}

// New loads the GPU resources. The window must already exist.
func New() *Renderer {
	pixels := make([]byte, AtlasBytes)
	generateAtlas(pixels)
	tiles, err := NewTileArray(tileLayers(pixels), TileSize, int(tileCount))
	if err != nil {
		log.Fatal(err)
	}
	shader := rl.LoadShaderFromMemory(BlockVertexShader, BlockFragmentShader)
	// Set up on a local, cgo rejects a pointer into a heap struct that holds Go pointers.
	material := rl.LoadMaterialDefault()
	material.Shader = shader

	return &Renderer{
		cracks:   LoadCracks(),
		view:     NewViewModel(),
		material: material,
		shader:   shader,
		tiles:    tiles,
		atlas:    loadAtlas(pixels),
	}
}

// Close releases the GPU resources.
func (r *Renderer) Close() {
	r.world.Close()
	r.cracks.Unload()
	rl.UnloadTexture(r.atlas)
	r.tiles.Unload()
	rl.UnloadShader(r.shader)
}

// Update rebuilds stale meshes and advances the animations by dt seconds.
func (r *Renderer) Update(s *game.Session, dt float32) {
	r.view.Update(dt, s.Player.WalkSpeed(), s.Player.Swing.Progress(), s.Selected)
	r.world.Update(s.World, s.Player.Pos[0], s.Player.Pos[2], meshBudget)
}

// Draw renders the world, the arm and the HUD.
func (r *Renderer) Draw(s *game.Session) {
	eye, dir := s.Player.Eye(), s.Player.Forward()
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
	r.tiles.Bind()
	r.world.Draw(r.material, s.Player.Pos[0], s.Player.Pos[2], viewDistance)
	if s.HasTarget {
		rl.DrawCubeWiresV(
			rl.NewVector3(float32(s.Target.Pos[0])+0.5, float32(s.Target.Pos[1])+0.5, float32(s.Target.Pos[2])+0.5),
			rl.NewVector3(1.003, 1.003, 1.003),
			black,
		)
	}
	if s.Mining.Active && s.Mining.Shown > 0.03 {
		r.cracks.Draw(s.Mining.Pos, s.Mining.Shown)
	}
	rl.EndMode3D()
	r.view.Draw(s.Holding(), r.atlas)
	r.drawHUD(s)
	rl.EndDrawing()
}

// loadAtlas uploads the atlas for 2D drawing. It is sampled without filtering or mipmaps.
func loadAtlas(pixels []byte) rl.Texture2D {
	img := rl.NewImage(pixels, AtlasSize, AtlasSize, 1, rl.UncompressedR8g8b8a8)
	atlas := rl.LoadTextureFromImage(img)
	rl.SetTextureFilter(atlas, rl.FilterPoint)
	return atlas
}
