package main

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type HUDOverlay struct {
	// Later:
	// timer
	startTime time.Time
	// checkpoints
	// cash
	// energy
}

func NewHUDOverlay() *HUDOverlay {
	return &HUDOverlay{
		startTime: time.Now(),
	}
}

func (h *HUDOverlay) Update() error {
	// No logic needed here for a simple timer,
	// though you could "pause" the timer by storing a duration instead.
	return nil
}

// Calculate elapsed time since start of race
func elapsedTime(h *HUDOverlay) string {
	// Calculate elapsed time
	elapsed := time.Since(h.startTime)

	// Breakdown the duration
	// We use % 1000 to get just the 3-digit millisecond or 6-digit microsecond remainder
	h_val := int(elapsed.Hours())
	m_val := int(elapsed.Minutes()) % 60
	s_val := int(elapsed.Seconds()) % 60
	ms_val := elapsed.Microseconds() % 100 // Microseconds (6 digits)

	// Format: HH:MM:SS:µµ
	timeStr := fmt.Sprintf("TIME %02d:%02d:%02d:%02d", h_val, m_val, s_val, ms_val)
	return timeStr
}

func (h *HUDOverlay) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, elapsedTime(h), 10, 10)
	ebitenutil.DebugPrintAt(screen, "CHECKPOINTS 1/3", 130, 10)
	ebitenutil.DebugPrintAt(screen, "$ 123", 260, 10)

}
