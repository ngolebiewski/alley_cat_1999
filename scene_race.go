package main

import (
	"image"
	"image/color"

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

	bikerImg       *ebiten.Image
	bikerX, bikerY float64

	// animation
	bikerFrame     int
	bikerFrameTick int
	facingRight    bool
}

func NewRaceScene(game *Game) *RaceScene {
	return &RaceScene{
		game:        game,
		hud:         NewHUDOverlay(),
		bikerImg:    game.assets.BikerImage,
		bikerX:      50,
		bikerY:      120,
		facingRight: true,
	}
}

func (s *RaceScene) Update() error {
	const moveSpeed = 1.5
	moving := false

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		s.bikerX += moveSpeed
		moving = true
		s.facingRight = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		s.bikerX -= moveSpeed
		moving = true
		s.facingRight = false
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		s.bikerY -= moveSpeed
		moving = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		s.bikerY += moveSpeed
		moving = true
	}

	// animation
	s.bikerFrameTick++
	if moving {
		if s.bikerFrameTick%9 == 0 { // advance frame every 10 ticks
			s.bikerFrame++
			if s.bikerFrame > 2 { // loop moving frames 0,1,2
				s.bikerFrame = 0
			}
		}
	} else {
		s.bikerFrame = 3 // idle frame
		s.bikerFrameTick = 0
	}

	// temporary exit
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		s.game.scene = NewEndScene(s.game)
	}

	return nil
}

func (s *RaceScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 30, G: 30, B: 40, A: 220}) // dark grey
	ebitenutil.DebugPrintAt(screen, "RACE RUNNING SCENE\nPress ESC to exit\n'F' for Full Screen", 0, 100)

	const frameSize = 32
	op := &ebiten.DrawImageOptions{}

	// flip left/right
	if !s.facingRight {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(frameSize, 0) // adjust origin after flip
	}

	// move to position
	op.GeoM.Translate(s.bikerX, s.bikerY)

	// draw current frame
	sx := s.bikerFrame * frameSize
	screen.DrawImage(
		s.bikerImg.SubImage(image.Rect(sx, 0, sx+frameSize, frameSize)).(*ebiten.Image),
		op,
	)

	// draw HUD / overlay
	s.hud.Draw(screen)
	if s.overlay != nil {
		s.overlay.Draw(screen)
	}
}
