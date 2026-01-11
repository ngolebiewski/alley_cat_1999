package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	FadeIn = iota
	FadeOut
)

type Fader struct {
	alpha    float32
	speed    float32
	mode     int
	Finished bool
}

// NewFader creates a fader. Mode: FadeIn (1->0) or FadeOut (0->1)
func NewFader(mode int, seconds float64) *Fader {
	initialAlpha := float32(0.0)
	if mode == FadeIn {
		initialAlpha = 1.0
	}
	return &Fader{
		alpha:    initialAlpha,
		speed:    float32(1.0 / (60.0 * seconds)),
		mode:     mode,
		Finished: false,
	}
}

func (f *Fader) Update() {
	if f.Finished {
		return
	}

	if f.mode == FadeIn {
		f.alpha -= f.speed
		if f.alpha <= 0 {
			f.alpha = 0
			f.Finished = true
		}
	} else {
		f.alpha += f.speed
		if f.alpha >= 1 {
			f.alpha = 1
			f.Finished = true
		}
	}
}

func (f *Fader) Draw(screen *ebiten.Image) {
	// Optimization: don't draw if fully transparent
	if f.alpha <= 0 && f.mode == FadeIn {
		return
	}

	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	ebitenutil.DrawRect(screen, 0, 0, float64(w), float64(h), color.NRGBA{0, 0, 0, uint8(f.alpha * 255)})
}
