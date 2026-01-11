package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ngolebiewski/alley_cat_1999/retrotrack"
	"github.com/ngolebiewski/alley_cat_1999/tiled"
)

type BikerState int

const (
	StateRiding BikerState = iota
	StateWalking
	StateHospital
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
	cash        int
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
		cash:   100,
	}
}

// UpdateInput calculates desired velocity and handles state.
func (p *Player) UpdateInput(inputX, inputY float64, toggleAxis, toggleMount bool) {
	// If in Hospital, freeze input and movement
	if p.state == StateHospital {
		p.velX = 0
		p.velY = 0
		return
	}

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

	// 4. Calculate Velocity
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

func (p *Player) wouldCollideAt(newX, newY float64, grid *tiled.CollisionGrid) bool {
	if grid == nil {
		return false
	}
	const tileSize = 32

	x1 := int(newX) / tileSize
	y1 := int(newY) / tileSize
	x2 := int(newX+p.w-1) / tileSize
	y2 := int(newY+p.h-1) / tileSize

	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if y < 0 || y >= grid.Height || x < 0 || x >= grid.Width {
				continue
			}
			if grid.Solid[y][x] {
				return true
			}
		}
	}
	return false
}

func (p *Player) OnCollision(other Entity, grid *tiled.CollisionGrid) { // Added grid parameter
	if p.state == StateHospital {
		return
	}

	switch e := other.(type) {
	case *Taxi:
		if p.invulFrames > 0 {
			return
		}

		retrotrack.PlayCrash()

		// 1. Kickback Physics
		p.velX = -p.velX * 1.5
		p.velY = -p.velY * 1.5

		// 2. SAFE Position Ejection
		tx, ty := e.x+(e.width*e.scale)/2, e.y+(e.height*e.scale)/2
		px, py := p.Center()

		ejectAmt := 12.0 // Slightly larger to ensure they clear the taxi

		dirX := 1.0
		if px < tx {
			dirX = -1.0
		}
		dirY := 1.0
		if py < ty {
			dirY = -1.0
		}

		// Attempt to eject X
		if !p.wouldCollideAt(p.x+(dirX*ejectAmt), p.y, grid) {
			p.x += (dirX * ejectAmt)
		}
		// Attempt to eject Y
		if !p.wouldCollideAt(p.x, p.y+(dirY*ejectAmt), grid) {
			p.y += (dirY * ejectAmt)
		}

		// 3. Damage & Invulnerability
		p.health -= 25
		if p.health <= 0 {
			p.health = 0
			p.state = StateHospital
			retrotrack.Stop() // Kill music on hospitalization
		} else {
			p.invulFrames = 45
		}

		if isDebugMode {
			fmt.Printf("HIT! Health: %d\n", p.health)
		}
	}
}

// --- Animation & Drawing ---

func (p *Player) updateAnimation(moving bool) {
	p.frameTick++
	if p.state == StateHospital {
		p.frame = 14 // Assuming frame 14 is a 'knocked out' sprite
		return
	}

	if p.state == StateRiding {
		if moving {
			if p.frameTick%8 == 0 {
				switch p.dir {
				case 3, 2:
					p.frame = (p.frame + 1) % 3
				case 1, 0:
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

	// If hospitalized, rotate the player 90 degrees to look like they are lying down
	if p.state == StateHospital {
		op.GeoM.Rotate(math.Pi / 2)
		op.GeoM.Translate(size, 0)
	}

	// Handle flipping for Left direction
	if p.dir == 2 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(size, 0)
	}

	// 1. Move to world position relative to camera
	op.GeoM.Translate(p.x-cam.X, p.y-cam.Y)

	// 2. APPLY GLOBAL ZOOM
	op.GeoM.Scale(float64(zoom), float64(zoom))

	sx := p.frame * size
	sub := p.img.SubImage(image.Rect(sx, 0, sx+size, size)).(*ebiten.Image)
	screen.DrawImage(sub, op)

	// --- DEBUG HITBOX ---
	if isDebugMode {
		vector.StrokeRect(screen,
			float32((p.x-cam.X)*float64(zoom)),
			float32((p.y-cam.Y)*float64(zoom)),
			float32(p.w*float64(zoom)),
			float32(p.h*float64(zoom)),
			1,
			color.RGBA{0, 255, 255, 255},
			false,
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
