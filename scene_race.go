// package main

// import (
// 	"image"
// 	"image/color"

// 	"github.com/hajimehoshi/ebiten/v2"
// 	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
// )

// type Overlay interface {
// 	Update() error
// 	Draw(screen *ebiten.Image)
// }

// type RaceScene struct {
// 	game    *Game
// 	hud     *HUDOverlay
// 	overlay Overlay // nil = no overlay

// 	bikerImg       *ebiten.Image
// 	bikerX, bikerY float64

// 	// animation
// 	bikerFrame     int
// 	bikerFrameTick int
// 	facingRight    bool
// }

// func NewRaceScene(game *Game) *RaceScene {
// 	return &RaceScene{
// 		game:        game,
// 		hud:         NewHUDOverlay(),
// 		bikerImg:    game.assets.BikerImage,
// 		bikerX:      50,
// 		bikerY:      120,
// 		facingRight: true,
// 	}
// }

// func (s *RaceScene) Update() error {
// 	const moveSpeed = 1.5
// 	moving := false

// 	if ebiten.IsKeyPressed(ebiten.KeyRight) {
// 		s.bikerX += moveSpeed
// 		moving = true
// 		s.facingRight = true
// 	}
// 	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
// 		s.bikerX -= moveSpeed
// 		moving = true
// 		s.facingRight = false
// 	}
// 	if ebiten.IsKeyPressed(ebiten.KeyUp) {
// 		s.bikerY -= moveSpeed
// 		moving = true
// 	}
// 	if ebiten.IsKeyPressed(ebiten.KeyDown) {
// 		s.bikerY += moveSpeed
// 		moving = true
// 	}

// 	// animation
// 	s.bikerFrameTick++
// 	if moving {
// 		if s.bikerFrameTick%9 == 0 { // advance frame every 10 ticks
// 			s.bikerFrame++
// 			if s.bikerFrame > 2 { // loop moving frames 0,1,2
// 				s.bikerFrame = 0
// 			}
// 		}
// 	} else {
// 		s.bikerFrame = 3 // idle frame
// 		s.bikerFrameTick = 0
// 	}

// 	// temporary exit
// 	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
// 		s.game.scene = NewEndScene(s.game)
// 	}

// 	return nil
// }

// func (s *RaceScene) Draw(screen *ebiten.Image) {
// 	screen.Fill(color.RGBA{R: 30, G: 30, B: 40, A: 220}) // dark grey
// 	ebitenutil.DebugPrintAt(screen, "RACE RUNNING SCENE\nPress ESC to exit\n'F' for Full Screen", 0, 100)

// 	const frameSize = 32
// 	op := &ebiten.DrawImageOptions{}

// 	// flip left/right
// 	if !s.facingRight {
// 		op.GeoM.Scale(-1, 1)
// 		op.GeoM.Translate(frameSize, 0) // adjust origin after flip
// 	}

// 	// move to position
// 	op.GeoM.Translate(s.bikerX, s.bikerY)

// 	// draw current frame
// 	sx := s.bikerFrame * frameSize
// 	screen.DrawImage(
// 		s.bikerImg.SubImage(image.Rect(sx, 0, sx+frameSize, frameSize)).(*ebiten.Image),
// 		op,
// 	)

// 	// draw HUD / overlay
// 	s.hud.Draw(screen)
// 	if s.overlay != nil {
// 		s.overlay.Draw(screen)
// 	}
// }

package main

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type BikerState int

const (
	StateRiding BikerState = iota
	StateWalking
)

type joystick struct {
	active       bool
	id           ebiten.TouchID
	baseX, baseY float64
	currX, currY float64
}

type RaceScene struct {
	game   *Game
	hud    *HUDOverlay
	state  BikerState
	paused bool

	bikerImg       *ebiten.Image
	bikerX, bikerY float64
	velX, velY     float64

	dir            int // 0: Down, 1: Up, 2: Left, 3: Right
	bikerFrame     int
	bikerFrameTick int

	stick joystick
}

