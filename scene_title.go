package main

import (
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
	// Space → start race
	if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		s.game.scene = NewRaceScene(s.game)
	}
	s.touchIDs = inpututil.AppendJustPressedTouchIDs(s.touchIDs[:0])
	if len(s.touchIDs) > 0 {
		s.game.scene = NewRaceScene(s.game)
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
