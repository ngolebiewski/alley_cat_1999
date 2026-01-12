package main

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type Camera struct {
	X, Y float64

	ViewportW int
	ViewportH int

	WorldW int
	WorldH int

	// Screen Shake intensity
	Shake float64
}

func NewCamera(viewW, ViewH, worldW, worldH int) *Camera {
	return &Camera{
		ViewportW: viewW,
		ViewportH: ViewH,
		WorldW:    worldW,
		WorldH:    worldH,
	}
}

// Update handles the decay of the screen shake
func (c *Camera) Update() {
	if c.Shake > 0 {
		c.Shake *= 0.9 // Reduce shake by 10% every frame
		if c.Shake < 0.1 {
			c.Shake = 0
		}
	}
}

func (c *Camera) Follow(x, y float64) {
	c.X = x - float64(c.ViewportW)/2
	c.Y = y - float64(c.ViewportH)/2
	c.clamp()
}

func (c *Camera) clamp() {
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}

	maxX := float64(c.WorldW - c.ViewportW)
	maxY := float64(c.WorldH - c.ViewportH)

	if c.X > maxX {
		c.X = maxX
	}
	if c.Y > maxY {
		c.Y = maxY
	}
}

// Apply now includes the random shake offset
func (c *Camera) Apply(op *ebiten.DrawImageOptions) {
	var offsetX, offsetY float64
	if c.Shake > 0 {
		// Random value between -Shake and +Shake
		offsetX = (rand.Float64()*2 - 1) * c.Shake
		offsetY = (rand.Float64()*2 - 1) * c.Shake
	}
	op.GeoM.Translate(-c.X+offsetX, -c.Y+offsetY)
}