func NewRaceScene(game *Game) *RaceScene {
	return &RaceScene{
		game:     game,
		hud:      NewHUDOverlay(),
		bikerImg: game.assets.BikerImage,
		bikerX:   160,
		bikerY:   120,
		state:    StateRiding,
		dir:      3, // Default facing Right
	}
}

func (s *RaceScene) Update() error {
	// 1. Scene Navigation & Pause
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.game.scene = NewEndScene(s.game)
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || s.isButtonPressed("START") {
		s.paused = !s.paused
	}
	if s.paused {
		return nil
	}

	const accel = 0.2
	const friction = 0.92
	const walkSpeed = 1.2

	moving := false
	var inputX, inputY float64

	// 2. Input Gathering
	jx, jy := s.getJoystickVector()
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || jx < -0.2 {
		inputX = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || jx > 0.2 {
		inputX = 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || jy < -0.2 {
		inputY = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || jy > 0.2 {
		inputY = 1
	}

	// 3. A Button: Dismount Logic
	if inpututil.IsKeyJustPressed(ebiten.KeyX) || s.isButtonJustPressed("A") {
		currentVel := math.Sqrt(s.velX*s.velX + s.velY*s.velY)
		if currentVel < 0.5 {
			if s.state == StateRiding {
				s.state = StateWalking
			} else {
				s.state = StateRiding
			}
		}
	}

	// 4. Movement Physics
	if s.state == StateRiding {
		s.velX += inputX * accel
		s.velY += inputY * accel
		s.velX *= friction
		s.velY *= friction
		s.bikerX += s.velX
		s.bikerY += s.velY

		if math.Abs(s.velX) > 0.1 || math.Abs(s.velY) > 0.1 {
			moving = true
			s.updateDirection(inputX, inputY)
		}
	} else {
		if inputX != 0 || inputY != 0 {
			s.bikerX += inputX * walkSpeed
			s.bikerY += inputY * walkSpeed
			moving = true
			s.updateDirection(inputX, inputY)
		}
	}

	s.updateAnimation(moving)

	if isMobile {
		s.updateJoystick()
	}
	return nil
}

func (s *RaceScene) updateDirection(x, y float64) {
	if math.Abs(x) > math.Abs(y) {
		if x > 0 {
			s.dir = 3
		} else {
			s.dir = 2
		}
	} else {
		if y > 0 {
			s.dir = 0
		} else {
			s.dir = 1
		}
	}
}

func (s *RaceScene) updateAnimation(moving bool) {
	s.bikerFrameTick++

	if s.state == StateRiding {
		if moving {
			if s.bikerFrameTick%8 == 0 {
				switch s.dir {
				case 3, 2: // Side to Side
					s.bikerFrame = (s.bikerFrame + 1) % 3 // 0, 1, 2
				case 1: // Up
					if s.bikerFrame < 4 || s.bikerFrame > 5 {
						s.bikerFrame = 4
					}
					s.bikerFrame = 4 + (s.bikerFrameTick/8)%2
				case 0: // Down
					if s.bikerFrame < 6 || s.bikerFrame > 7 {
						s.bikerFrame = 6
					}
					s.bikerFrame = 6 + (s.bikerFrameTick/8)%2
				}
			}
		} else {
			s.bikerFrame = 3 // Standard bike idle
		}
	} else {
		if moving {
			if s.bikerFrameTick%10 == 0 {
				switch s.dir {
				case 3, 2:
					if s.bikerFrame < 8 || s.bikerFrame > 9 {
						s.bikerFrame = 8
					}
					s.bikerFrame = 8 + (s.bikerFrameTick/10)%2
				case 1:
					if s.bikerFrame < 10 || s.bikerFrame > 11 {
						s.bikerFrame = 10
					}
					s.bikerFrame = 10 + (s.bikerFrameTick/10)%2
				case 0:
					if s.bikerFrame < 11 || s.bikerFrame > 12 {
						s.bikerFrame = 11
					}
					s.bikerFrame = 11 + (s.bikerFrameTick/10)%2
				}
			}
		} else {
			// Walking Idles
			switch s.dir {
			case 3, 2:
				s.bikerFrame = 13
			case 1:
				s.bikerFrame = 14
			case 0:
				s.bikerFrame = 15
			}
		}
	}
}

func (s *RaceScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 40, 255})

	const size = 32
	op := &ebiten.DrawImageOptions{}

	// Apply Mirroring for Left (Dir 2)
	if s.dir == 2 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(size, 0)
	}

	op.GeoM.Translate(s.bikerX, s.bikerY)

	// Draw Sprite
	sx := s.bikerFrame * size
	screen.DrawImage(s.bikerImg.SubImage(image.Rect(sx, 0, sx+size, size)).(*ebiten.Image), op)

	// HUD and UI
	s.hud.Draw(screen)
	if isMobile {
		s.drawMobileUI(screen)
	}

	if s.paused {
		ebitenutil.DebugPrintAt(screen, "PAUSED", 140, 110)
	}
}

