package game

import (
	. "github.com/tacf/goxel/math"
	"testing"
)

// newTestGame builds a session around a test world.
func newTestGame(t *testing.T, slot int) *Session {
	t.Helper()
	return &Session{World: NewWorld(1), Selected: slot}
}

// flatSpot returns a column and the y of its surface block.
func flatSpot(g *Session) (x, y, z int) {
	x, z = 40, 40
	return x, g.World.SurfaceHeight(x, z), z
}

func TestPlaceBuildsOnTheFaceUnderTheCrosshair(t *testing.T) {
	g := newTestGame(t, 3) // stone
	x, y, z := flatSpot(g)
	g.Player.Pos = Vec3{10.5, float32(y + 1), 10.5}
	g.Target = Hit{Pos: [3]int{x, y, z}, Normal: [3]int{0, 1, 0}}
	g.place()
	if got := g.World.Get(x, y+1, z); got != Stone {
		t.Fatalf("block above the target is %v, want Stone", got)
	}
	if g.cooldown <= 0 || !g.Player.Swing.Active() {
		t.Error("placing should start the cooldown and the swing")
	}
}

func TestPlaceRefusesInvalidSpots(t *testing.T) {
	cases := []struct {
		name   string
		slot   int
		normal [3]int
		inside bool
	}{
		{"pickaxe in hand", 0, [3]int{0, 1, 0}, false},
		{"ray started inside a block", 3, [3]int{}, false},
		{"cell overlaps the player", 3, [3]int{0, 1, 0}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newTestGame(t, tc.slot)
			x, y, z := flatSpot(g)
			g.Player.Pos = Vec3{10.5, float32(y + 1), 10.5}
			if tc.inside {
				g.Player.Pos = Vec3{float32(x) + 0.5, float32(y + 1), float32(z) + 0.5}
			}
			g.Target = Hit{Pos: [3]int{x, y, z}, Normal: tc.normal}
			g.place()
			if got := g.World.Get(x, y+1, z); got != Air {
				t.Errorf("block was placed: %v", got)
			}
		})
	}
}

func TestMiningBreaksBlockAfterEnoughStrikes(t *testing.T) {
	g := newTestGame(t, 0) // pickaxe
	x, y, z := flatSpot(g)
	g.World.Set(x, y+1, z, Stone)
	g.Target = Hit{Pos: [3]int{x, y + 1, z}, Normal: [3]int{0, 1, 0}}

	g.mine(false)
	if g.Mining.Progress != 0 || !g.Mining.Active {
		t.Fatal("aiming alone must not deal damage")
	}
	// The pickaxe is 4 times faster on stone, so three strikes of 0.35s/(3s/4) break it.
	strikes := 0
	for g.World.Get(x, y+1, z) == Stone && strikes < 10 {
		g.mine(true)
		strikes++
	}
	if g.World.Get(x, y+1, z) != Air || strikes != 3 {
		t.Errorf("after %d strikes block is %v, want Air after 3", strikes, g.World.Get(x, y+1, z))
	}
	if g.Mining.Active {
		t.Error("mining state should reset once the block breaks")
	}
}

func TestAimingElsewhereResetsDamage(t *testing.T) {
	var m Mining
	m.aimAt([3]int{1, 2, 3})
	m.hit(0.4)
	m.aimAt([3]int{1, 2, 3})
	if m.Progress != 0.4 {
		t.Errorf("same block lost progress: %v", m.Progress)
	}
	m.aimAt([3]int{1, 2, 4})
	if m.Progress != 0 || m.Pos != [3]int{1, 2, 4} {
		t.Errorf("new block should start fresh, got %+v", m)
	}
}
