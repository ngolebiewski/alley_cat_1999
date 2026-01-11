package main

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
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
}

func NewPlayer(img *ebiten.Image, startX, startY float64, width, height float64) *Player {
	return &Player{
		img:   img,
		x:     startX,
		y:     startY,
		state: StateRiding,
		dir:   3, // Facing Right
		w:     width,
		h:     height,
	}
}

// --- Input & Update ---

func (p *Player) Update(inputX, inputY float64, toggleAxis, toggleMount bool) {
	const accel = 0.2
	const friction = 0.92
	const walkSpeed = 1.2

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
		dirOrder := []int{3, 0, 2, 1} // Right -> Down -> Left -> Up
		for i, d := range dirOrder {
			if d == p.dir {
				p.dir = dirOrder[(i+1)%len(dirOrder)]
				break
			}
		}
	}

	// 3. Auto-flip facing direction
	if p.state == StateRiding {
		if p.dir == 2 || p.dir == 3 { // Horizontal mode
			if inputX < 0 {
				p.dir = 2
			}
			if inputX > 0 {
				p.dir = 3
			}
		} else { // Vertical mode
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

	// 4. Physics
	moving := (inputX != 0 || inputY != 0)
	if p.state == StateRiding {
		p.velX += inputX * accel
		p.velY += inputY * accel
		p.velX *= friction
		p.velY *= friction
		p.x += p.velX
		p.y += p.velY
	} else {
		if moving {
			p.x += inputX * walkSpeed
			p.y += inputY * walkSpeed
		}
	}

	p.updateAnimation(moving)
}

func (p *Player) UpdateInput(inputX, inputY float64, toggleAxis, toggleMount bool) {
	const accel = 0.2
	const friction = 0.92
	const walkSpeed = 1.2

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

	if toggleAxis && p.state == StateRiding {
		dirOrder := []int{3, 0, 2, 1}
		for i, d := range dirOrder {
			if d == p.dir {
				p.dir = dirOrder[(i+1)%len(dirOrder)]
				break
			}
		}
	}

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

	// Physics (compute velocity, do not move yet)
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

// separate movement, used by scene for collision handling
func (p *Player) Move(dx, dy float64) {
	p.x += dx
	p.y += dy
}

// --- Animation ---

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

// --- Drawing ---

func (p *Player) Draw(screen *ebiten.Image) {
	const size = 32
	op := &ebiten.DrawImageOptions{}

	if p.dir == 2 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(size, 0)
	}

	op.GeoM.Translate(p.x, p.y)
	sx := p.frame * size
	screen.DrawImage(p.img.SubImage(image.Rect(sx, 0, sx+size, size)).(*ebiten.Image), op)
}

func (p *Player) DrawWithCamera(screen *ebiten.Image, cam *Camera) {
	const size = 32
	op := &ebiten.DrawImageOptions{}

	if p.dir == 2 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(size, 0)
	}

	op.GeoM.Translate(p.x-cam.X, p.y-cam.Y)
	sx := p.frame * size
	screen.DrawImage(p.img.SubImage(image.Rect(sx, 0, sx+size, size)).(*ebiten.Image), op)
}

// --- Utility ---

func (p *Player) Center() (float64, float64) {
	return p.x + float64(p.w/2), p.y + float64(p.h/2)
}

// --- Entity interface implementation ---

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
		// Stop player on collision
		p.velX = 0
		p.velY = 0

		// TODO: Flash red, reduce health
		fmt.Println("Player collided with taxi at", e.x, e.y)
	default:
		// handle other entity collisions if needed
	}
}
