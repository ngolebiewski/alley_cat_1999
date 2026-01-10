package main

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ngolebiewski/alley_cat_1999/tiled"
)

// TaxiSpawn represents a spawn point for a taxi
type TaxiSpawn struct {
	X, Y      float64
	Direction string
}

// TaxiManager manages all taxis and particle effects
type TaxiManager struct {
	taxis     []*Taxi
	scale     float64
	worldW    float64
	worldH    float64
	particles *ParticleSystem // for crash effects
}

// NewTaxiManager initializes taxis and sets world size
func NewTaxiManager(tileset *ebiten.Image, scale float64, worldW, worldH float64, spawnMap *tiled.Map) *TaxiManager {
	tm := &TaxiManager{
		scale:  scale,
		worldW: worldW,
		worldH: worldH,
	}

	tileSize := 16
	tilesetWidth := tileset.Bounds().Dx()

	// SIDE-VIEW TAXIS (32x16) → 3 frames for wheel animation
	sideFrames := []*ebiten.Image{
		subImageRect(tileset, tileRectFromTwoHorizTiles(120, 121, tileSize, tilesetWidth)),
		subImageRect(tileset, tileRectFromTwoHorizTiles(124, 125, tileSize, tilesetWidth)),
		subImageRect(tileset, tileRectFromTwoHorizTiles(127, 128, tileSize, tilesetWidth)),
	}

	// TOP-DOWN TAXIS (16x32) → single frame
	upFrames := []*ebiten.Image{
		subImageRect(tileset, tileRectFromTwoVertTiles(22, 32, tileSize, tilesetWidth)),
	}

	// Initialize stub particle system
	tm.particles = &ParticleSystem{}

	// Extract taxi spawns from the Tiled map
	tiledSpawns := tiled.ExtractTaxiSpawns(spawnMap)
	fmt.Println("Taxi spawns:", tiledSpawns)

	// Create taxis at spawn points
	for _, s := range tiledSpawns {
		var frames []*ebiten.Image
		if s.Direction == "UP" {
			frames = upFrames
		} else {
			frames = sideFrames
		}

		tm.taxis = append(tm.taxis,
			NewTaxi(tm, frames, s.X*scale, s.Y*scale, 1.0, s.Direction, scale),
		)
	}

	return tm
}

// TaxiInFront checks if another taxi is blocking the given taxi in its moving direction
func (tm *TaxiManager) TaxiInFront(t *Taxi) bool {
	const buffer = 2.0
	const laneTolerance = 1.0 // allow minor positional difference

	for _, other := range tm.taxis {
		if other == t {
			continue
		}

		switch t.dir {
		case "RIGHT":
			if math.Abs(other.y-t.y) < laneTolerance && other.x > t.x && other.x-t.x < t.speed+buffer {
				return true
			}
		case "LEFT":
			if math.Abs(other.y-t.y) < laneTolerance && other.x < t.x && t.x-other.x < t.speed+buffer {
				return true
			}
		case "DOWN":
			if math.Abs(other.x-t.x) < laneTolerance && other.y > t.y && other.y-t.y < t.speed+buffer {
				return true
			}
		case "UP":
			if math.Abs(other.x-t.x) < laneTolerance && other.y < t.y && t.y-other.y < t.speed+buffer {
				return true
			}
		}
	}

	return false
}

// ---- helpers ----

// Combine two horizontal tiles into one 32x16 rect
func tileRectFromTwoHorizTiles(tile1, tile2, tileSize, tilesetWidth int) image.Rectangle {
	tilesPerRow := tilesetWidth / tileSize
	x1 := (tile1 % tilesPerRow) * tileSize
	y1 := (tile1 / tilesPerRow) * tileSize

	x2 := (tile2 % tilesPerRow) * tileSize

	return image.Rect(x1, y1, x2+tileSize, y1+tileSize)
}

// Combine two vertical tiles into one 16x32 rect
func tileRectFromTwoVertTiles(tile1, tile2, tileSize, tilesetWidth int) image.Rectangle {
	tilesPerRow := tilesetWidth / tileSize
	x1 := (tile1 % tilesPerRow) * tileSize
	y1 := (tile1 / tilesPerRow) * tileSize

	y2 := (tile2 / tilesPerRow) * tileSize

	return image.Rect(x1, y1, x1+tileSize, y2+tileSize)
}

// subimage helper from a rectangle
func subImageRect(tileset *ebiten.Image, rect image.Rectangle) *ebiten.Image {
	return tileset.SubImage(rect).(*ebiten.Image)
}

// Update all taxis and particles (pass player for avoidance/collisions)
func (tm *TaxiManager) Update(player *Player) {
	for _, t := range tm.taxis {
		t.Update(player)
	}
	tm.particles.Update()
}

// Draw all taxis and particles
func (tm *TaxiManager) Draw(screen *ebiten.Image, cam *Camera) {
	for _, t := range tm.taxis {
		t.Draw(screen, cam)
	}
	tm.particles.Draw(screen, screen) // stub, just pass screen twice for now
}
