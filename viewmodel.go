package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	sleeveColor = rl.DarkBlue
	skinColor   = rl.Beige
	handleColor = rl.Brown
	ironColor   = rl.DarkGray
)

const (
	swingTime = 0.35
	equipTime = 0.25
	// Swing progress at which the tool reaches the target.
	impactAt = 0.5
)

// Pose places the held item in view space.
type Pose struct {
	Pos        Vec3
	RX, RY, RZ float32
}

func (p Pose) apply(v Vec3) Vec3 {
	r := rotate(v, p.RX, p.RY, p.RZ)
	return Vec3{r[0] + p.Pos[0], r[1] + p.Pos[1], r[2] + p.Pos[2]}
}

// rotate applies yaw, then pitch, then roll.
func rotate(p Vec3, rx, ry, rz float32) Vec3 {
	y1 := Vec3{p[0]*cos(ry) + p[2]*sin(ry), p[1], -p[0]*sin(ry) + p[2]*cos(ry)}
	x1 := Vec3{y1[0], y1[1]*cos(rx) - y1[2]*sin(rx), y1[1]*sin(rx) + y1[2]*cos(rx)}
	return Vec3{x1[0]*cos(rz) - x1[1]*sin(rz), x1[0]*sin(rz) + x1[1]*cos(rz), x1[2]}
}

// faceColors is the flat colour of the top, sides and bottom of a box.
type faceColors struct{ top, side, bottom color.RGBA }

func uniform(c color.RGBA) faceColors { return faceColors{c, c, c} }

// ViewModel animates the first person arm and the item in its hand.
type ViewModel struct {
	// swing is the progress of the current swing in 0..1, negative when idle.
	swing float32
	// equip is the progress of the raise animation after switching items.
	equip float32
	// impact is set for one update when the swing reaches its strike point.
	impact   bool
	bobPhase float32
	bobAmp   float32
}

func NewViewModel() *ViewModel { return &ViewModel{swing: -1, equip: 1} }

func (v *ViewModel) StartSwing() {
	if v.swing < 0 {
		v.swing = 0
	}
}

func (v *ViewModel) StartEquip() { v.equip = 0 }

// TakeImpact is true once per strike, then clears.
func (v *ViewModel) TakeImpact() bool {
	struck := v.impact
	v.impact = false
	return struck
}

func (v *ViewModel) Update(dt, walkSpeed float32) {
	if v.swing >= 0 {
		before := v.swing
		v.swing += dt / swingTime
		if before < impactAt && v.swing >= impactAt {
			v.impact = true
		}
		if v.swing >= 1 {
			v.swing = -1
		}
	}
	v.equip = min(1, v.equip+dt/equipTime)
	target := float32(0)
	if walkSpeed > 0.1 {
		target = 1
	}
	v.bobAmp += (target - v.bobAmp) * min(1, dt*10)
	v.bobPhase += dt * walkSpeed * 2.4
}

// pose is where the arm sits in view space for the current animation state.
func (v *ViewModel) pose() Pose {
	s := max(v.swing, 0)
	sw := sin(s * pi)
	sq := sin(sqrt(s) * pi)
	lift := (1 - v.equip) * (1 - v.equip)

	// At the strike point the tool has crossed towards the crosshair.
	return Pose{
		Pos: Vec3{
			0.36 + sin(v.bobPhase)*0.012*v.bobAmp - 0.25*sq,
			-0.36 - abs(cos(v.bobPhase))*0.016*v.bobAmp + 0.2*sw - lift*0.7,
			-0.28 - 0.15*sw,
		},
		RX: rad(8) - sw*rad(30),
		RY: rad(14) + sq*rad(20),
		RZ: -sin(s*s*pi) * rad(14),
	}
}

// Draw renders the arm over the world, using real depth testing against itself only.
func (v *ViewModel) Draw(item Item, atlas rl.Texture2D) {
	pose := v.pose()
	cam := rl.Camera3D{
		Position:   rl.NewVector3(0, 0, 0),
		Target:     rl.NewVector3(0, 0, -1),
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       70,
		Projection: rl.CameraPerspective,
	}
	clearDepth()
	rl.BeginMode3D(cam)

	drawBox(pose, Vec3{0, 0, -0.18}, Vec3{0.09, 0.09, 0.18}, 0, uniform(sleeveColor))
	drawBox(pose, Vec3{0, 0, -0.52}, Vec3{0.085, 0.085, 0.16}, 0, uniform(skinColor))
	if item.Kind == ItemPickaxe {
		drawBox(pose, Vec3{0, 0.13, -0.66}, Vec3{0.018, 0.2, 0.018}, 0, uniform(handleColor))
		drawBox(pose, Vec3{0, 0.31, -0.66}, Vec3{0.022, 0.03, 0.2}, 0, uniform(ironColor))
	} else {
		drawTexturedBox(pose, Vec3{0, 0.1, -0.74}, Vec3{0.11, 0.11, 0.11}, rad(35), item.Block, atlas)
	}

	rl.EndMode3D()
}

func cornerOffset(v [3]int, half Vec3) Vec3 {
	return Vec3{
		(float32(v[0])*2 - 1) * half[0],
		(float32(v[1])*2 - 1) * half[1],
		(float32(v[2])*2 - 1) * half[2],
	}
}

func drawBox(pose Pose, center, half Vec3, localYaw float32, colors faceColors) {
	for fi := range faces {
		face := &faces[fi]
		var p [4]rl.Vector3
		for k, v := range face.V {
			r := rotate(cornerOffset(v, half), 0, localYaw, 0)
			w := pose.apply(Vec3{center[0] + r[0], center[1] + r[1], center[2] + r[2]})
			p[k] = rl.NewVector3(w[0], w[1], w[2])
		}
		rgb := colors.side
		switch fi {
		case 0:
			rgb = colors.top
		case 1:
			rgb = colors.bottom
		}
		col := color.RGBA{
			uint8(float32(rgb.R) * face.Shade),
			uint8(float32(rgb.G) * face.Shade),
			uint8(float32(rgb.B) * face.Shade),
			255,
		}
		rl.DrawTriangle3D(p[0], p[1], p[2], col)
		rl.DrawTriangle3D(p[0], p[2], p[3], col)
	}
}

// drawTexturedBox draws a block cube using the world texture atlas.
func drawTexturedBox(pose Pose, center, half Vec3, localYaw float32, block Block, atlas rl.Texture2D) {
	rl.SetTexture(atlas.ID)
	rl.Begin(rl.Quads)
	for fi := range faces {
		face := &faces[fi]
		r := tileRect(tileOf(block, fi))
		shade := uint8(face.Shade * 255)
		rl.Color4ub(shade, shade, shade, 255)
		uvs := [4][2]float32{{r.U0, r.V1}, {r.U1, r.V1}, {r.U1, r.V0}, {r.U0, r.V0}}
		for k, v := range face.V {
			o := rotate(cornerOffset(v, half), 0, localYaw, 0)
			w := pose.apply(Vec3{center[0] + o[0], center[1] + o[1], center[2] + o[2]})
			rl.TexCoord2f(uvs[k][0], uvs[k][1])
			rl.Vertex3f(w[0], w[1], w[2])
		}
	}
	rl.End()
	rl.SetTexture(0)
}
