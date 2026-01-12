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
	manifest *Manifest
}

func NewGameOverScene(game *Game, manifest *Manifest) *GameOverScene {
	resetManifestCheckins(manifest) // since we are doing the same manifest, we need to reset the checkins
	return &GameOverScene{game: game, manifest: manifest}
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
				retrotrack.Start()
				s.game.scene = NewRaceScene(s.game, s.manifest) // Restart Level
			}
		}
		s.touchIDs = inpututil.AppendJustPressedTouchIDs(s.touchIDs[:0])
		if len(s.touchIDs) > 0 {
			retrotrack.PlayCityStartSound()
			retrotrack.Start()
			s.game.scene = NewRaceScene(s.game, s.manifest)
		}

	}
	return nil
}

func (s *GameOverScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 0, 0, 255}) // Dark red background

	msg := "GAME OVER\n\nPLAY AGAIN?"
	ebitenutil.DebugPrintAt(screen, msg, 80, 100)
}
