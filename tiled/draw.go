package tiled

import (
	"image"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	flipH    = 0x80000000
	flipV    = 0x40000000
	flipD    = 0x20000000
	flipMask = flipH | flipV | flipD
)

type Renderer struct {
	Map     *Map
	Tileset *ebiten.Image
	Scale   float64
}

func NewRenderer(m *Map, tileset *ebiten.Image, scale float64) *Renderer {
	return &Renderer{
		Map:     m,
		Tileset: tileset,
		Scale:   scale,
	}
}

func (r *Renderer) Draw(screen *ebiten.Image) {
	for _, layer := range r.Map.Layers {
		r.drawLayer(screen, layer)
	}
}

func (r *Renderer) drawLayer(screen *ebiten.Image, layer Layer) {
	if !layer.Visible {
		return
	}

	if layer.Type == "group" {
		for _, sub := range layer.Layers {
			r.drawLayer(screen, sub)
		}
		return
	}

	if layer.Type != "tilelayer" {
		return
	}

	tilesPerRow := r.Tileset.Bounds().Dx() / r.Map.TileWidth

	for i, raw := range layer.Data {
		if raw == 0 {
			continue
		}

		gid := raw &^ flipMask
		tileIndex := int(gid - 1)

		sx := (tileIndex % tilesPerRow) * r.Map.TileWidth
		sy := (tileIndex / tilesPerRow) * r.Map.TileHeight

		tile := r.Tileset.SubImage(
			image.Rect(
				sx,
				sy,
				sx+r.Map.TileWidth,
				sy+r.Map.TileHeight,
			),
		).(*ebiten.Image)

		x := i % r.Map.Width
		y := i / r.Map.Width

		op := &ebiten.DrawImageOptions{}

		// flips
		if raw&flipH != 0 {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(r.Map.TileWidth), 0)
		}
		if raw&flipV != 0 {
			op.GeoM.Scale(1, -1)
			op.GeoM.Translate(0, float64(r.Map.TileHeight))
		}

		op.GeoM.Scale(r.Scale, r.Scale)
		op.GeoM.Translate(
			float64(x*r.Map.TileWidth)*r.Scale,
			float64(y*r.Map.TileHeight)*r.Scale,
		)

		screen.DrawImage(tile, op)
	}
}

// Utility for later collision logic
func IsCollideLayer(name string) bool {
	return strings.Contains(strings.ToUpper(name), "COLLIDE")
}
