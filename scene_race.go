package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/ngolebiewski/alley_cat_1999/retrotrack"
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

	// Mission Data
	manifest *Manifest

	// Fade-in & Fade-out stuff
	fader     *Fader
	isExiting bool

	// CPU entities + Collision System
	taxiManager  *TaxiManager
	npcManager   *NPCManager
	collisionSys *CollisionSystem
}

func NewRaceScene(game *Game, mfest *Manifest) *RaceScene {
	// m, err := tiled.LoadMapFS(embeddedAssets, "assets/nyc_1_TEST..tmj")
	m, err := tiled.LoadMapFS(embeddedAssets, "assets/nyc_1..tmj")
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
		game:         game,
		hud:          NewHUDOverlay(),
		player:       NewPlayer(game.assets.BikerImage, 50, 400, 32, 32),
		mapData:      m,
		mapDraw:      renderer,
		fader:        NewFader(0, 0.5), // <--- Start at 1.0 (fully black)
		collisionSys: &CollisionSystem{game: game},
		manifest:     mfest,
	}

	scene.npcManager = NewNPCManager(160, 420, scene)

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
	scene.hud.maxCheck = len(mfest.Checkpoints) //sets the number of checkpoints on the HUD

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
	// 1. Update the Fader first (handles alpha timing)
	s.fader.Update()

	// 2. Scene Transition Logic
	// If we are fading out and the fader hit 1.0 alpha, swap the scene
	if s.isExiting && s.fader.Finished {
		retrotrack.Stop()
		s.game.scene = NewEndScene(s.game, s.hud.elapsedTimeStr(), s.player.cash)
		return nil
	}

	// 3. Handle Exit Trigger
	// Start the fade out process when ESC is pressed
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && !s.isExiting {
		s.isExiting = true
		s.fader = NewFader(FadeOut, 0.25) // 0.5 seconds to black
	}

	// 4. Pause Logic
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || s.isButtonPressed("START") {
		s.paused = !s.paused
	}

	// 5. Early Return
	// Stop world updates if the game is paused or we are currently fading out
	if s.paused || s.isExiting {
		return nil
	}

	// 6. Gather Input
	var inX, inY float64
	jx, jy := s.getJoystickVector()

	// Keyboard or Virtual Joystick
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || jx < -0.2 {
		inX = -1
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) || jx > 0.2 {
		inX = 1
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) || jy < -0.2 {
		inY = -1
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) || jy > 0.2 {
		inY = 1
	}

	toggleAxis := inpututil.IsKeyJustPressed(ebiten.KeySpace) || s.isButtonJustPressed("A")
	toggleMount := inpututil.IsKeyJustPressed(ebiten.KeyB) || s.isButtonJustPressed("B")

	// 7. Physics & Movement (The Order Matters!)

	// A. Calculate Player Velocity/Animation based on input
	s.player.UpdateInput(inX, inY, toggleAxis, toggleMount)

	// B. Move player and resolve Tiled map collisions (walls)
	s.movePlayerWithCollisionGrid()
	s.clampPlayer()

	// C. Update Taxis (Movement & internal timers)
	s.taxiManager.Update(s.player.x, s.player.y)

	// NEW: Update NPCs
	if s.npcManager != nil {
		// Pass manifest for AI targets and HUD time for their finish times
		// s.npcManager.Update(s.manifest, s.taxiManager.taxis, s, s.hud.timer.Seconds())
		s.npcManager.Update(s.manifest, s.taxiManager.taxis, s, s.collide, 666.0)
	}

	// D. Resolve Entity Collisions (Player vs Taxis, Taxi vs Taxi)
	// This uses the collision_system.go logic we discussed
	s.collisionSys.Update(s.player, s.taxiManager.taxis, s.npcManager.Bikers, s.collide, s.camera)

	s.hud.health = float32(s.player.health) / 100.0
	s.hud.cash = s.player.cash

	// 8. Camera & UI
	px, py := s.player.Center()
	s.camera.Follow(px, py)

	if isMobile {
		s.updateJoystick()
	}

	// CHECKPOINTS STUFF
	// --- CHECKPOINT LOGIC ---
	if s.manifest != nil {
		px, py := s.player.Center()
		allRegularDone := true

		// First Pass: Update all NPCs and check regular checkpoints
		for _, cp := range s.manifest.Checkpoints {
			// Always update the NPC animation
			cp.Client.Update()

			if cp.IsFinishLine {
				continue // Skip finish line for now to check if others are done
			}

			if !cp.IsComplete {
				allRegularDone = false // At least one stop is left!

				// Distance Check
				dx, dy := px-cp.X, py-cp.Y
				if (dx*dx + dy*dy) < 32*32 {
					cp.IsComplete = true
					s.player.cash += 100
					retrotrack.PlayManifestSound()

					// --- UPDATE HUD HERE ---
					s.hud.checkpoints += 1

					fmt.Printf("DEBUG: Delivered to %s!\n", cp.Name)
				}
			}
		}

		// Second Pass: Check Finish Line only if others are done
		if allRegularDone {
			for _, cp := range s.manifest.Checkpoints {
				if cp.IsFinishLine && !cp.IsComplete {
					dx, dy := px-cp.X, py-cp.Y
					if (dx*dx + dy*dy) < 32*32 {
						cp.IsComplete = true

						// Start the Exit Transition to End Scene
						s.isExiting = true
						s.fader = NewFader(FadeOut, 0.25)
					}
				}
			}
		}
	}

	// GAME OVER
	if s.player.state == StateHospital && !s.isExiting {
		s.isExiting = true
		// Small delay or fader before switching
		s.game.scene = NewGameOverScene(s.game, s.manifest)
		return nil
	}

	return nil
}

