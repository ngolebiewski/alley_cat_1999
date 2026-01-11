package main

import (
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ngolebiewski/alley_cat_1999/retrotrack"
)

const tileSize = 16

type GetManifestScene struct {
	game        *Game
	tileset     *ebiten.Image
	manifestImg *ebiten.Image
	touchIDs    []ebiten.TouchID
	animTime    float64 // seconds elapsed since start of animation
	animDone    bool    // stop animating after 1 second
}

func NewGetManifestScene(game *Game) *GetManifestScene {
	tileset := game.assets.TilesetImage

	return &GetManifestScene{
		game:        game,
		tileset:     tileset,
		manifestImg: buildManifestImage(tileset, tileSize),
	}
}

func (s *GetManifestScene) Update() error {
	// Increment animation time if not done
	if !s.animDone {
		s.animTime += 1.0 / 60.0 // 1 tick = 1/60 second
		if s.animTime >= 1.0 {
			s.animTime = 1.0
			s.animDone = true
			retrotrack.PlayManifestSound()
		}
	}

	// Space / click → start race
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		retrotrack.Start()
		s.game.scene = NewRaceScene(s.game)
	}
	s.touchIDs = inpututil.AppendJustPressedTouchIDs(s.touchIDs[:0])
	if len(s.touchIDs) > 0 {
		retrotrack.Start()
		s.game.scene = NewRaceScene(s.game)
	}

	return nil
}

func (s *GetManifestScene) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(
		screen,
		"Here's your new Manifest.\nThese are all the checkpoints you need to complete\nbefore heading to the finish line!",
	)

	finalScale := 5.0

	// animation progress 0 → 1
	t := s.animTime / 1.0
	if t > 1 {
		t = 1
	}

	// simple ease-out
	ease := t * (2 - t)

	scale := finalScale * ease
	rotation := (1 - ease) * 2 * 3.1415926 // 1 full spin

	w := s.manifestImg.Bounds().Dx()
	h := s.manifestImg.Bounds().Dy()

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest

	op.GeoM.Translate(-float64(w)/2, -float64(h)/2) // center origin
	op.GeoM.Rotate(rotation)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(screenWidth)/2, float64(screenHeight)/2) // back to center

	screen.DrawImage(s.manifestImg, op)
}

func buildManifestImage(
	tileset *ebiten.Image,
	tileSize int,
) *ebiten.Image {

	// 2x2 tiles → 32x32 image
	img := ebiten.NewImage(tileSize*2, tileSize*2)

	tilesetWidth := tileset.Bounds().Dx()

	tileIDs := []int{
		203, 204,
		205, 206,
	}

	for i, tileID := range tileIDs {
		col := i % 2
		row := i / 2

		tilesPerRow := tilesetWidth / tileSize
		x := (tileID % tilesPerRow) * tileSize
		y := (tileID / tilesPerRow) * tileSize

		src := image.Rect(x, y, x+tileSize, y+tileSize)
		sub := tileset.SubImage(src).(*ebiten.Image)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			float64(col*tileSize),
			float64(row*tileSize),
		)

		img.DrawImage(sub, op)
	}

	return img
}
