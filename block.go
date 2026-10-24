package goxel

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

// tileOf picks the atlas tile for one face of a block, in the order of the faces table.
func TileOf(b Block, face int) Tile {
	switch b {
	case Grass:
		switch face {
		case 0:
			return TileGrassTop
		case 1:
			return TileDirt
		}
		return TileGrassSide
	case Dirt:
		return TileDirt
	case Stone:
		return TileStone
	case Sand:
		return TileSand
	case Cobble:
		return TileCobble
	case Planks:
		return TilePlanks
	case Leaves:
		return TileLeaves
	case Wood:
		if face < 2 {
			return TileWoodTop
		}
		return TileWoodSide
	}
	return TileStone
}

// iconTile is the tile shown for the block in the hotbar.
func iconTile(b Block) Tile {
	if b == Grass {
		return TileGrassTop
	}
	return TileOf(b, 2)
}
