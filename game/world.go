package game

import (
	"math"

	. "github.com/tacf/goxel/math"
)

const (
	ChunkSize  = 16
	Height     = 64
	ChunksXZ   = 12
	SizeXZ     = ChunkSize * ChunksXZ
	ChunkCount = ChunksXZ * ChunksXZ
)

// Hit is the result of a ray hitting a block.
type Hit struct {
	Pos [3]int
	// Normal points from the hit block towards the empty cell the ray came from.
	Normal [3]int
}

// World holds the voxel grid.
type World struct {
	blocks []Block
	// revs counts the edits of every chunk, so consumers can tell when their copy is stale.
	revs [ChunkCount]uint32
}

func NewWorld(seed uint32) *World {
	w := &World{blocks: make([]Block, SizeXZ*SizeXZ*Height)}
	for i := range w.revs {
		w.revs[i] = 1
	}
	w.generate(seed)
	return w
}

func index(x, y, z int) int { return (y*SizeXZ+z)*SizeXZ + x }

func inside(x, y, z int) bool {
	return x >= 0 && x < SizeXZ && z >= 0 && z < SizeXZ && y >= 0 && y < Height
}

// Get reads a block. Outside the world it is air, except the floor which is stone.
func (w *World) Get(x, y, z int) Block {
	if y < 0 {
		return Stone
	}
	if !inside(x, y, z) {
		return Air
	}
	return w.blocks[index(x, y, z)]
}

// GetCell is Get for a position array.
func (w *World) GetCell(p [3]int) Block { return w.Get(p[0], p[1], p[2]) }

// SolidAt is the collision query. The horizontal world border is a solid wall.
func (w *World) SolidAt(x, y, z int) bool {
	if y >= Height {
		return false
	}
	if y >= 0 && (x < 0 || x >= SizeXZ || z < 0 || z >= SizeXZ) {
		return true
	}
	return w.Get(x, y, z) != Air
}

// Set places a block and marks the chunks whose meshes it changes.
func (w *World) Set(x, y, z int, b Block) {
	if !inside(x, y, z) {
		return
	}
	w.blocks[index(x, y, z)] = b
	cx, lx := x/ChunkSize, x%ChunkSize
	cz, lz := z/ChunkSize, z%ChunkSize
	w.markDirty(cx, cz)
	// Edits on a chunk border change the faces of the neighbour.
	if lx == 0 {
		w.markDirty(cx-1, cz)
	}
	if lx == ChunkSize-1 {
		w.markDirty(cx+1, cz)
	}
	if lz == 0 {
		w.markDirty(cx, cz-1)
	}
	if lz == ChunkSize-1 {
		w.markDirty(cx, cz+1)
	}
}

// SetCell is Set for a position array.
func (w *World) SetCell(p [3]int, b Block) { w.Set(p[0], p[1], p[2], b) }

// ChunkRev is the edit revision of chunk i (z*ChunksXZ+x). It is never zero.
func (w *World) ChunkRev(i int) uint32 { return w.revs[i] }

func (w *World) markDirty(cx, cz int) {
	if cx < 0 || cx >= ChunksXZ || cz < 0 || cz >= ChunksXZ {
		return
	}
	w.revs[cz*ChunksXZ+cx]++
}

// SurfaceHeight is the y of the highest solid block in a column.
func (w *World) SurfaceHeight(x, z int) int {
	for y := Height - 1; y >= 0; y-- {
		if w.Get(x, y, z) != Air {
			return y
		}
	}
	return 0
}

// Raycast walks the voxel grid and returns the first solid block along the ray.
func (w *World) Raycast(origin, dir Vec3, maxDist float32) (Hit, bool) {
	var cell, step [3]int
	var tMax, tDelta Vec3
	inf := float32(math.Inf(1))
	for a := range 3 {
		cell[a] = int(Floor(origin[a]))
		switch {
		case dir[a] > 0:
			step[a] = 1
			tDelta[a] = 1 / dir[a]
			tMax[a] = (float32(cell[a]+1) - origin[a]) * tDelta[a]
		case dir[a] < 0:
			step[a] = -1
			tDelta[a] = -1 / dir[a]
			tMax[a] = (origin[a] - float32(cell[a])) * tDelta[a]
		default:
			tDelta[a], tMax[a] = inf, inf
		}
	}
	var normal [3]int
	for {
		if w.GetCell(cell) != Air {
			return Hit{cell, normal}, true
		}
		axis := 0
		if tMax[1] < tMax[axis] {
			axis = 1
		}
		if tMax[2] < tMax[axis] {
			axis = 2
		}
		if tMax[axis] > maxDist {
			return Hit{}, false
		}
		cell[axis] += step[axis]
		tMax[axis] += tDelta[axis]
		normal = [3]int{}
		normal[axis] = -step[axis]
	}
}
