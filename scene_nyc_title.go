package main

import (
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ngolebiewski/alley_cat_1999/retrotrack"
)

type TitleSceneNYC struct {
	game     *Game
	img      *ebiten.Image
	touchIDs []ebiten.TouchID
}

func NewTitleSceneNYC(game *Game) *TitleSceneNYC {

	return &TitleSceneNYC{
		game: game,
		img:  game.assets.TitleImageNYC,
	}
}

func (s *TitleSceneNYC) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		retrotrack.PlayCityStartSound()
		s.game.scene = NewGetManifestScene(s.game)
	}
	s.touchIDs = inpututil.AppendJustPressedTouchIDs(s.touchIDs[:0])
	if len(s.touchIDs) > 0 {
		retrotrack.PlayCityStartSound()
		s.game.scene = NewGetManifestScene(s.game)
	}
	return nil
}

func (s *TitleSceneNYC) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "STAGE 1\nPress Space or Tap to Start")

	op := &ebiten.DrawImageOptions{}
	size := s.img.Bounds().Size()
	op.GeoM.Translate(
		float64((screenWidth-size.X)/2),
		float64((screenHeight-size.Y)/2),
	)
	screen.DrawImage(s.img, op)
}
