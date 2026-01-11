package main

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type HUDOverlay struct {
	startTime   time.Time
	health      float32 // 0.0 to 1.0 (synced from Player.health / 100)
	cash        int
	checkpoints int
	maxCheck    int
}

func NewHUDOverlay() *HUDOverlay {
	return &HUDOverlay{
		startTime: time.Now(),
		health:    1.0,
		maxCheck:  3,
	}
}

func (h *HUDOverlay) Update(paused bool) {
	// Logic-wise, we don't need to do anything for time here anymore
	// since time.Since() handles the heavy lifting in Draw.
}

func (h *HUDOverlay) Draw(screen *ebiten.Image) {
	// 1. Top Bar Alpha Plate (Heavier alpha for contrast)
	barHeight := float32(25 * zoom)
	vector.DrawFilledRect(screen, 0, 0, float32(screen.Bounds().Dx()), barHeight, color.RGBA{0, 0, 0, 220}, false)

	// 2. Draw 4 Big Hearts (3x Scale)
	maxHearts := 4
	activeHearts := int(math.Ceil(float64(h.health * float32(maxHearts))))

	for i := 0; i < maxHearts; i++ {
		// Spacing adjusted so they are close but chunky
		xPos := float32((10 + (i * 18)) * zoom)
		yPos := float32(8 * zoom)

		if i < activeHearts {
			h.drawBigHeart(screen, xPos, yPos, color.RGBA{255, 0, 0, 255})
		} else {
			// Dark hollow hearts
			h.drawBigHeart(screen, xPos, yPos, color.RGBA{40, 40, 40, 200})
		}
	}

	// 3. Stats Text
	timeStr := h.elapsedTimeStr()
	ebitenutil.DebugPrintAt(screen, timeStr, 110*zoom, 5*zoom)

	chkStr := fmt.Sprintf("CHK %d/%d", h.checkpoints, h.maxCheck)
	ebitenutil.DebugPrintAt(screen, chkStr, 210*zoom, 5*zoom)

	cashStr := fmt.Sprintf("$ %d", h.cash)
	ebitenutil.DebugPrintAt(screen, cashStr, 270*zoom, 5*zoom)

	// 4. Hospital State Wash
	if h.health <= 0 {
		vector.DrawFilledRect(screen, 0, 0, float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()), color.RGBA{255, 255, 255, 100}, false)
		ebitenutil.DebugPrintAt(screen, "HOSPITALIZED - CLICK TO RECOVER", 100*zoom, 120*zoom)
	}
}

// drawBigHeart creates a 3x magnified pixel heart
func (h *HUDOverlay) drawBigHeart(screen *ebiten.Image, x, y float32, clr color.Color) {
	s := float32(3 * zoom) // Individual pixel size

	// Row 1:  . X . X .
	vector.DrawFilledRect(screen, x+s, y, s, s, clr, false)
	vector.DrawFilledRect(screen, x+3*s, y, s, s, clr, false)
	// Row 2:  X X X X X
	vector.DrawFilledRect(screen, x, y+s, 5*s, s, clr, false)
	// Row 3:  . X X X .
	vector.DrawFilledRect(screen, x+s, y+2*s, 3*s, s, clr, false)
	// Row 4:  . . X . .
	vector.DrawFilledRect(screen, x+2*s, y+3*s, s, s, clr, false)
}

// elapsedTimeStr uses your preferred time.Since logic
func (h *HUDOverlay) elapsedTimeStr() string {
	elapsed := time.Since(h.startTime)
	h_val := int(elapsed.Hours())
	m_val := int(elapsed.Minutes()) % 60
	s_val := int(elapsed.Seconds()) % 60

	// Format: HH:MM:SS
	return fmt.Sprintf("TIME %02d:%02d:%02d", h_val, m_val, s_val)
}

func (h *HUDOverlay) Reset() {
	h.startTime = time.Now()
	h.health = 1.0
}

func (s *RaceScene) drawPauseOverlay(screen *ebiten.Image) {
	// 1. Draw a dark semi-transparent overlay
	vector.FillRect(screen, 0, 0, float32(screenWidth), float32(screenHeight), color.RGBA{0, 0, 0, 180}, false)

	// 2. Draw the list
	yOff := 50
	ebitenutil.DebugPrintAt(screen, "--- MISSION MANIFEST (PAUSED) ---", 40, yOff)
	yOff += 30

	if s.manifest != nil {
		for i, cp := range s.manifest.Checkpoints {
			status := "[ ] "
			if cp.IsComplete {
				status = "[x] "
			}

			line := status + cp.Name
			if cp.IsFinishLine {
				line = "[FINISH] " + cp.Name
			}

			ebitenutil.DebugPrintAt(screen, line, 50, yOff+(i*15))
		}
	}

	ebitenutil.DebugPrintAt(screen, "\nPRESS ENTER TO RESUME", 40, screenHeight-40)
}
