package main

import (
	"image/color"

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

	mapData *tiled.Map
	mapDraw *tiled.Renderer
}

func NewRaceScene(game *Game) *RaceScene {
	m, err := tiled.LoadMapFS(embeddedAssets, "assets/nyc_1_TEST..tmj")
	if err != nil {
		panic(err)
	}

	renderer := tiled.NewRenderer(
		m,
		game.assets.TilesetImage, // nyc_tileset.png
		2.0,                      // 16px → 32px
	)

	scene := &RaceScene{
		game:    game,
		hud:     NewHUDOverlay(),
		player:  NewPlayer(game.assets.BikerImage, 160, 120, 32, 32),
		mapData: m,
		mapDraw: renderer,
	}

	worldW := m.Width * m.TileWidth * 2
	worldH := m.Height * m.TileHeight * 2

	scene.camera = NewCamera(
		screenWidth,
		screenHeight,
		worldW,
		worldH,
	)

	return scene
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
	s.player.Update(inX, inY, toggleAxis, toggleMount)

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
	s.hud.Draw(screen)

	if isMobile {
		s.drawMobileUI(screen)
	}
	if s.paused {
		ebitenutil.DebugPrintAt(screen, "PAUSED", 140, 110)
	}
}
