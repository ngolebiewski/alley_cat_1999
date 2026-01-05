package main

import (
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type TitleScene struct {
	game *Game
	img  *ebiten.Image
}

func NewTitleScene(game *Game) *TitleScene {

	return &TitleScene{
		game: game,
		img:  game.assets.TitleImage,
	}
}

func (s *TitleScene) Update() error {
	// Space → start race
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		s.game.scene = NewRaceScene(s.game)
	}
	return nil
}

func (s *TitleScene) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Alley Cat 1999\nPress Space to Start")

	op := &ebiten.DrawImageOptions{}
	size := s.img.Bounds().Size()
	op.GeoM.Translate(
		float64((320-size.X)/2),
		float64((240-size.Y)/2),
	)
	screen.DrawImage(s.img, op)
}
