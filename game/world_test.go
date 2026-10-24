package game

import (
	"slices"
	"testing"
)

// newTestWorld builds the seed 1 world.
func newTestWorld(t *testing.T) *World {
	t.Helper()
	return NewWorld(1)
}

func TestTerrainDeterministicAndLayered(t *testing.T) {
	a, b := newTestWorld(t), newTestWorld(t)
	if !slices.Equal(a.blocks, b.blocks) {
		t.Fatal("same seed produced different terrain")
	}
	h := newTerrain(1).height(50, 50)
	if a.Get(50, h-6, 50) != Stone {
		t.Error("expected stone below the surface layers")
	}
	if a.Get(50, h-1, 50) == Air {
		t.Error("expected solid ground below the surface")
	}
	if a.Get(50, Height-1, 50) != Air {
		t.Error("expected air at the top of the world")
	}
}

func TestSetMarksNeighbouringChunksDirty(t *testing.T) {
	w := newTestWorld(t)
	before := w.revs
	w.Set(ChunkSize, 10, 5, Stone)
	if w.ChunkRev(1) == before[1] || w.ChunkRev(0) == before[0] {
		t.Error("edit on a chunk border should dirty both chunks")
	}
	if w.ChunkRev(2) != before[2] {
		t.Error("unrelated chunk was marked dirty")
	}
}

func TestRaycastHitsSurfaceFromAbove(t *testing.T) {
	w := newTestWorld(t)
	h := w.SurfaceHeight(40, 40)
	origin := [3]float32{40.5, float32(h + 5), 40.5}
	hit, ok := w.Raycast(origin, [3]float32{0, -1, 0}, 20)
	if !ok || hit.Pos != [3]int{40, h, 40} || hit.Normal != [3]int{0, 1, 0} {
		t.Fatalf("unexpected hit %+v ok=%v", hit, ok)
	}
	if _, ok := w.Raycast(origin, [3]float32{0, 1, 0}, 20); ok {
		t.Error("ray pointing at the sky should miss")
	}
}
