package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/ngolebiewski/alley_cat_1999/tiled"
)

type RaceScene struct {
	game   *Game
	hud    *HUDOverlay
	player *Player
	camera *Camera
	paused bool
	stick  joystick // used by input.go for mobile/touch screen devices only its a virtual joystick!

	worldW float64
	worldH float64

	mapData *tiled.Map
	mapDraw *tiled.Renderer
	collide *tiled.CollisionGrid

	// CPU entities
	taxiManager *TaxiManager
}

func NewRaceScene(game *Game) *RaceScene {
	m, err := tiled.LoadMapFS(embeddedAssets, "assets/nyc_1_TEST..tmj")
	// m, err := tiled.LoadMapFS(embeddedAssets, "assets/nyc_1..tmj")
	if err != nil {
		panic(err)
	}

	scale := 2

	renderer := tiled.NewRenderer(
		m,
		game.assets.TilesetImage, // nyc_tileset.png
		float64(scale),           // 16px → 32px
	)

	scene := &RaceScene{
		game:    game,
		hud:     NewHUDOverlay(),
		player:  NewPlayer(game.assets.BikerImage, 160, 400, 32, 32),
		mapData: m,
		mapDraw: renderer,
	}

	worldW := m.Width * m.TileWidth * scale
	worldH := m.Height * m.TileHeight * scale

	scene.camera = NewCamera(
		screenWidth,
		screenHeight,
		worldW,
		worldH,
	)

	scene.worldW = float64(worldW)
	scene.worldH = float64(worldH)
	// scene.taxiManager.worldW = scene.worldW
	// scene.taxiManager.worldH = scene.worldH
	scene.taxiManager = NewTaxiManager(game.assets.TilesetImage, 2.0, scene.worldW, scene.worldH, m) // scale 2x
	scene.collide = tiled.BuildCollisionGrid(m)

	return scene
}

func (s *RaceScene) clampPlayer() {
	pw := s.player.w
	ph := s.player.h

	if s.player.x < 0 {
		s.player.x = 0
	}
	if s.player.y < 0 {
		s.player.y = 0
	}

	if s.player.x+pw > s.worldW {
		s.player.x = s.worldW - pw
	}
	if s.player.y+ph > s.worldH {
		s.player.y = s.worldH - ph
	}
}

func (s *RaceScene) Update() error {
	StartRaceMusic()
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		StopRaceMusic()
		s.game.scene = NewEndScene(s.game)
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || s.isButtonPressed("START") {
		s.paused = !s.paused
	}
	if s.paused {
		return nil
	}

	// 1. Gather Inputs
	var inX, inY float64
	jx, jy := s.getJoystickVector()
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || jx < -0.2 {
		inX = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || jx > 0.2 {
		inX = 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || jy < -0.2 {
		inY = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || jy > 0.2 {
		inY = 1
	}

	toggleAxis := inpututil.IsKeyJustPressed(ebiten.KeySpace) || s.isButtonJustPressed("A")
	toggleMount := inpututil.IsKeyJustPressed(ebiten.KeyB) || s.isButtonJustPressed("B")

	// 2. Update Player
	// s.player.Update(inX, inY, toggleAxis, toggleMount) // without collission checking

	// 1. Update velocity & animation
	s.player.UpdateInput(inX, inY, toggleAxis, toggleMount)

	// 2. Move with collision grid
	s.movePlayerWithCollisionGrid()
	s.clampPlayer()

	s.taxiManager.Update()

	// 3. Update Camera
	px, py := s.player.Center()
	s.camera.Follow(px, py)

	if isMobile {
		s.updateJoystick()
	}
	return nil
}

func (s *RaceScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 20, 255})
	ebitenutil.DebugPrintAt(screen, "Press ESC to exit | 'F' for Full Screen\n'SPACE' to flip vert/horiz | 'B' to get on/off bike", 0, screenHeight-30)

	// MAP FIRST
	s.mapDraw.Draw(screen, s.camera.X, s.camera.Y)

	//ENTITIES
	// s.player.Draw(screen) // this was the non camera way to draw
	s.player.DrawWithCamera(screen, s.camera)

	s.taxiManager.Draw(screen, s.camera)

	s.hud.Draw(screen)
	// s.drawCollisionDebug(screen)

	if isMobile {
		s.drawMobileUI(screen)
	}
	if s.paused {
		ebitenutil.DebugPrintAt(screen, "PAUSED", 140, 110)
	}
}

func (s *RaceScene) movePlayerWithCollisionGrid() {
	const tileSize = 32

	// 1. Compute attempted new positions
	newX := s.player.x + s.player.velX
	newY := s.player.y + s.player.velY

	// 2. Clamp X/Y to map bounds
	newX = math.Max(0, math.Min(newX, float64(s.worldW)-s.player.w))
	newY = math.Max(0, math.Min(newY, float64(s.worldH*tileSize)-s.player.h))

	// 3. Check X movement
	if !s.collidesAt(newX, s.player.y) {
		s.player.x = newX
	} else {
		s.player.velX = 0
	}

	// 4. Check Y movement
	if !s.collidesAt(s.player.x, newY) {
		s.player.y = newY
	} else {
		s.player.velY = 0
	}
}

// Check collision using CollisionGrid
func (s *RaceScene) collidesAt(px, py float64) bool {
	const tileSize = 32

	grid := s.collide
	if grid == nil {
		return false
	}

	// Convert player rectangle into tile coordinates
	x1 := int(px) / tileSize
	y1 := int(py) / tileSize
	x2 := int(px+s.player.w-1) / tileSize
	y2 := int(py+s.player.h-1) / tileSize

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

func (s *RaceScene) drawCollisionDebug(screen *ebiten.Image) {
	const tile = 32
	for y := 0; y < s.collide.Height; y++ {
		for x := 0; x < s.collide.Width; x++ {
			if s.collide.Solid[y][x] {
				ebitenutil.DrawRect(
					screen,
					float64(x*tile)-s.camera.X,
					float64(y*tile)-s.camera.Y,
					tile,
					tile,
					color.RGBA{255, 0, 0, 80},
				)
			}
		}
	}
}