// --- Mobile Helpers ---

func (s *RaceScene) updateJoystick() {
	touches := ebiten.TouchIDs()
	if !s.stick.active {
		for _, id := range touches {
			x, y := ebiten.TouchPosition(id)
			if x < 160 {
				s.stick.active = true
				s.stick.id = id
				s.stick.baseX, s.stick.baseY = float64(x), float64(y)
				s.stick.currX, s.stick.currY = float64(x), float64(y)
				break
			}
		}
	} else {
		found := false
		for _, id := range touches {
			if id == s.stick.id {
				x, y := ebiten.TouchPosition(id)
				s.stick.currX, s.stick.currY = float64(x), float64(y)
				found = true
				break
			}
		}
		if !found {
			s.stick.active = false
		}
	}
}

func (s *RaceScene) getJoystickVector() (float64, float64) {
	if !s.stick.active {
		return 0, 0
	}
	dx, dy := s.stick.currX-s.stick.baseX, s.stick.currY-s.stick.baseY
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 4 {
		return 0, 0
	}
	return dx / dist, dy / dist
}

func (s *RaceScene) isButtonPressed(label string) bool {
	touches := ebiten.TouchIDs()
	var r image.Rectangle
	switch label {
	case "A":
		r = image.Rect(270, 170, 310, 210)
	case "B":
		r = image.Rect(220, 170, 260, 210)
	case "START":
		r = image.Rect(130, 210, 190, 235)
	}
	for _, id := range touches {
		x, y := ebiten.TouchPosition(id)
		if image.Pt(x, y).In(r) {
			return true
		}
	}
	return false
}

func (s *RaceScene) isButtonJustPressed(label string) bool {
	touches := inpututil.AppendJustPressedTouchIDs(nil)
	var r image.Rectangle
	if label == "A" {
		r = image.Rect(270, 170, 310, 210)
	}
	for _, id := range touches {
		x, y := ebiten.TouchPosition(id)
		if image.Pt(x, y).In(r) {
			return true
		}
	}
	return false
}
func (s *RaceScene) drawMobileUI(screen *ebiten.Image) {
	if s.stick.active {
		vector.FillCircle(screen, float32(s.stick.baseX), float32(s.stick.baseY), 20, color.RGBA{255, 255, 255, 40}, true)
		vector.FillCircle(screen, float32(s.stick.currX), float32(s.stick.currY), 10, color.RGBA{255, 255, 255, 120}, true)
	}
	// A Button
	vector.FillCircle(screen, 290, 190, 20, color.RGBA{200, 0, 0, 100}, true)
	ebitenutil.DebugPrintAt(screen, "A", 285, 182)
	// Start Button
	vector.DrawFilledRect(screen, 135, 215, 50, 15, color.RGBA{100, 100, 100, 150}, true)
	ebitenutil.DebugPrintAt(screen, "START", 142, 215)
}
