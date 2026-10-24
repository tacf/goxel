package goxel

import . "github.com/tacf/goxel/math"

const (
	crackTile = 16
	// crackNone marks a pixel that never cracks, it is above every damage level.
	crackNone   = 2.0
	crackCenter = 7.5
	// crackReach is the distance from the middle at which a crack reaches its full extent.
	crackReach = 9.5
	// ringRadius is the radius of the ring crack that joins neighbouring arms.
	ringRadius = 4.6
)

// Birth holds the damage level at which each pixel of the crack overlay appears.
type Birth [crackTile][crackTile]float32

// Damage level at which a pixel at this distance from the middle cracks.
func birthAt(x, y int) float32 {
	dx := float32(x) - crackCenter
	dy := float32(y) - crackCenter
	return 0.03 + 0.9*Sqrt(dx*dx+dy*dy)/crackReach
}

// xorshift is a tiny deterministic random generator.
type xorshift uint32

func (s *xorshift) next() uint32 {
	*s ^= *s << 13
	*s ^= *s >> 17
	*s ^= *s << 5
	return uint32(*s)
}

func (s *xorshift) float() float32 { return float32(s.next()>>8) / 16777216.0 }

// generateCracks fills the birth map: arms spread evenly around the impact, fork and are joined by a ring.
func generateCracks(birth *Birth, seed uint32) {
	for y := range birth {
		for x := range birth[y] {
			birth[y][x] = crackNone
		}
	}
	state := xorshift(seed)

	const arms = 7
	var angles [arms]float32
	spin := state.float() * 2 * PI
	for k := range arms {
		// One arm per equal sector, wobbled inside it so no direction is favoured.
		slot := (float32(k) + 0.5 + (state.float()-0.5)*0.25) / arms
		angles[k] = spin + slot*2*PI
		steps := 9 + int(state.next()%2)
		walkCrack(birth, &state, crackCenter, crackCenter, angles[k], steps, 0)
	}

	// A shorter ring crack that circles the impact and joins neighbouring arms.
	for k := range arms {
		if state.float() < 0.5 {
			continue
		}
		from := angles[k] + 0.12
		var wrapAngle float32
		if k+1 == arms {
			wrapAngle = 2 * PI
		}
		to := angles[(k+1)%arms] + wrapAngle - 0.12
		span := to - from
		count := int(Ceil(Abs(span) * ringRadius))
		a, r := from, float32(ringRadius)
		for range count {
			a += span / float32(count)
			r += (state.float() - 0.5) * 0.7
			markCrack(birth, int(Floor(crackCenter+Cos(a)*r)), int(Floor(crackCenter+Sin(a)*r)))
		}
	}
}

// walkCrack draws a wandering crack that keeps roughly to its heading.
func walkCrack(birth *Birth, state *xorshift, x0, y0, heading float32, steps, depth int) {
	x, y, angle := x0, y0, heading
	for i := range steps {
		angle += (state.float()-0.5)*0.6 + (heading-angle)*0.25
		x += Cos(angle)
		y += Sin(angle)
		px, py := int(Floor(x)), int(Floor(y))
		if px < 0 || py < 0 || px >= crackTile || py >= crackTile {
			return
		}
		markCrack(birth, px, py)
		if depth < 2 && i >= 2 && steps-i > 2 && state.float() < 0.22 {
			side := float32(-1)
			if state.next()%2 == 0 {
				side = 1
			}
			walkCrack(birth, state, x, y, angle+side*(0.5+state.float()*0.7), min(steps-i, 5), depth+1)
		}
	}
}

func markCrack(birth *Birth, x, y int) {
	if x < 0 || y < 0 || x >= crackTile || y >= crackTile {
		return
	}
	birth[y][x] = min(birth[y][x], birthAt(x, y))
}
