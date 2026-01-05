package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type EndScene struct {
	game *Game
}

func NewEndScene(game *Game) *EndScene {
	return &EndScene{game: game}
}

func (s *EndScene) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		s.game.scene = NewTitleScene(s.game)
	}
	return nil
}

func (s *EndScene) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "RACE OVER\n\nBest Times\nCC 01:10:18\nAL 01:10:19\nNG 01:30:45\nHH 1:34:12\nDFL: DT\n\n**Press Enter**\n\nGame: https://github.com/ngolebiewski/alley_cat_1999")
}
