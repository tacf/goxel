# Zube (Go)

A small Minecraft style voxel game written in Go with raylib.

## Build

Needs Go 1.27 and a C compiler, since raylib is built from source through cgo. On Linux install the X11 or Wayland development headers and OpenGL.

```sh
make run        # Build and run
make release    # Stripped binary
make test       # Unit tests
```

Without `make`, use `go run .`, `go build` and `go test ./...`.

## Controls

| Input                 | Action                                            |
| --------------------- | ------------------------------------------------- |
| W A S D               | Move                                              |
| Mouse                 | Look                                              |
| Space                 | Jump (fly up while flying)                        |
| Left Shift            | Fly down                                          |
| Left Ctrl             | Sprint                                            |
| F                     | Toggle flying                                     |
| Left click (hold)     | Mine the targeted block, cracks show the progress |
| Right click           | Place the held block                              |
| 1 to 9 or mouse wheel | Select pickaxe or block                           |
| F3                    | Toggle debug text                                 |
| Esc                   | Quit                                              |

## Textures

All block textures are generated at startup in `textures.go`, with no image files. Each 16 by 16 tile is built from tiling value noise, fractal noise and Worley (Voronoi) noise. For example cobblestone uses warped Voronoi cells with mortar lines, and grass layers fractal noise with blade streaks. The world draws them from an OpenGL array texture with one layer per tile, so filtering and mipmaps never bleed between neighbouring tiles. A packed atlas is kept for the hotbar icons and the held block.
