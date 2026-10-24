package goxel

import (
	"math"

	. "github.com/tacf/goxel/math"
	. "github.com/tacf/goxel/renderer"
)

const (
	halfWidth  = 0.3
	bodyHeight = 1.8
	eyeHeight  = 1.62

	gravity     = 28.0
	jumpSpeed   = 9.0
	walkSpeed   = 4.3
	sprintSpeed = 6.5
	flySpeed    = 12.0
	maxFall     = 50.0
	// contactSkin keeps the body a hair away from the block it rests against.
	contactSkin = 1e-4
	// maxSubstep is the longest distance moved before collisions are checked again.
	maxSubstep = 0.2
)

// Input is the movement requested for one frame.
type Input struct {
	Forward, Strafe float32
	Jump, Descend   bool
	Sprint          bool
}

// Player is a first person body with an axis aligned box for collisions.
type Player struct {
	// Pos is the position of the feet.
	Pos      Vec3
	Vel      Vec3
	Yaw      float32
	Pitch    float32
	OnGround bool
	Flying   bool
	arm      *ViewModel
}

func (p *Player) Eye() Vec3 {
	return Vec3{p.Pos[0], p.Pos[1] + eyeHeight, p.Pos[2]}
}

func (p *Player) Forward() Vec3 {
	cp := Cos(p.Pitch)
	return Vec3{Sin(p.Yaw) * cp, Sin(p.Pitch), -Cos(p.Yaw) * cp}
}

// Look turns the view and stops short of looking straight up or down.
func (p *Player) Look(dyaw, dpitch float32) {
	p.Yaw += dyaw
	limit := PI/2 - 0.01
	p.Pitch = Clamp(p.Pitch+dpitch, -limit, limit)
}

// Update applies input and physics for one frame.
func (p *Player) Update(w *World, in Input, dt float32) {
	speed := float32(walkSpeed)
	if p.Flying {
		speed = flySpeed
	} else if in.Sprint {
		speed = sprintSpeed
	}
	fx, fz := Sin(p.Yaw), -Cos(p.Yaw)
	rx, rz := Cos(p.Yaw), Sin(p.Yaw)
	wx := fx*in.Forward + rx*in.Strafe
	wz := fz*in.Forward + rz*in.Strafe
	if l := Sqrt(wx*wx + wz*wz); l > 1 {
		wx /= l
		wz /= l
	}
	p.Vel[0], p.Vel[2] = wx*speed, wz*speed

	if p.Flying {
		var climb float32
		if in.Jump {
			climb++
		}
		if in.Descend {
			climb--
		}
		p.Vel[1] = climb * speed
	} else {
		if in.Jump && p.OnGround {
			p.Vel[1] = jumpSpeed
		}
		p.Vel[1] = max(p.Vel[1]-gravity*dt, -maxFall)
	}

	// Small sub-steps keep fast falls from tunnelling through blocks.
	delta := Vec3{p.Vel[0] * dt, p.Vel[1] * dt, p.Vel[2] * dt}
	biggest := max(Abs(delta[0]), Abs(delta[1]), Abs(delta[2]))
	steps := max(1, int(math.Ceil(float64(biggest/maxSubstep))))
	inv := 1 / float32(steps)
	p.OnGround = false
	for range steps {
		for _, axis := range [...]int{0, 2, 1} {
			p.moveAxis(w, axis, delta[axis]*inv)
		}
	}
}

// moveAxis moves along one axis and, on a collision, snaps flush against the block entered.
func (p *Player) moveAxis(w *World, axis int, d float32) {
	if d == 0 {
		return
	}
	p.Pos[axis] += d
	if !p.collides(w) {
		return
	}
	vertical := axis == 1
	if d > 0 {
		top := float32(halfWidth)
		if vertical {
			top = bodyHeight
		}
		p.Pos[axis] = Floor(p.Pos[axis]+top) - top - contactSkin
	} else if vertical {
		p.Pos[axis] = Floor(p.Pos[axis]) + 1
		p.OnGround = true
	} else {
		p.Pos[axis] = Floor(p.Pos[axis]-halfWidth) + 1 + halfWidth + contactSkin
	}
	p.Vel[axis] = 0
}

func (p *Player) collides(w *World) bool {
	x0, x1 := int(Floor(p.Pos[0]-halfWidth)), int(Floor(p.Pos[0]+halfWidth))
	y0, y1 := int(Floor(p.Pos[1])), int(Floor(p.Pos[1]+bodyHeight))
	z0, z1 := int(Floor(p.Pos[2]-halfWidth)), int(Floor(p.Pos[2]+halfWidth))
	for y := y0; y <= y1; y++ {
		for z := z0; z <= z1; z++ {
			for x := x0; x <= x1; x++ {
				if w.SolidAt(x, y, z) {
					return true
				}
			}
		}
	}
	return false
}

// OverlapsCell reports whether the unit cube at the cell overlaps the player body.
func (p *Player) OverlapsCell(x, y, z int) bool {
	fx, fy, fz := float32(x), float32(y), float32(z)
	return p.Pos[0]+halfWidth > fx && p.Pos[0]-halfWidth < fx+1 &&
		p.Pos[1]+bodyHeight > fy && p.Pos[1] < fy+1 &&
		p.Pos[2]+halfWidth > fz && p.Pos[2]-halfWidth < fz+1
}
