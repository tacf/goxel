package renderer

import (
	"math"
	"testing"
)

func TestCracksSpreadEvenlyAroundCentre(t *testing.T) {
	for _, seed := range []uint32{0x2545f491, 1, 2, 3, 4, 5, 6, 7} {
		var birth Birth
		generateCracks(&birth, seed)
		var sector [6]int
		minX, maxX, minY, maxY := crackTile, -1, crackTile, -1
		for y := range crackTile {
			for x := range crackTile {
				if birth[y][x] > 1.2 {
					continue
				}
				ang := math.Atan2(float64(float32(y)-crackCenter), float64(float32(x)-crackCenter)) + math.Pi
				sector[min(5, int(ang/(2*math.Pi)*6))]++
				minX, maxX = min(minX, x), max(maxX, x)
				minY, maxY = min(minY, y), max(maxY, y)
			}
		}
		for i, n := range sector {
			if n < 6 {
				t.Errorf("seed %d: sector %d has only %d crack pixels", seed, i, n)
			}
		}
		if maxX-minX < 11 || maxY-minY < 11 {
			t.Errorf("seed %d: cracks span %dx%d, want at least 12", seed, maxX-minX+1, maxY-minY+1)
		}
	}
}
