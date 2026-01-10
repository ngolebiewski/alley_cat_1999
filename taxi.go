package main

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type Taxi struct {
	x, y          float64         // current position
	speed         float64         // movement speed
	scale         float64         // scale factor for drawing
	frames        []*ebiten.Image // animation frames
	frame         int
	frameTick     int
	dir           string       // "LEFT", "RIGHT", "UP", "DOWN"
	laneX, laneY  float64      // lane position for perpendicular wiggle
	width, height float64      // sprite dimensions
	manager       *TaxiManager // reference to TaxiManager
}

// Constructor
func NewTaxi(manager *TaxiManager, frames []*ebiten.Image, x, y, speed float64, dir string, scale float64) *Taxi {
	w := float64(frames[0].Bounds().Dx())
	h := float64(frames[0].Bounds().Dy())
	return &Taxi{
		manager: manager,
		frames:  frames,
		x:       x,
		y:       y,
		speed:   speed,
		dir:     dir,
		scale:   scale,
		width:   w,
		height:  h,
		laneX:   x,
		laneY:   y,
	}
}

// Random float helper
func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// Update taxi each frame
func (t *Taxi) Update() {
	const wiggle = 0.3
	const frameSpeed = 8

	// Wiggle along perpendicular axis
	if t.dir == "LEFT" || t.dir == "RIGHT" {
		t.y += randFloat(-wiggle, wiggle)
	} else {
		t.x += randFloat(-wiggle, wiggle)
	}

	// Move if no taxi in front
	if !t.manager.TaxiInFront(t) {
		switch t.dir {
		case "RIGHT":
			t.x += t.speed
		case "LEFT":
			t.x -= t.speed
		case "UP":
			t.y -= t.speed
		case "DOWN":
			t.y += t.speed
		}
	}

	// Animate frames
	t.frameTick++
	if t.frameTick%frameSpeed == 0 {
		t.frame = (t.frame + 1) % len(t.frames)
	}

	// Check world bounds → respawn if out
	tWidth := t.width * t.scale
	tHeight := t.height * t.scale

	out := false
	switch t.dir {
	case "LEFT":
		if t.x+tWidth < 0 {
			out = true
		}
	case "RIGHT":
		if t.x > t.manager.worldW {
			out = true
		}
	case "UP":
		if t.y+tHeight < 0 {
			out = true
		}
	case "DOWN":
		if t.y > t.manager.worldH {
			out = true
		}
	}

	if out {
		t.Respawn()
	}
}

// Draw taxi on screen with camera offset
func (t *Taxi) Draw(screen *ebiten.Image, scale float64, cam *Camera) {
	op := &ebiten.DrawImageOptions{}

	switch t.dir {
	case "RIGHT":
		// flip horizontally
		op.GeoM.Scale(-scale, scale)
		op.GeoM.Translate(-t.width*scale, 0)
	case "DOWN":
		// flip vertically
		op.GeoM.Scale(scale, -scale)
		op.GeoM.Translate(0, -t.height*scale)
	default:
		op.GeoM.Scale(scale, scale)
	}

	// camera offset
	op.GeoM.Translate(-cam.X, -cam.Y)
	op.GeoM.Translate(t.x, t.y)

	// choose frame to draw
	img := t.frames[t.frame]
	screen.DrawImage(img, op)
}

// Respawn taxi at opposite side
func (t *Taxi) Respawn() {
	t.frame = 0
	t.frameTick = 0

	// Pick a spawn side based on direction
	switch t.dir {
	case "LEFT":
		t.x = t.manager.worldW
	case "RIGHT":
		t.x = -t.width * t.scale
	case "UP":
		t.y = t.manager.worldH
	case "DOWN":
		t.y = -t.height * t.scale
	}

	// Reset perpendicular lane
	if t.dir == "LEFT" || t.dir == "RIGHT" {
		t.y = t.laneY
	} else {
		t.x = t.laneX
	}
}
