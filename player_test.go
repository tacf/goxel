package main

import (
	"math"
	"testing"
)

func TestPlayerFallsAndRestsOnGround(t *testing.T) {
	w := newTestWorld(t)
	h := w.SurfaceHeight(60, 60)
	p := Player{Pos: Vec3{60.5, float32(h + 10), 60.5}}
	for range 300 {
		p.Update(w, Input{}, 1.0/60.0)
	}
	if !p.OnGround {
		t.Error("player should be on the ground")
	}
	if math.Abs(float64(p.Pos[1]-float32(h+1))) > 0.001 {
		t.Errorf("player rests at %v, want %v", p.Pos[1], h+1)
	}
}

func TestPlayerCannotLeaveWorld(t *testing.T) {
	w := newTestWorld(t)
	h := w.SurfaceHeight(2, 60)
	p := Player{Pos: Vec3{2.5, float32(h + 1), 60.5}, Yaw: -pi / 2}
	for range 300 {
		p.Update(w, Input{Forward: 1}, 1.0/60.0)
	}
	if p.Pos[0] < halfWidth {
		t.Errorf("player left the world at x=%v", p.Pos[0])
	}
}