func (s *RaceScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 20, 255})
	ebitenutil.DebugPrintAt(screen, "Press ESC to exit | 'F' for Full Screen\n'SPACE' to flip vert/horiz | 'B' to get on/off bike", 0, screenHeight-30)

	// MAP FIRST
	s.mapDraw.Draw(screen, s.camera.X, s.camera.Y)

	// 2. CHECKPOINTS / CLIENTS
	if s.manifest != nil {
		for _, cp := range s.manifest.Checkpoints {
			cp.Draw(screen, s.camera) // This uses the Draw method in manifest.go
		}
	}
	// 3. RIVAL NPC BIKERS
	// // Removing because the AI they run on is really ANNOYING, and decreases the fun at the moment.
	// // A simpler approach could be good. Like setting up a node network/graph...not happening this late in the game jam!
	// so commenting out for now. Even though it will still run in the background. Hopefully it will all still compile -Nick

	// if s.npcManager != nil {
	// 	// We use the player's biker image as the base,
	// 	// the NPC Draw function handles the color tinting.
	// 	s.npcManager.Draw(screen, s.camera, s.game.assets.BikerImage)
	// }

	//ENTITIES
	// s.player.Draw(screen) // this was the non camera way to draw
	s.player.DrawWithCamera(screen, s.camera)

	s.taxiManager.Draw(screen, s.camera)

	s.hud.Draw(screen)

	if isDebugMode {
		s.drawCollisionDebug(screen)
	}

	if isMobile {
		s.drawMobileUI(screen)
	}
	if s.paused {
		ebitenutil.DebugPrintAt(screen, "PAUSED", 140, 110)
		s.drawPauseOverlay(screen)
	}

	// 4. DRAW FADER LAST
	s.fader.Draw(screen)
}

func (s *RaceScene) movePlayerWithCollisionGrid() {
	// Check if player is ALREADY inside a wall before we do anything
	// This happens if a taxi ejection forced them into a building
	wasStuck := s.collidesAt(s.player.x, s.player.y)

	// 1. Resolve X Movement
	oldX := s.player.x
	s.player.x += s.player.velX

	if s.collidesAt(s.player.x, s.player.y) {
		// If we weren't stuck but now we are: BLOCK IT (Normal wall behavior)
		// If we WERE stuck and we are STILL stuck: BLOCK IT (Prevents moving deeper into wall)
		if !wasStuck || s.collidesAt(s.player.x, s.player.y) {
			s.player.x = oldX
			s.player.velX = 0
		}
		// Note: If wasStuck was true, but s.collidesAt is now false,
		// the code allows the move because it means the player is escaping!
	}

	// 2. Resolve Y Movement
	oldY := s.player.y
	s.player.y += s.player.velY

	if s.collidesAt(s.player.x, s.player.y) {
		// If this move results in a collision...
		if !wasStuck {
			// Normal case: Just hit a wall, snap back
			s.player.y = oldY
			s.player.velY = 0
		} else {
			// Anti-stuck case: If we are already inside a wall,
			// only snap back if the move doesn't get us out.
			if s.collidesAt(s.player.x, s.player.y) {
				s.player.y = oldY
				s.player.velY = 0
			}
		}
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

// Returns the Tile GID at a specific world coordinate for a specific layer name
func (s *RaceScene) getTileIDAt(worldX, worldY float64, layerName string) int {
	tx, ty := int(worldX/32), int(worldY/32)

	// Boundary Check
	if tx < 0 || tx >= s.mapData.Width || ty < 0 || ty >= s.mapData.Height {
		return 0
	}

	// Use the existing recursive function to find the layer
	layer := findLayerRecursive(s.mapData.Layers, layerName)
	if layer == nil {
		return 0
	}

	idx := ty*s.mapData.Width + tx
	if idx >= 0 && idx < len(layer.Data) {
		// FIX: Cast the uint32 value to int to match function signature
		return int(layer.Data[idx])
	}

	return 0
}

// Checks if the tile at Layer 3 (Blocked) has a non-zero GID
func (s *RaceScene) isTileBlocked(worldX, worldY float64) bool {
	blockedID := s.getTileIDAt(worldX, worldY, "COLLIDE-Road and Sidewalks BLOCKED")
	return blockedID != 0
}
