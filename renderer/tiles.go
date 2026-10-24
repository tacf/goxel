package renderer

import "github.com/tacf/goxel/game"

// TileOf picks the atlas tile for one face of a block, in the order of the faces table.
func TileOf(b game.Block, face int) Tile {
	switch b {
	case game.Grass:
		switch face {
		case 0:
			return TileGrassTop
		case 1:
			return TileDirt
		}
		return TileGrassSide
	case game.Dirt:
		return TileDirt
	case game.Stone:
		return TileStone
	case game.Sand:
		return TileSand
	case game.Cobble:
		return TileCobble
	case game.Planks:
		return TilePlanks
	case game.Leaves:
		return TileLeaves
	case game.Wood:
		if face < 2 {
			return TileWoodTop
		}
		return TileWoodSide
	}
	return TileStone
}

// iconTile is the tile shown for the block in the hotbar.
func iconTile(b game.Block) Tile {
	if b == game.Grass {
		return TileGrassTop
	}
	return TileOf(b, 2)
}
