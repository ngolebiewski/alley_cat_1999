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
	ebitenutil.DebugPrint(screen, "RACE OVER\nPress Enter")
}
