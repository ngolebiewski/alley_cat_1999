package main

import (
	"image"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// Taxi implements Entity
type Taxi struct {
	x, y          float64
	speed         float64
	scale         float64
	frames        []*ebiten.Image
	frame         int
	frameTick     int
	dir           string
	laneX, laneY  float64
	width, height float64
	manager       *TaxiManager

	crashed   bool
	crashTime float64
}

// Particle struct for crash effect
type Particle struct {
	x, y   float64
	dx, dy float64
	life   float64
}

// Particle system in TaxiManager
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
	t.Crash()
}

// ---- Taxi constructor ----

func NewTaxi(manager *TaxiManager, frames []*ebiten.Image, x, y, speed float64, dir string, scale float64) *Taxi {
	w := float64(frames[0].Bounds().Dx())
	h := float64(frames[0].Bounds().Dy())
	return &Taxi{
		manager: manager,
		frames:  frames,
		x:       x,
		y:       y,
		speed:   speed,
		dir:     dir,
		scale:   scale,
		width:   w,
		height:  h,
		laneX:   x,
		laneY:   y,
	}
}

// ---- Update ----

func (t *Taxi) Update(player *Player) {
	const wiggle = 0.3
	const frameSpeed = 8

	if t.crashed {
		t.crashTime -= 1.0 / 60.0
		if t.crashTime <= 0 {
			t.crashed = false
			t.speed = 1.0
		}
		return
	}

	// Wiggle perpendicular
	if t.dir == "LEFT" || t.dir == "RIGHT" {
		t.y += randFloat(-wiggle, wiggle)
	} else {
		t.x += randFloat(-wiggle, wiggle)
	}

	// Calculate next position
	nextX, nextY := t.x, t.y
	switch t.dir {
	case "LEFT":
		nextX -= t.speed
	case "RIGHT":
		nextX += t.speed
	case "UP":
		nextY -= t.speed
	case "DOWN":
		nextY += t.speed
	}

	// Check player collision
	if intersects(nextX, nextY, t.width*t.scale, t.height*t.scale,
		player.x, player.y, player.w, player.h) {
		t.Crash()
		// TODO: reduce player health & flash red
		nextX, nextY = t.x, t.y
	}

	// Check collisions with other taxis
	for _, other := range t.manager.taxis {
		if other == t {
			continue
		}
		if intersects(nextX, nextY, t.width*t.scale, t.height*t.scale,
			other.x, other.y, other.width, other.height) {
			t.Crash()
			other.Crash()
			nextX, nextY = t.x, t.y
		}
	}

	t.x = nextX
	t.y = nextY

	// Animate frames
	t.frameTick++
	if t.frameTick%frameSpeed == 0 {
		t.frame = (t.frame + 1) % len(t.frames)
	}

	// Check world bounds
	tWidth := t.width * t.scale
	tHeight := t.height * t.scale
	out := false
	switch t.dir {
	case "LEFT":
		if t.x+tWidth < 0 {
			out = true
		}
	case "RIGHT":
		if t.x > t.manager.worldW {
			out = true
		}
	case "UP":
		if t.y+tHeight < 0 {
			out = true
		}
	case "DOWN":
		if t.y > t.manager.worldH {
			out = true
		}
	}
	if out {
		t.Respawn()
	}
}

// ---- Draw ----

func (t *Taxi) Draw(screen *ebiten.Image, cam *Camera) {
	op := &ebiten.DrawImageOptions{}

	switch t.dir {
	case "RIGHT":
		op.GeoM.Scale(-t.scale, t.scale)
		op.GeoM.Translate(-t.width*t.scale, 0)
	case "DOWN":
		op.GeoM.Scale(t.scale, -t.scale)
		op.GeoM.Translate(0, -t.height*t.scale)
	default:
		op.GeoM.Scale(t.scale, t.scale)
	}

	// camera offset
	op.GeoM.Translate(-cam.X, -cam.Y)
	op.GeoM.Translate(t.x, t.y)

	// draw frame
	screen.DrawImage(t.frames[t.frame], op)
}

// ---- Respawn ----

func (t *Taxi) Respawn() {
	t.frame = 0
	t.frameTick = 0

	switch t.dir {
	case "LEFT":
		t.x = t.manager.worldW
	case "RIGHT":
		t.x = -t.width * t.scale
	case "UP":
		t.y = t.manager.worldH
	case "DOWN":
		t.y = -t.height * t.scale
	}

	if t.dir == "LEFT" || t.dir == "RIGHT" {
		t.y = t.laneY
	} else {
		t.x = t.laneX
	}
}

// ---- Crash ----

func (t *Taxi) Crash() {
	if t.crashed {
		return
	}
	t.crashed = true
	t.crashTime = 2.0
	t.speed = 0
	// spawn particles
	for i := 0; i < 20; i++ {
		t.manager.particles.Spawn(t.x+t.width*t.scale/2, t.y+t.height*t.scale/2)
	}
}

// ---- Particle helpers ----

func (ps *ParticleSystem) Spawn(x, y float64) {
	ps.particles = append(ps.particles, Particle{
		x:    x,
		y:    y,
		dx:   randFloat(-1.5, 1.5),
		dy:   randFloat(-4, -1),
		life: randFloat(0.5, 1.2),
	})
}

func (ps *ParticleSystem) Update() {
	alive := ps.particles[:0]
	for _, p := range ps.particles {
		p.x += p.dx
		p.y += p.dy
		p.dy += 0.1 // gravity
		p.life -= 1.0 / 60.0
		if p.life > 0 {
			alive = append(alive, p)
		}
	}
	ps.particles = alive
}

func (ps *ParticleSystem) Draw(screen *ebiten.Image, tile *ebiten.Image) {
	for _, p := range ps.particles {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.x, p.y)
		op.GeoM.Scale(ps.scale, ps.scale)
		screen.DrawImage(tile, op)
	}
}

// ---- Helpers ----

func intersects(x, y, w, h, ox, oy, ow, oh float64) bool {
	return x < ox+ow && x+w > ox && y < oy+oh && y+h > oy
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
