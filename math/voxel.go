package math

// Face describes one side of a cube.
type Face struct {
	N     [3]int
	V     [4][3]int
	Shade float32
}

// faces lists the cube sides with corners counter clockwise as seen from outside.
// The order is top, bottom, +x, -x, +z, -z.
var VoxelFaces = [6]Face{
	{[3]int{0, 1, 0}, [4][3]int{{0, 1, 0}, {0, 1, 1}, {1, 1, 1}, {1, 1, 0}}, 1.0},
	{[3]int{0, -1, 0}, [4][3]int{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}}, 0.5},
	{[3]int{1, 0, 0}, [4][3]int{{1, 0, 0}, {1, 1, 0}, {1, 1, 1}, {1, 0, 1}}, 0.8},
	{[3]int{-1, 0, 0}, [4][3]int{{0, 0, 0}, {0, 0, 1}, {0, 1, 1}, {0, 1, 0}}, 0.8},
	{[3]int{0, 0, 1}, [4][3]int{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}}, 0.65},
	{[3]int{0, 0, -1}, [4][3]int{{1, 0, 0}, {0, 0, 0}, {0, 1, 0}, {1, 1, 0}}, 0.65},
}
