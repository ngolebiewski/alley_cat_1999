package main

import "github.com/hajimehoshi/ebiten/v2"

type Camera struct {
	X, Y float64

	ViewportW int
	ViewportH int

	WorldW int
	WorldH int
}

// Constructor for Camera
func NewCamera(viewW, ViewH, worldW, worldH int) *Camera {
	return &Camera{
		ViewportW: viewW,
		ViewportH: ViewH,
		WorldW:    worldW,
		WorldH:    worldH,
	}
}

// Follow the player
func (c *Camera) Follow(x, y float64) {
	c.X = x - float64(c.ViewportW)/2
	c.Y = y - float64(c.ViewportH)/2

	c.clamp()
}

// Clamp to camera bounds, i.e. don't bike off the screen.
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

// Camera offset to draw ops
func (c *Camera) Apply(op *ebiten.DrawImageOptions) {
	op.GeoM.Translate(-c.X, -c.Y)
}
