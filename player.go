package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type BikerState int

const (
	StateRiding BikerState = iota
	StateWalking
)

type Player struct {
	x, y       float64
	w, h       float64
	velX, velY float64
	state      BikerState
	dir        int // 0: Down, 1: Up, 2: Left, 3: Right
	frame      int
	frameTick  int
	img        *ebiten.Image

	// Gameplay stats
	health      int
	invulFrames int
	energy      int
}

func NewPlayer(img *ebiten.Image, startX, startY float64, width, height float64) *Player {
	return &Player{
		img:    img,
		x:      startX,
		y:      startY,
		state:  StateRiding,
		dir:    3, // Facing Right
		w:      width,
		h:      height,
		health: 100,
		energy: 100,
	}
}

// UpdateInput calculates desired velocity and handles state.
// The actual X/Y movement is handled by the RaceScene's collision grid.
func (p *Player) UpdateInput(inputX, inputY float64, toggleAxis, toggleMount bool) {
	const accel = 0.2
	const friction = 0.92
	const walkSpeed = 1.2

	// Handle invulnerability timer
	if p.invulFrames > 0 {
		p.invulFrames--
	}

	// 1. Handle Mounting/Dismounting
	if toggleMount {
		speed := math.Sqrt(p.velX*p.velX + p.velY*p.velY)
		if speed < 0.8 {
			if p.state == StateRiding {
				p.state = StateWalking
			} else {
				p.state = StateRiding
			}
		}
	}

	// 2. Handle Axis/Direction Logic
	if toggleAxis && p.state == StateRiding {
		dirOrder := []int{3, 0, 2, 1}
		for i, d := range dirOrder {
			if d == p.dir {
				p.dir = dirOrder[(i+1)%len(dirOrder)]
				break
			}
		}
	}

	// 3. Set Direction
	if p.state == StateRiding {
		if p.dir == 2 || p.dir == 3 {
			if inputX < 0 {
				p.dir = 2
			}
			if inputX > 0 {
				p.dir = 3
			}
		} else {
			if inputY < 0 {
				p.dir = 1
			}
			if inputY > 0 {
				p.dir = 0
			}
		}
	} else {
		if inputX < 0 {
			p.dir = 2
		}
		if inputX > 0 {
			p.dir = 3
		}
	}

	// 4. Calculate Velocity (Do not add to X/Y here; RaceScene does that)
	moving := (inputX != 0 || inputY != 0)
	if p.state == StateRiding {
		p.velX += inputX * accel
		p.velY += inputY * accel
		p.velX *= friction
		p.velY *= friction
	} else {
		if moving {
			p.velX = inputX * walkSpeed
			p.velY = inputY * walkSpeed
		} else {
			p.velX = 0
			p.velY = 0
		}
	}

	p.updateAnimation(moving)
}

// --- Entity Interface & Collision ---

func (p *Player) Bounds() image.Rectangle {
	return image.Rect(
		int(p.x),
		int(p.y),
		int(p.x+p.w),
		int(p.y+p.h),
	)
}

func (p *Player) OnCollision(other Entity) {
	switch e := other.(type) {
	case *Taxi:
		if p.invulFrames > 0 {
			return
		}

		// 1. Kickback Physics: Reverse and boost velocity
		p.velX = -p.velX * 1.5
		p.velY = -p.velY * 1.5

		// 2. POSITION EJECTION (The Fix for Phasing)
		// We move the player slightly away from the taxi center immediately
		// so they aren't overlapping in the next frame's move check.
		tx, ty := e.x+(e.width*e.scale)/2, e.y+(e.height*e.scale)/2
		px, py := p.Center()

		if px < tx {
			p.x -= 4
		} else {
			p.x += 4
		}
		if py < ty {
			p.y -= 4
		} else {
			p.y += 4
		}

		// 3. Damage & Invulnerability
		p.health -= 15
		p.invulFrames = 45

		if isDebugMode {
			fmt.Printf("HIT! Health: %d | Pos Ejected\n", p.health)
		}
	}
}

// --- Animation & Drawing ---

func (p *Player) updateAnimation(moving bool) {
	p.frameTick++
	if p.state == StateRiding {
		if moving {
			if p.frameTick%8 == 0 {
				switch p.dir {
				case 3, 2: // Side
					p.frame = (p.frame + 1) % 3
				case 1, 0: // Vertical
					if p.velY < 0 {
						if p.frame < 4 || p.frame > 5 {
							p.frame = 4
						}
						p.frame = 4 + (p.frameTick/8)%2
					} else {
						if p.frame < 6 || p.frame > 7 {
							p.frame = 6
						}
						p.frame = 6 + (p.frameTick/8)%2
					}
				}
			}
		} else {
			if p.dir == 2 || p.dir == 3 {
				p.frame = 3
			} else {
				p.frame = 14
			}
		}
	} else {
		if moving {
			if p.frameTick%10 == 0 {
				if p.frame < 9 || p.frame > 12 {
					p.frame = 9
				}
				p.frame = 9 + (p.frameTick/10)%4
			}
		} else {
			p.frame = 8 // Walk Idle
		}
	}
}

func (p *Player) DrawWithCamera(screen *ebiten.Image, cam *Camera) {
	// Simple flash effect when invulnerable
	if p.invulFrames > 0 && (p.frameTick/4)%2 == 0 {
		return
	}

	const size = 32
	op := &ebiten.DrawImageOptions{}

	if p.dir == 2 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(size, 0)
	}

	op.GeoM.Translate(p.x-cam.X, p.y-cam.Y)
	sx := p.frame * size
	screen.DrawImage(p.img.SubImage(image.Rect(sx, 0, sx+size, size)).(*ebiten.Image), op)

	// --- DEBUG HITBOX ---
	if isDebugMode {
		// Cyan outline for player
		vector.StrokeRect(screen,
			float32(p.x-cam.X),
			float32(p.y-cam.Y),
			float32(p.w),
			float32(p.h),
			1, // Stroke width
			color.RGBA{0, 255, 255, 255},
			false, // don't use anti-alias for hitboxes (sharper)
		)
	}
}

// --- Helpers ---

func (p *Player) Center() (float64, float64) {
	return p.x + p.w/2, p.y + p.h/2
}

func (p *Player) Move(dx, dy float64) {
	p.x += dx
	p.y += dy
}
