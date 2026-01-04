package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type RaceScene struct {
	game *Game
}

func NewRaceScene(game *Game) *RaceScene {
	return &RaceScene{game: game}
}

func (s *RaceScene) Update() error {
	// Escape → end scene (for now)
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		s.game.scene = NewEndScene(s.game)
	}
	return nil
}

func (s *RaceScene) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "RACE RUNNING")
}
