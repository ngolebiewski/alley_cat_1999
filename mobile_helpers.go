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

// --- Mobile Helpers ---

type joystick struct {
	active       bool
	id           ebiten.TouchID
	baseX, baseY float64
	currX, currY float64
}

func (s *RaceScene) updateJoystick() {
	touches := ebiten.TouchIDs()
	if !s.stick.active {
		for _, id := range touches {
			x, y := ebiten.TouchPosition(id)
			if x < 160*zoom {
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

// Update the touch-detection to recognize the B button rectangle
func (s *RaceScene) isButtonPressed(label string) bool {
	touches := ebiten.TouchIDs()
	var r image.Rectangle
	switch label {
	case "A":
		r = image.Rect(270*zoom, 170*zoom, 310*zoom, 210*zoom)
	case "B":
		r = image.Rect(270*zoom, 170*zoom, 310*zoom, 210*zoom) // Added this range
	case "START":
		r = image.Rect(130*zoom, 210*zoom, 190*zoom, int(math.Floor(235.0*zoom)))
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
		r = image.Rect(270*zoom, 170*zoom, 310*zoom, 210*zoom)
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
	// Virtual Joystick
	if s.stick.active {
		vector.FillCircle(screen, float32(s.stick.baseX), float32(s.stick.baseY), 20, color.RGBA{255, 255, 255, 40}, true)
		vector.FillCircle(screen, float32(s.stick.currX), float32(s.stick.currY), 10, color.RGBA{255, 255, 255, 120}, true)
	}

	// B Button (Brake/Skid) - Positioned at (240, 190)
	vector.FillCircle(screen, 240*zoom, 190*zoom, 20*zoom, color.RGBA{180, 180, 0, 100}, true) // Yellowish-gold
	ebitenutil.DebugPrintAt(screen, "B", 234*zoom, 182*zoom)

	// A Button (Interact/Mount) - Positioned at (290, 190)
	vector.FillCircle(screen, 290*zoom, 190*zoom, 20*zoom, color.RGBA{200, 0, 0, 100}, true) // Red
	ebitenutil.DebugPrintAt(screen, "A", 284*zoom, 182*zoom)

	// Start Button
	vector.FillRect(screen, 135*zoom, 215*zoom, 50*zoom, 14*zoom, color.RGBA{100, 100, 100, 150}, true)
	ebitenutil.DebugPrintAt(screen, "START", 142*zoom, 214*zoom)
}
