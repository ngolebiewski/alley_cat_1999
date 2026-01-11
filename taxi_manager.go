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
	spawnMap  *tiled.Map      // Pointer to the full map data
	roadLayer *tiled.Layer    // Cached reference to the specific road layer
}

// NewTaxiManager initializes taxis and sets world size
func NewTaxiManager(tileset *ebiten.Image, scale float64, worldW, worldH float64, spawnMap *tiled.Map) *TaxiManager {
	tm := &TaxiManager{
		scale:    scale,
		worldW:   worldW,
		worldH:   worldH,
		spawnMap: spawnMap,
	}

	tm.roadLayer = findLayerRecursive(spawnMap.Layers, "Roads and Sidewalks")

	if isDebugMode {
		if tm.roadLayer == nil {
			fmt.Println("CRITICAL ERROR: 'Roads and Sidewalks' layer not found!")
		} else {
			fmt.Printf("Success: Found road layer with %d tiles\n", len(tm.roadLayer.Data))
		}
	}

	tileSize := 16
	tilesetWidth := tileset.Bounds().Dx()
	tilesPerRow := tilesetWidth / tileSize

	// Setup Particle Sprite (Sprite 75)
	smokeX := (75 % tilesPerRow) * tileSize
	smokeY := (75 / tilesPerRow) * tileSize
	smokeImg := tileset.SubImage(image.Rect(smokeX, smokeY, smokeX+tileSize, smokeY+tileSize)).(*ebiten.Image)

	tm.particles = &ParticleSystem{
		tile:  smokeImg,
		scale: scale,
	}

	if isDebugMode {
		fmt.Printf("Tileset Width: %d | Tiles Per Row: %d\n", tileset.Bounds().Dx(), tileset.Bounds().Dx()/16)
	}
	// SIDE-VIEW (32x16)
	sideFrames := []*ebiten.Image{
		subImageRect(tileset, tileRectFromTwoHorizTiles(120, 121, tileSize, tilesetWidth)),
		subImageRect(tileset, tileRectFromTwoHorizTiles(124, 125, tileSize, tilesetWidth)),
		subImageRect(tileset, tileRectFromTwoHorizTiles(127, 128, tileSize, tilesetWidth)),
	}

	// TOP-DOWN (16x32) --> the second tile brought up a traffic light!
	// upFrames := []*ebiten.Image{
	// 	subImageRect(tileset, tileRectFromTwoVertTiles(22, 32, tileSize, tilesetWidth)),
	// }

	// TOP-DOWN (16x32)
	// This uses your 22 and 32 logic but handles the fact they aren't in the same column
	upFrames := []*ebiten.Image{
		createVerticalTaxi(tileset, 22, 32, tileSize, tilesetWidth),
	}

	// Spawn Taxis
	tiledSpawns := tiled.ExtractTaxiSpawns(spawnMap)
	for _, s := range tiledSpawns {
		var frames []*ebiten.Image
		if s.Direction == "UP" || s.Direction == "DOWN" {
			frames = upFrames
		} else {
			frames = sideFrames
		}

		baseSpeed := 1.0 + (math.Mod(s.X+s.Y, 0.5))

		tm.taxis = append(tm.taxis,
			NewTaxi(tm, frames, s.X*scale, s.Y*scale, baseSpeed, s.Direction, scale),
		)
	}

	return tm
}

// TaxiInFront checks if another taxi is blocking the path
func (tm *TaxiManager) TaxiInFront(t *Taxi) bool {
	// We increase the buffer for top-down taxis vs side-view taxis
	buffer := 20.0 * tm.scale
	laneTolerance := 8.0 * tm.scale

	for _, other := range tm.taxis {
		if other == t || other.crashed {
			continue
		}

		switch t.dir {
		case "RIGHT":
			if math.Abs(other.y-t.y) < laneTolerance && other.x > t.x && other.x-t.x < buffer {
				return true
			}
		case "LEFT":
			if math.Abs(other.y-t.y) < laneTolerance && other.x < t.x && t.x-other.x < buffer {
				return true
			}
		case "DOWN":
			if math.Abs(other.x-t.x) < laneTolerance && other.y > t.y && other.y-t.y < buffer {
				return true
			}
		case "UP":
			if math.Abs(other.x-t.x) < laneTolerance && other.y < t.y && t.y-other.y < buffer {
				return true
			}
		}
	}
	return false
}

func (tm *TaxiManager) Update(px, py float64) {
	for _, t := range tm.taxis {
		// Pass the raw player coordinates into the taxi
		// The taxi will now handle its own "is player in front" logic
		t.Update(px, py)
	}
	tm.particles.Update()
}

func (tm *TaxiManager) Draw(screen *ebiten.Image, cam *Camera) {
	for _, t := range tm.taxis {
		t.Draw(screen, cam)
	}
	tm.particles.Draw(screen, cam)
}

// ---- Helpers ----

func findLayerRecursive(layers []tiled.Layer, name string) *tiled.Layer {
	for i := range layers {
		if layers[i].Name == name && layers[i].Type == "tilelayer" && len(layers[i].Data) > 0 {
			return &layers[i]
		}
		if layers[i].Type == "group" && len(layers[i].Layers) > 0 {
			found := findLayerRecursive(layers[i].Layers, name)
			if found != nil {
				return found
			}
		}
	}
	return nil
}

func tileRectFromTwoHorizTiles(tile1, tile2, tileSize, tilesetWidth int) image.Rectangle {
	tilesPerRow := tilesetWidth / tileSize
	x1 := (tile1 % tilesPerRow) * tileSize
	y1 := (tile1 / tilesPerRow) * tileSize
	x2 := (tile2 % tilesPerRow) * tileSize
	return image.Rect(x1, y1, x2+tileSize, y1+tileSize)
}

func subImageRect(tileset *ebiten.Image, rect image.Rectangle) *ebiten.Image {
	return tileset.SubImage(rect).(*ebiten.Image)
}

func createVerticalTaxi(tileset *ebiten.Image, topID, botID, tileSize, tilesetWidth int) *ebiten.Image {
	tilesPerRow := tilesetWidth / tileSize

	// Calculate Top Tile Source
	tx := (topID % tilesPerRow) * tileSize
	ty := (topID / tilesPerRow) * tileSize
	topImg := tileset.SubImage(image.Rect(tx, ty, tx+tileSize, ty+tileSize)).(*ebiten.Image)

	// Calculate Bottom Tile Source
	bx := (botID % tilesPerRow) * tileSize
	by := (botID / tilesPerRow) * tileSize
	botImg := tileset.SubImage(image.Rect(bx, by, bx+tileSize, by+tileSize)).(*ebiten.Image)

	// Create a new blank 16x32 image
	result := ebiten.NewImage(tileSize, tileSize*2)

	// Draw Top
	result.DrawImage(topImg, nil)

	// Draw Bottom (shifted down 16px)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, float64(tileSize))
	result.DrawImage(botImg, op)

	return result
}
