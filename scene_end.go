package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ngolebiewski/alley_cat_1999/retrotrack"
)

type EndScene struct {
	game     *Game
	time     string
	cash     int
	touchIDs []ebiten.TouchID
}

func NewEndScene(game *Game, time string, cash int) *EndScene {
	return &EndScene{game: game,
		time: time,
		cash: cash,
	}
}

func (s *EndScene) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		{
			retrotrack.PlayCityStartSound()
			s.game.scene = NewTitleScene(s.game)
		}

	}
	s.touchIDs = inpututil.AppendJustPressedTouchIDs(s.touchIDs[:0])
	if len(s.touchIDs) > 0 {
		retrotrack.PlayCityStartSound()
		s.game.scene = NewTitleScene(s.game)
	}
	return nil
}

func (s *EndScene) Draw(screen *ebiten.Image) {
	results := fmt.Sprintf(
		"RACE OVER\n\n"+
			"YOUR TIME: %s\n"+
			"CASH EARNED: $%d\n\n"+
			"--- Leaderboard ---\n"+
			"CC 01:10:18\n"+
			"AL 01:10:19\n"+
			"NG 01:30:45\n"+
			"HH 01:34:12\n"+
			"DFL: DT\n\n"+
			"Press [ENTER] to Restart\n\n"+
			"Game: https://github.com/ngolebiewski/alley_cat_1999",
		s.time, s.cash,
	)

	ebitenutil.DebugPrint(screen, results)
}
