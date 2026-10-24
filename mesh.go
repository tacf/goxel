package goxel

import (
	"math"

	. "github.com/tacf/goxel/math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// aoLevels is the brightness of a vertex by how many neighbours occlude it.
var aoLevels = [4]float32{0.5, 0.7, 0.85, 1.0}

// Chunk keeps the GPU handles of one mesh. The vertex arrays are scratch data.
type Chunk struct {
	mesh    rl.Mesh
	hasMesh bool
	dirty   bool
}

// meshData is the CPU side vertex data of the chunk being built.
type meshData struct {
	verts  []float32
	colors []uint8
	// uvs are the coordinates inside a tile, 0..1.
	uvs []float32
	// layers hold the tile array layer in x, as the second texture coordinate.
	layers []float32
}

func (m *meshData) reset() {
	m.verts, m.colors = m.verts[:0], m.colors[:0]
	m.uvs, m.layers = m.uvs[:0], m.layers[:0]
}

func (m *meshData) vertexCount() int { return len(m.verts) / 3 }

func (m *meshData) addVertex(pos Vec3, rgba [4]uint8, uv [2]float32, layer Tile) {
	m.verts = append(m.verts, pos[:]...)
	m.colors = append(m.colors, rgba[:]...)
	m.uvs = append(m.uvs, uv[:]...)
	m.layers = append(m.layers, float32(layer), 0)
}

// Close releases the GPU meshes.
func (w *World) Close() {
	for i := range w.chunks {
		if w.chunks[i].hasMesh {
			rl.UnloadMesh(&w.chunks[i].mesh)
		}
	}
}

// UpdateMeshes rebuilds dirty chunks, nearest to the position first, until the time budget is spent.
func (w *World) UpdateMeshes(nearX, nearZ float32, budgetSeconds float64) {
	start := rl.GetTime()
	for {
		best := -1
		bestDist := float32(math.MaxFloat32)
		for i := range w.chunks {
			if !w.chunks[i].dirty {
				continue
			}
			if d := chunkDistSq(i, nearX, nearZ); d < bestDist {
				bestDist, best = d, i
			}
		}
		if best < 0 {
			return
		}
		w.chunks[best].dirty = false
		w.meshChunk(best)
		w.uploadChunk(&w.chunks[best])
		if rl.GetTime()-start > budgetSeconds {
			return
		}
	}
}

func chunkDistSq(i int, x, z float32) float32 {
	cx := float32(i%ChunksXZ*ChunkSize + ChunkSize/2)
	cz := float32(i/ChunksXZ*ChunkSize + ChunkSize/2)
	return (cx-x)*(cx-x) + (cz-z)*(cz-z)
}

// Draw renders the chunks within the view distance of the camera.
func (w *World) Draw(material rl.Material, camX, camZ, viewDist float32) {
	identity := rl.MatrixIdentity()
	for i := range w.chunks {
		if !w.chunks[i].hasMesh || chunkDistSq(i, camX, camZ) > viewDist*viewDist {
			continue
		}
		rl.DrawMesh(w.chunks[i].mesh, material, identity)
	}
}

// meshChunk fills the scratch arrays with the visible faces of a chunk.
func (w *World) meshChunk(i int) {
	w.scratch.reset()
	ox, oz := i%ChunksXZ*ChunkSize, i/ChunksXZ*ChunkSize
	for y := range Height {
		for lz := range ChunkSize {
			for lx := range ChunkSize {
				if b := w.Get(ox+lx, y, oz+lz); b != Air {
					w.addBlockFaces(b, ox+lx, y, oz+lz)
				}
			}
		}
	}
}

// uploadChunk replaces the GPU mesh of a chunk with the scratch arrays.
func (w *World) uploadChunk(chunk *Chunk) {
	if chunk.hasMesh {
		rl.UnloadMesh(&chunk.mesh)
		chunk.mesh, chunk.hasMesh = rl.Mesh{}, false
	}
	count := w.scratch.vertexCount()
	if count == 0 {
		return
	}

	mesh := rl.Mesh{
		VertexCount:   int32(count),
		TriangleCount: int32(count / 3),
		Vertices:      &w.scratch.verts[0],
		Colors:        &w.scratch.colors[0],
		Texcoords:     &w.scratch.uvs[0],
		Texcoords2:    &w.scratch.layers[0],
	}
	rl.UploadMesh(&mesh, false)
	// The scratch arrays are reused, so drop the pointers and keep the GPU handles only.
	chunk.mesh = rl.Mesh{
		VertexCount:   mesh.VertexCount,
		TriangleCount: mesh.TriangleCount,
		VaoID:         mesh.VaoID,
		VboID:         mesh.VboID,
	}
	chunk.hasMesh = true
}

func (w *World) addBlockFaces(b Block, x, y, z int) {
	jitter := 0.94 + 0.06*Rnd(x+y*7919, z, 77)
	for fi := range VoxelFaces {
		face := &VoxelFaces[fi]
		nx, ny, nz := x+face.N[0], y+face.N[1], z+face.N[2]
		if w.Get(nx, ny, nz) != Air {
			continue
		}

		axis := 2
		if face.N[0] != 0 {
			axis = 0
		} else if face.N[1] != 0 {
			axis = 1
		}
		t1, t2 := (axis+1)%3, (axis+2)%3
		base := [3]int{nx, ny, nz}
		tile := TileOf(b, fi)

		var pos [4]Vec3
		var ao [4]int
		var rgba [4][4]uint8
		var uv [4][2]float32
		for k, v := range face.V {
			pos[k] = Vec3{float32(x + v[0]), float32(y + v[1]), float32(z + v[2])}
			// Occlusion comes from the two neighbours beside the vertex and the one diagonal.
			s1, s2 := v[t1]*2-1, v[t2]*2-1
			p1, p2 := base, base
			p1[t1] += s1
			p2[t2] += s2
			pc := p1
			pc[t2] += s2
			ao[k] = vertexAO(
				w.GetCell(p1) != Air,
				w.GetCell(p2) != Air,
				w.GetCell(pc) != Air,
			)

			// Colour is only light, the texture supplies the hue.
			light := uint8(Clamp(face.Shade*aoLevels[ao[k]]*jitter*255, 0, 255))
			rgba[k] = [4]uint8{light, light, light, 255}
			uv[k] = faceUV(fi, v)
		}

		// Split along the diagonal that hides the least interpolation artefacts.
		order := [6]int{0, 1, 2, 0, 2, 3}
		if ao[0]+ao[2] > ao[1]+ao[3] {
			order = [6]int{1, 2, 3, 1, 3, 0}
		}
		for _, k := range order {
			w.scratch.addVertex(pos[k], rgba[k], uv[k], tile)
		}
	}
}

// vertexAO is the occlusion level 0..3 of a vertex, 3 being fully lit.
func vertexAO(side1, side2, corner bool) int {
	if side1 && side2 {
		return 0
	}
	level := 3
	for _, occluded := range [...]bool{side1, side2, corner} {
		if occluded {
			level--
		}
	}
	return level
}

// faceUV maps a face corner onto its tile, 0..1 in both directions.
func faceUV(face int, v [3]int) [2]float32 {
	a := float32(v[0])
	if face == 2 || face == 3 {
		a = float32(v[2])
	}
	b := 1 - float32(v[1])
	if face < 2 {
		b = float32(v[2])
	}
	return [2]float32{a, b}
}
