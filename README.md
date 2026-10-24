# Zube (Go)

A small Minecraft style voxel game written in Go with raylib, ported from the Zig version.

## Build

Needs Go 1.27 and a C compiler, since raylib is built from source through cgo. On Linux install the X11 or Wayland development headers and OpenGL.

```sh
make run        # Build and run
make release    # Stripped binary
make test       # Unit tests
```

Without `make`, use `go run .`, `go build` and `go test ./...`.

## Controls

| Input | Action |
|---|---|
| W A S D | Move |
| Mouse | Look |
| Space | Jump (fly up while flying) |
| Left Shift | Fly down |
| Left Ctrl | Sprint |
| F | Toggle flying |
| Left click (hold) | Mine the targeted block, cracks show the progress |
| Right click | Place the held block |
| 1 to 9 or mouse wheel | Select pickaxe or block |
| F3 | Toggle debug text |
| Esc | Quit |

## Mining

Each block has a hardness. Holding left click swings the tool and every swing that lands deals damage, shown as branching cracks that grow with each hit. The block breaks at full damage. The pickaxe is much faster on stone and cobblestone. A first person arm holds the selected item and swings while mining or placing.

## Textures

All block textures are generated at startup in `textures.go`, with no image files. Each 16 by 16 tile is built from tiling value noise, fractal noise and Worley (Voronoi) noise. For example cobblestone uses warped Voronoi cells with mortar lines, and grass layers fractal noise with blade streaks. The tiles are packed into one atlas.
