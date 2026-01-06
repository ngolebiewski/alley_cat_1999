package main

import (
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
	velX, velY float64
	state      BikerState
	dir        int // 0: Down, 1: Up, 2: Left, 3: Right
	frame      int
	frameTick  int
	img        *ebiten.Image
}

func NewPlayer(img *ebiten.Image, startX, startY float64) *Player {
	return &Player{
		img:   img,
		x:     startX,
		y:     startY,
		state: StateRiding,
		dir:   3, // Facing Right
	}
}

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
		// WALKING AUTO-FLIP: Flip based on horizontal movement
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

func (p *Player) Draw(screen *ebiten.Image) {
	const size = 32
	op := &ebiten.DrawImageOptions{}

	// Draw() uses the current p.dir to decide whether to flip.
	// Since we update p.dir in walking state above, this works perfectly.
	if p.dir == 2 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(size, 0)
	}

	op.GeoM.Translate(p.x, p.y)
	sx := p.frame * size
	screen.DrawImage(p.img.SubImage(image.Rect(sx, 0, sx+size, size)).(*ebiten.Image), op)
}
