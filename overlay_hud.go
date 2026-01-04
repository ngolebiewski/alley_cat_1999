package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type HUDOverlay struct {
	// Later:
	// timer
	// checkpoints
	// cash
	// energy
}

func NewHUDOverlay() *HUDOverlay {
	return &HUDOverlay{}
}

func (h *HUDOverlay) Update() error {
	return nil
}

func (h *HUDOverlay) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "TIME 01:23", 10, 10)
	ebitenutil.DebugPrintAt(screen, "CHECKPOINTS 1/3", 130, 10)
	ebitenutil.DebugPrintAt(screen, "$ 123", 260, 10)

}
