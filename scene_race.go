package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Overlay interface {
	Update() error
	Draw(screen *ebiten.Image)
}

type RaceScene struct {
	game    *Game
	hud     *HUDOverlay
	overlay Overlay // nil = no overlay
}

func NewRaceScene(game *Game) *RaceScene {
	return &RaceScene{
		game: game,
		hud:  NewHUDOverlay(),
	}
}

func (s *RaceScene) Update() error {
	// Escape → end scene (for now)
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		s.game.scene = NewEndScene(s.game)
	}
	return nil
}

func (s *RaceScene) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "RACE RUNNING\nTo be coded :)\nPress ESC to continue", 0, 100)
	s.hud.Draw(screen)
	if s.overlay != nil {
		s.overlay.Draw(screen)
	}
}
