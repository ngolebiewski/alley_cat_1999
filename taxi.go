package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ngolebiewski/alley_cat_1999/retrotrack"
)

// Taxi implements Entity
type Taxi struct {
	x, y          float64
	speed         float64 // Current movement speed
	baseSpeed     float64 // The original speed assigned at spawn
	targetSpeed   float64 // The speed the taxi wants to be at
	scale         float64
	frames        []*ebiten.Image
	frame         int
	frameTick     int
	dir           string
	laneX, laneY  float64
	width, height float64
	manager       *TaxiManager

	crashed       bool
	crashTime     float64
	recoveryTimer float64
	hasHonked     bool
}

// Particle struct for the crash effect
type Particle struct {
	x, y    float64
	dx, dy  float64
	life    float64
	maxLife float64
	angle   float64
	rotVel  float64
	alpha   float32
}

type ParticleSystem struct {
	particles []Particle
	tile      *ebiten.Image
	scale     float64
}

// ---- Entity interface ----

func (t *Taxi) Bounds() image.Rectangle {
	return image.Rect(
		int(t.x),
		int(t.y),
		int(t.x+t.width*t.scale),
		int(t.y+t.height*t.scale),
	)
}

func (t *Taxi) OnCollision(other Entity) {
	if t.crashed || t.recoveryTimer > 0 {
		return
	}

	_, isPlayer := other.(*Player)
	if isPlayer {
		t.SilentCrash()
	} else {
		if otherTaxi, ok := other.(*Taxi); ok {
			if otherTaxi.recoveryTimer > 0 {
				return
			}
		}
		t.Crash()
	}
}

// ---- Taxi logic ----

func NewTaxi(manager *TaxiManager, frames []*ebiten.Image, x, y, speed float64, dir string, scale float64) *Taxi {
	w := float64(frames[0].Bounds().Dx())
	h := float64(frames[0].Bounds().Dy())

	s := speed + rand.Float64()*0.5
	return &Taxi{
		manager:     manager,
		frames:      frames,
		x:           x,
		y:           y,
		speed:       s,
		baseSpeed:   s,
		targetSpeed: s,
		dir:         dir,
		scale:       scale,
		width:       w,
		height:      h,
		laneX:       x,
		laneY:       y,
	}
}

func (t *Taxi) canDriveAt(px, py float64) bool {
	if t.manager == nil || t.manager.roadLayer == nil || t.manager.spawnMap == nil {
		return false
	}
	m := t.manager.spawnMap
	gridSize := 16.0 * t.scale
	tileX := int(px / gridSize)
	tileY := int(py / gridSize)

	if tileX < 0 || tileY < 0 || tileX >= m.Width || tileY >= m.Height {
		return false
	}
	idx := tileY*m.Width + tileX
	if idx < 0 || idx >= len(t.manager.roadLayer.Data) {
		return false
	}
	return t.manager.roadLayer.Data[idx] == 2
}

func (t *Taxi) Update(playerX, playerY float64) {
	if t.crashed {
		t.crashTime -= 1.0 / 60.0
		t.frameTick++
		if t.crashTime > 1.0 && t.frameTick%15 == 0 {
			t.manager.particles.Spawn(t.x+(t.width*t.scale)/2, t.y, 1)
		}
		if t.crashTime <= 0 {
			t.crashed = false
			t.speed = 0 // Start from stop after waking up
			t.targetSpeed = t.baseSpeed
			t.recoveryTimer = 2.5
		}
		return
	}

	if t.recoveryTimer > 0 {
		t.recoveryTimer -= 1.0 / 60.0
	}

	// 1. Unified Sensing
	blockedByTaxi := t.manager.TaxiInFront(t)
	blockedByPlayer := t.isPlayerInFront(playerX, playerY)
	isBlocked := blockedByTaxi || blockedByPlayer

	// 2. Decision & Momentum Logic
	if isBlocked {
		t.targetSpeed = 0
		if blockedByPlayer && !t.hasHonked {
			retrotrack.PlayHonk()
			t.hasHonked = true
		}

		// Swerve if it's just a taxi blocking us
		if blockedByTaxi && !blockedByPlayer {
			nudge := 0.6
			if t.dir == "UP" || t.dir == "DOWN" {
				if t.canDriveAt(t.x+15*t.scale, t.y) {
					t.x += nudge
				} else if t.canDriveAt(t.x-15*t.scale, t.y) {
					t.x -= nudge
				}
			} else {
				if t.canDriveAt(t.x, t.y+15*t.scale) {
					t.y += nudge
				} else if t.canDriveAt(t.x, t.y-15*t.scale) {
					t.y -= nudge
				}
			}
		}
	} else {
		t.targetSpeed = t.baseSpeed
		t.hasHonked = false
	}

	// 3. Smooth Acceleration/Deceleration
	if t.speed > t.targetSpeed {
		t.speed -= 0.08 // Brake force
	} else if t.speed < t.targetSpeed {
		t.speed += 0.03 // Accel force
	}

	// Clamp speed
	if t.speed < 0.05 && t.targetSpeed == 0 {
		t.speed = 0
	}

	// 4. Movement
	switch t.dir {
	case "LEFT":
		t.x -= t.speed
	case "RIGHT":
		t.x += t.speed
	case "UP":
		t.y -= t.speed
	case "DOWN":
		t.y += t.speed
	}

	t.frameTick++
	if t.frameTick%8 == 0 {
		t.frame = (t.frame + 1) % len(t.frames)
	}

	if t.isOutOfBounds() {
		t.Respawn()
	}
}

