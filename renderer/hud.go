package renderer

import (
	"fmt"
	"image/color"

	"github.com/tacf/goxel/game"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	white  = color.RGBA{255, 255, 255, 255}
	black  = color.RGBA{0, 0, 0, 255}
	shadow = color.RGBA{0, 0, 0, 120}
)

const (
	hotbarSlot = 48
	hotbarGap  = 4
)

// drawHUD draws the crosshair, the hotbar and the debug line in logical pixels.
func (r *Renderer) drawHUD(s *game.Session) {
	w, h := int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight())

	rl.DrawRectangle(w/2-10, h/2-1, 20, 2, white)
	rl.DrawRectangle(w/2-1, h/2-10, 2, 20, white)

	r.drawHotbar(s, w, h)
	if s.ShowDebug {
		drawDebug(s)
	}
}

func (r *Renderer) drawHotbar(s *game.Session, w, h int32) {
	total := int32(len(game.Hotbar))*(hotbarSlot+hotbarGap) - hotbarGap
	x := (w - total) / 2
	y := h - hotbarSlot - 16
	for i, item := range game.Hotbar {
		rl.DrawRectangle(x, y, hotbarSlot, hotbarSlot, shadow)
		if item.Kind == game.ItemBlock {
			src := tilePixels(iconTile(item.Block))
			rl.DrawTexturePro(
				r.atlas,
				rl.NewRectangle(src[0], src[1], src[2], src[3]),
				rl.NewRectangle(float32(x+6), float32(y+6), hotbarSlot-12, hotbarSlot-12),
				rl.NewVector2(0, 0), 0, white,
			)
		} else {
			drawPickaxeIcon(x, y, hotbarSlot)
		}
		if i == s.Selected {
			rl.DrawRectangleLinesEx(rl.NewRectangle(float32(x-2), float32(y-2), hotbarSlot+4, hotbarSlot+4), 3, white)
		}
		x += hotbarSlot + hotbarGap
	}
	name := s.Holding().String()
	rl.DrawText(name, (w-rl.MeasureText(name, 20))/2, y-28, 20, white)
}

func drawDebug(s *game.Session) {
	flying := ""
	if s.Player.Flying {
		flying = "  flying"
	}
	p := s.Player.Pos
	text := fmt.Sprintf("%d fps  xyz %.1f %.1f %.1f%s", rl.GetFPS(), p[0], p[1], p[2], flying)
	rl.DrawText(text, 11, 11, 20, black)
	rl.DrawText(text, 10, 10, 20, white)
}

func drawPickaxeIcon(x, y, slot int32) {
	fx, fy, fs := float32(x), float32(y), float32(slot)
	brown := color.RGBA{120, 88, 48, 255}
	iron := color.RGBA{170, 170, 178, 255}
	rl.DrawLineEx(rl.NewVector2(fx+fs*0.25, fy+fs*0.8), rl.NewVector2(fx+fs*0.7, fy+fs*0.3), 4, brown)
	rl.DrawLineEx(rl.NewVector2(fx+fs*0.4, fy+fs*0.2), rl.NewVector2(fx+fs*0.85, fy+fs*0.55), 6, iron)
}
