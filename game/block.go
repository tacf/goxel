package game

// Block is one voxel type.
type Block uint8

const (
	Air Block = iota
	Grass
	Dirt
	Stone
	Sand
	Wood
	Leaves
	Planks
	Cobble
	// BlockCount is the number of block types, Air included.
	BlockCount
)

var blockNames = [...]string{"Air", "Grass", "Dirt", "Stone", "Sand", "Wood", "Leaves", "Planks", "Cobblestone"}

func (b Block) String() string { return blockNames[b] }

// Hardness is the number of seconds needed to break the block with bare hands.
func (b Block) Hardness() float32 {
	switch b {
	case Grass:
		return 1.2
	case Dirt, Sand:
		return 1.0
	case Stone:
		return 3.0
	case Cobble:
		return 4.0
	case Wood:
		return 2.0
	case Planks:
		return 1.6
	case Leaves:
		return 0.4
	}
	return 0
}
