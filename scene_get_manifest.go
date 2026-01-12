package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ngolebiewski/alley_cat_1999/retrotrack"
	"github.com/ngolebiewski/alley_cat_1999/tiled"
)

const tileSize = 16

var darkBox = color.RGBA{0, 0, 0, 204}

type GetManifestScene struct {
	game           *Game
	tileset        *ebiten.Image
	manifestImg    *ebiten.Image
	touchIDs       []ebiten.TouchID
	animTime       float64 // seconds elapsed
	animDone       bool    // stop animating after 1 second
	activeManifest *Manifest
}

func NewGetManifestScene(game *Game) *GetManifestScene {
	tileset := game.assets.TilesetImage

	// 1. Load map just to get checkpoint data
	// We do this here so we can show the names on screen before the race starts

	// m, err := tiled.LoadMapFS(embeddedAssets, "assets/nyc_1_TEST..tmj")
	m, err := tiled.LoadMapFS(embeddedAssets, "assets/nyc_1..tmj")
	if err != nil {
		fmt.Printf("DEBUG ERROR: Could not load map: %v\n", err)
		panic(err)
	}

	// 2. Create the data object that will persist into the race
	// Scale is 2.0 because our RaceScene scales the map by 2
	manifestData := NewManifest(m, game.assets.PeopleImage, 2.0)
	fmt.Printf("DEBUG: Manifest logic complete. %d stops planned.\n", len(manifestData.Checkpoints))

	return &GetManifestScene{
		game:           game,
		tileset:        tileset,
		manifestImg:    buildManifestImage(tileset, tileSize),
		activeManifest: manifestData,
	}
}

func (s *GetManifestScene) Update() error {
	// Increment animation time if not done
	if !s.animDone {
		s.animTime += 1.0 / 60.0
		if s.animTime >= 1.0 {
			s.animTime = 1.0
			s.animDone = true
			retrotrack.PlayManifestSound()
		}
	}

	// Check for input to switch to the actual Race
	startPressed := inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0)

	s.touchIDs = inpututil.AppendJustPressedTouchIDs(s.touchIDs[:0])
	if len(s.touchIDs) > 0 {
		startPressed = true
	}

	if startPressed {
		fmt.Println("DEBUG: Switching to RaceScene. Passing manifest data...")
		retrotrack.Start()
		// We pass the manifest we generated so the RaceScene doesn't have to reload it
		s.game.scene = NewRaceScene(s.game, s.activeManifest)
	}

	return nil
}

func (s *GetManifestScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// 1. Draw Text Instructions
	ebitenutil.DebugPrint(
		screen,
		"--- MANIFEST RECEIVED ---\nCheck-in at all points on the list!\nWatch out for Taxis!",
	)

	// 3. Animation Logic for the spinning Manifest Sprite
	t := s.animTime / 1.0
	if t > 1 {
		t = 1
	}
	ease := t * (2 - t) // Ease Out

	scale := 5.0 * ease
	rotation := (1 - ease) * 2 * 3.1415926

	w := s.manifestImg.Bounds().Dx()
	h := s.manifestImg.Bounds().Dy()

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest

	// Center, rotate, scale, and move to middle of screen
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Rotate(rotation)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(screenWidth)/2, float64(screenHeight)/2)

	screen.DrawImage(s.manifestImg, op)

	// 2. Draw the list of locations from our actual data

	if s.animDone {
		vector.FillRect(
			screen,
			0,
			0,
			float32(screenWidth),
			float32(screenHeight),
			darkBox,
			false, // antialias (doesn't matter for axis-aligned rects)
		)
		yOff := 60 * zoom
		for i, cp := range s.activeManifest.Checkpoints {
			prefix := "[ ] "
			if cp.IsFinishLine {
				prefix = "[FINISH] "
			}
			ebitenutil.DebugPrintAt(screen, prefix+cp.Name, 40*zoom, yOff+(i*15))
		}
		ebitenutil.DebugPrintAt(screen, "PRESS SPACE TO START RACE", 10, screenHeight-20)
	}
}

func buildManifestImage(tileset *ebiten.Image, tileSize int) *ebiten.Image {
	fmt.Println("DEBUG: Building composite manifest sprite (Tiles 203, 204, 205, 206)...")
	img := ebiten.NewImage(tileSize*2, tileSize*2)
	tilesetWidth := tileset.Bounds().Dx()
	tilesPerRow := tilesetWidth / tileSize

	tileIDs := []int{203, 204, 205, 206}

	for i, tileID := range tileIDs {
		col := i % 2
		row := i / 2

		x := (tileID % tilesPerRow) * tileSize
		y := (tileID / tilesPerRow) * tileSize

		src := image.Rect(x, y, x+tileSize, y+tileSize)
		sub := tileset.SubImage(src).(*ebiten.Image)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(col*tileSize), float64(row*tileSize))
		img.DrawImage(sub, op)
	}

	return img
}
