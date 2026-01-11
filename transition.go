package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Fader struct {
	alpha    float32 // 1.0 = fully black, 0.0 = fully transparent
	speed    float32 // How much alpha to remove per frame
	finished bool
}

func NewFadeIn(seconds float64) *Fader {
	// Standard Ebitengine TPS is 60.
	// Speed = 1.0 total alpha / (60 frames * seconds)
	return &Fader{
		alpha: 1.0,
		speed: float32(1.0 / (60.0 * seconds)),
	}
}

func (f *Fader) Update() {
	if f.alpha > 0 {
		f.alpha -= f.speed
	} else {
		f.alpha = 0
		f.finished = true
	}
}

func (f *Fader) Draw(screen *ebiten.Image) {
	if f.alpha > 0 {
		// Draw a black rectangle over the whole screen
		// The alpha value is converted to 0-255
		ebitenutil.DrawRect(screen, 0, 0, float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy()), color.NRGBA{0, 0, 0, uint8(f.alpha * 255)})
	}
}
