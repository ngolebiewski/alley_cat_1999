package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ngolebiewski/alley_cat_1999/retrotrack"
)

type GameOverScene struct {
	game     *Game
	counter  int
	touchIDs []ebiten.TouchID
}

func NewGameOverScene(game *Game) *GameOverScene {
	return &GameOverScene{game: game}
}

func (s *GameOverScene) Update() error {
	s.counter++
	if s.counter < 1 {
		retrotrack.PlayGameOverSound()
	}
	if s.counter > 60 {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
			{
				retrotrack.PlayCityStartSound()
				s.game.scene = NewRaceScene(s.game) // Restart
			}
		}
		s.touchIDs = inpututil.AppendJustPressedTouchIDs(s.touchIDs[:0])
		if len(s.touchIDs) > 0 {
			retrotrack.PlayCityStartSound()
			s.game.scene = NewRaceScene(s.game)
		}

	}
	return nil
}

func (s *GameOverScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 0, 0, 255}) // Dark red background

	msg := "GAME OVER\n\nPLAY AGAIN?"
	ebitenutil.DebugPrintAt(screen, msg, 80, 100)
}
