package main

// ItemKind tells apart what a hotbar slot holds.
type ItemKind uint8

const (
	ItemPickaxe ItemKind = iota
	ItemBlock
)

// Item is what the player holds. Block is only set for ItemBlock.
type Item struct {
	Kind  ItemKind
	Block Block
}

func (i Item) String() string {
	if i.Kind == ItemPickaxe {
		return "Pickaxe"
	}
	return i.Block.String()
}

// MineSpeed is the multiplier applied to the mining speed on the given block.
func (i Item) MineSpeed(target Block) float32 {
	if i.Kind != ItemPickaxe {
		return 1.0
	}
	if target == Stone || target == Cobble {
		return 4.0
	}
	return 1.5
}
