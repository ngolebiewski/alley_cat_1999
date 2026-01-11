package main

import (
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ngolebiewski/alley_cat_1999/retrotrack"
)

type TitleScene struct {
	game     *Game
	img      *ebiten.Image
	touchIDs []ebiten.TouchID
}

func NewTitleScene(game *Game) *TitleScene {

	return &TitleScene{
		game: game,
		img:  game.assets.TitleImage,
	}
}

func (s *TitleScene) Update() error {
	// Space → Stage Title
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		retrotrack.PlayStartSound()
		s.game.scene = NewTitleSceneNYC(s.game)
	}
	s.touchIDs = inpututil.AppendJustPressedTouchIDs(s.touchIDs[:0])
	if len(s.touchIDs) > 0 {
		isMobile = true // if you touch to activate the game, you probably want the virtual joystick!
		retrotrack.PlayStartSound()
		s.game.scene = NewTitleSceneNYC(s.game)
	}
	return nil
}

func (s *TitleScene) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Alley Cat 1999\nPress Space or Tap to Start")

	op := &ebiten.DrawImageOptions{}
	size := s.img.Bounds().Size()
	op.GeoM.Translate(
		float64((screenWidth-size.X)/2),
		float64((screenHeight-size.Y)/2),
	)
	screen.DrawImage(s.img, op)
}