func (t *Taxi) isPlayerInFront(px, py float64) bool {
	lookAhead := 55.0 * t.scale
	laneWidth := 14.0 * t.scale
	cx := t.x + (t.width*t.scale)/2
	cy := t.y + (t.height*t.scale)/2
	dx := px - cx
	dy := py - cy

	switch t.dir {
	case "RIGHT":
		return dx > 0 && dx < lookAhead && math.Abs(dy) < laneWidth
	case "LEFT":
		return dx < 0 && dx > -lookAhead && math.Abs(dy) < laneWidth
	case "UP":
		return dy < 0 && dy > -lookAhead && math.Abs(dx) < laneWidth
	case "DOWN":
		return dy > 0 && dy < lookAhead && math.Abs(dx) < laneWidth
	}
	return false
}

func (t *Taxi) isOutOfBounds() bool {
	limitW := t.manager.worldW
	limitH := t.manager.worldH
	buffer := 120.0 * t.scale

	switch t.dir {
	case "RIGHT":
		return t.x > limitW+buffer
	case "LEFT":
		return t.x < -buffer
	case "UP":
		return t.y < -buffer
	case "DOWN":
		return t.y > limitH+buffer
	}
	return false
}

func (t *Taxi) Draw(screen *ebiten.Image, cam *Camera) {
	op := &ebiten.DrawImageOptions{}

	if t.dir == "RIGHT" {
		op.GeoM.Scale(-t.scale, t.scale)
		op.GeoM.Translate(t.width*t.scale, 0)
	} else {
		op.GeoM.Scale(t.scale, t.scale)
	}

	if t.dir == "DOWN" {
		op.GeoM.Scale(1, -1)
		op.GeoM.Translate(0, t.height*t.scale)
	}

	op.GeoM.Translate(t.x-cam.X, t.y-cam.Y)

	if t.crashed {
		op.ColorScale.Scale(1, 0.4, 0.4, 1)
	} else if t.recoveryTimer > 0 {
		op.ColorScale.Scale(1, 1, 1, 0.5)
	} else if t.targetSpeed == 0 && t.speed > 0 {
		// Subtle brake light effect
		op.ColorScale.Scale(1.2, 0.8, 0.8, 1)
	}

	if isDebugMode {
		clr := color.RGBA{0, 255, 0, 255}
		if t.crashed {
			clr = color.RGBA{255, 0, 0, 255}
		} else if t.targetSpeed == 0 {
			clr = color.RGBA{255, 255, 0, 255}
		}

		vector.StrokeRect(screen, float32(t.x-cam.X), float32(t.y-cam.Y), float32(t.width*t.scale), float32(t.height*t.scale), 1, clr, false)
	}

	screen.DrawImage(t.frames[t.frame], op)
}

func (t *Taxi) Crash() {
	if t.crashed || t.recoveryTimer > 0 {
		return
	}
	t.crashed = true
	t.crashTime = 12.0 + rand.Float64()*5.0
	t.speed = 0
	t.manager.particles.Spawn(t.x+(t.width*t.scale)/2, t.y+(t.height*t.scale)/2, 12)
}

func (t *Taxi) SilentCrash() {
	if t.crashed || t.recoveryTimer > 0 {
		return
	}
	t.crashed = true
	t.crashTime = 1.2
	t.speed = 0
}

func (t *Taxi) Respawn() {
	t.crashed = false
	t.recoveryTimer = 2.0
	buffer := 64.0 * t.scale

	switch t.dir {
	case "RIGHT":
		t.x = -buffer
		t.y = t.laneY
	case "LEFT":
		t.x = t.manager.worldW + buffer
		t.y = t.laneY
	case "UP":
		t.y = t.manager.worldH + buffer
		t.x = t.laneX
	case "DOWN":
		t.y = -buffer
		t.x = t.laneX
	}

	t.x += (rand.Float64() - 0.5) * 15
	t.y += (rand.Float64() - 0.5) * 15

	newSpeed := 1.0 + rand.Float64()*1.5
	t.baseSpeed = newSpeed
	t.targetSpeed = newSpeed
	t.speed = newSpeed
}

// ---- Particle System Logic ----

func (ps *ParticleSystem) Spawn(x, y float64, count int) {
	for i := 0; i < count; i++ {
		ps.particles = append(ps.particles, Particle{
			x: x, y: y,
			dx: randFloat(-1.2, 1.2), dy: randFloat(-2.0, -0.4),
			life: 1.0, maxLife: 1.0,
			angle:  rand.Float64() * 2 * math.Pi,
			rotVel: randFloat(-0.08, 0.08),
			alpha:  0.7,
		})
	}
}

func (ps *ParticleSystem) Update() {
	alive := ps.particles[:0]
	for _, p := range ps.particles {
		p.x += p.dx
		p.y += p.dy
		p.dy -= 0.02
		p.angle += p.rotVel
		p.life -= 0.015
		if p.life > 0 {
			alive = append(alive, p)
		}
	}
	ps.particles = alive
}

func (ps *ParticleSystem) Draw(screen *ebiten.Image, cam *Camera) {
	for _, p := range ps.particles {
		op := &ebiten.DrawImageOptions{}
		alpha := float32(p.life)
		if p.life < 0.2 {
			alpha = float32(p.life / 0.2)
		}
		s := ps.scale * (0.4 + p.life*0.6)
		op.GeoM.Translate(-8, -8)
		op.GeoM.Rotate(p.angle)
		op.GeoM.Scale(s, s)
		op.GeoM.Translate(p.x-cam.X, p.y-cam.Y)
		op.ColorScale.ScaleAlpha(alpha * p.alpha)
		screen.DrawImage(ps.tile, op)
	}
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
