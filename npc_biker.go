package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ngolebiewski/alley_cat_1999/tiled"
)

type NPCBiker struct {
	Name       string
	x, y       float64
	w, h       float64
	velX, velY float64
	speed      float64
	color      ebiten.ColorScale

	// Progress
	Inventory     map[string]bool
	Finished      bool
	FinalTime     float64
	CurrentTarget *Checkpoint

	// Route Logic: Unique order of checkpoints for this specific NPC
	RouteOrder []string

	// AI Logic
	StuckTimer   int
	LastX, LastY float64
	ticks        int
	animOffset   int

	// Start Delay
	StartDelayTicks int

	// Animation State
	dir   int
	frame int
}

type NPCManager struct {
	Bikers []*NPCBiker
}

func NewNPCManager(startX, startY float64, scene *RaceScene) *NPCManager {
	manager := &NPCManager{}
	rand.Seed(time.Now().UnixNano())

	rivalConfigs := []struct {
		name  string
		color []float32
	}{
		{"Purple Haze", []float32{0.8, 0.4, 1.0}},
		{"Blue Streak", []float32{0.4, 0.4, 1.0}},
		{"Green Machine", []float32{0.4, 1.0, 0.4}},
		{"Yellow Jacket", []float32{1.0, 1.0, 0.4}},
	}

	// Prepare list of checkpoint names (excluding finish line)
	var cpNames []string
	for _, cp := range scene.manifest.Checkpoints {
		if !cp.IsFinishLine {
			cpNames = append(cpNames, cp.Name)
		}
	}

	for i, config := range rivalConfigs {
		cs := ebiten.ColorScale{}
		cs.Scale(config.color[0], config.color[1], config.color[2], 1.0)

		// Create a unique shuffled route for this specific NPC
		npcRoute := make([]string, len(cpNames))
		copy(npcRoute, cpNames)
		rand.Shuffle(len(npcRoute), func(i, j int) {
			npcRoute[i], npcRoute[j] = npcRoute[j], npcRoute[i]
		})

		biker := &NPCBiker{
			Name:            config.name,
			x:               startX + float64(i*32),
			y:               startY,
			w:               18,
			h:               18,
			speed:           1.5 + (rand.Float64() * 0.3),
			color:           cs,
			Inventory:       make(map[string]bool),
			RouteOrder:      npcRoute,
			dir:             3,
			animOffset:      rand.Intn(60),
			StartDelayTicks: 30 + rand.Intn(90),
		}
		manager.Bikers = append(manager.Bikers, biker)
	}
	return manager
}

func (m *NPCManager) Update(manifest *Manifest, taxis []*Taxi, scene *RaceScene, grid *tiled.CollisionGrid, currentTime float64) {
	for _, b := range m.Bikers {
		b.Update(manifest, taxis, scene, grid, currentTime)
	}
}

func (m *NPCManager) Draw(screen *ebiten.Image, cam *Camera, img *ebiten.Image) {
	for _, b := range m.Bikers {
		b.Draw(screen, cam, img)
	}
}

// --- Physics & Helpers ---

func (n *NPCBiker) Bounds() image.Rectangle {
	return image.Rect(int(n.x), int(n.y), int(n.x+n.w), int(n.y+n.h))
}

func (n *NPCBiker) wouldCollideAt(newX, newY float64, grid *tiled.CollisionGrid) bool {
	if grid == nil {
		return false
	}
	const tileSize = 32
	points := [][2]float64{
		{newX, newY}, {newX + n.w, newY},
		{newX, newY + n.h}, {newX + n.w, newY + n.h},
		{newX + n.w/2, newY + n.h/2},
	}
	for _, p := range points {
		ix, iy := int(p[0])/tileSize, int(p[1])/tileSize
		if iy >= 0 && iy < grid.Height && ix >= 0 && ix < grid.Width {
			if grid.Solid[iy][ix] {
				return true
			}
		}
	}
	return false
}

// --- AI & Update ---

func (n *NPCBiker) Update(manifest *Manifest, taxis []*Taxi, scene *RaceScene, grid *tiled.CollisionGrid, totalTime float64) {
	if n.Finished {
		return
	}
	n.ticks++
	if n.ticks < n.StartDelayTicks {
		return
	}

	n.findTarget(manifest)
	n.applyManhattanAI(scene, grid, taxis)

	oldX, oldY := n.x, n.y
	n.x += n.velX
	if n.wouldCollideAt(n.x, n.y, grid) {
		n.x = oldX
		if !n.wouldCollideAt(n.x, n.y+2, grid) {
			n.y += 0.3
		}
	}

	n.y += n.velY
	if n.wouldCollideAt(n.x, n.y, grid) {
		n.y = oldY
		if !n.wouldCollideAt(n.x+2, n.y, grid) {
			n.x += 0.3
		}
	}

	if math.Abs(n.x-n.LastX)+math.Abs(n.y-n.LastY) < 0.05 {
		n.StuckTimer++
	} else {
		n.StuckTimer = 0
	}
	n.LastX, n.LastY = n.x, n.y
	if n.StuckTimer > 180 {
		n.respawnOnRoad(scene, grid)
	}

	n.updateAnimation()
	n.checkCheckpoints(totalTime)
}

func (n *NPCBiker) applyManhattanAI(scene *RaceScene, grid *tiled.CollisionGrid, taxis []*Taxi) {
	if n.CurrentTarget == nil {
		return
	}

	dx, dy := n.CurrentTarget.X-n.x, n.CurrentTarget.Y-n.y
	distToTarget := math.Hypot(dx, dy)

	moveX, moveY := 0.0, 0.0

	// 1. Primary Manhattan Direction
	if math.Abs(dx) > math.Abs(dy) {
		moveX = n.speed
		if dx < 0 {
			moveX = -n.speed
		}
	} else {
		moveY = n.speed
		if dy < 0 {
			moveY = -n.speed
		}
	}

	// 2. Road-Staying Logic (Only if not right next to the target)
	if distToTarget > 64 {
		currentTile := scene.getTileIDAt(n.x+(n.w/2), n.y+(n.h/2), "Roads and Sidewalks")

		// If we are NOT on a roadway (Tile 2), try to nudge back toward it
		if currentTile != 2 {
			// Scan a small area around the NPC to find where the road is
			foundRoad := false
			offsets := []float64{-32, 32, -64, 64}

			for _, off := range offsets {
				// If moving vertically, check horizontal for road
				if moveY != 0 {
					if scene.getTileIDAt(n.x+off, n.y, "Roads and Sidewalks") == 2 {
						moveX = n.speed
						if off < 0 {
							moveX = -n.speed
						}
						foundRoad = true
						break
					}
				} else { // If moving horizontally, check vertical for road
					if scene.getTileIDAt(n.x, n.y+off, "Roads and Sidewalks") == 2 {
						moveY = n.speed
						if off < 0 {
							moveY = -n.speed
						}
						foundRoad = true
						break
					}
				}
			}

			// If we are really deep in the sidewalk and can't see the road,
			// just prioritize getting back to the center of the screen/map area
			if !foundRoad {
				// fallback: nudge toward middle of typical road widths
			}
		}
	}

	// 3. Taxi Avoidance (Keep your existing logic)
	for _, t := range taxis {
		if math.Hypot(n.x-t.x, n.y-t.y) < 60 {
			if moveX != 0 {
				moveY = n.speed
				if t.y > n.y {
					moveY = -n.speed
				}
			} else {
				moveX = n.speed
				if t.x > n.x {
					moveX = -n.speed
				}
			}
		}
	}

	// 4. Collision Look-ahead (Keep your existing logic)
	if n.wouldCollideAt(n.x+moveX*15, n.y+moveY*15, grid) {
		if moveX != 0 {
			moveX = 0
			moveY = n.speed
			if dy < 0 {
				moveY = -n.speed
			}
		} else {
			moveY = 0
			moveX = n.speed
			if dx < 0 {
				moveX = -n.speed
			}
		}
	}

	n.velX, n.velY = moveX, moveY
}

func (n *NPCBiker) findTarget(manifest *Manifest) {
	// Look through the shuffled route first
	for _, name := range n.RouteOrder {
		if !n.Inventory[name] {
			// Find the actual checkpoint object by name
			for _, cp := range manifest.Checkpoints {
				if cp.Name == name {
					n.CurrentTarget = cp
					return
				}
			}
		}
	}
	// If all route checkpoints are done, go to finish line
	for _, cp := range manifest.Checkpoints {
		if cp.IsFinishLine {
			n.CurrentTarget = cp
			return
		}
	}
}

func (n *NPCBiker) checkCheckpoints(totalTime float64) {
	if n.CurrentTarget == nil {
		return
	}
	if math.Hypot(n.x-n.CurrentTarget.X, n.y-n.CurrentTarget.Y) < 48 {
		n.Inventory[n.CurrentTarget.Name] = true
		if n.CurrentTarget.IsFinishLine {
			n.Finished = true
			n.FinalTime = totalTime
		}
	}
}

func (n *NPCBiker) respawnOnRoad(scene *RaceScene, grid *tiled.CollisionGrid) {
	n.StuckTimer = 0
	for radius := 32.0; radius < 400.0; radius += 32.0 {
		for angle := 0.0; angle < math.Pi*2; angle += math.Pi / 4 {
			tx, ty := n.x+math.Cos(angle)*radius, n.y+math.Sin(angle)*radius
			if !n.wouldCollideAt(tx, ty, grid) && scene.getTileIDAt(tx, ty, "Roads and Sidewalks") == 2 {
				n.x, n.y = tx, ty
				return
			}
		}
	}
}

func (n *NPCBiker) updateAnimation() {
	if n.velX == 0 && n.velY == 0 {
		n.frame = 0
		return
	}
	if math.Abs(n.velX) > math.Abs(n.velY) {
		if n.velX > 0 {
			n.dir = 3
		} else {
			n.dir = 2
		}
	} else {
		if n.velY > 0 {
			n.dir = 0
		} else {
			n.dir = 1
		}
	}
	if n.ticks%12 == 0 {
		n.frame = (n.frame + 1) % 3
	}
}

func (n *NPCBiker) Draw(screen *ebiten.Image, cam *Camera, playerImg *ebiten.Image) {
	const size = 32
	op := &ebiten.DrawImageOptions{}
	op.ColorScale = n.color
	if n.dir == 2 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(size, 0)
	}
	op.GeoM.Translate(n.x-cam.X, n.y-cam.Y)
	sx := n.frame * size
	sub := playerImg.SubImage(image.Rect(sx, 0, sx+size, size)).(*ebiten.Image)
	screen.DrawImage(sub, op)

	if isDebugMode {
		ebitenutil.DebugPrintAt(screen, n.Name, int(n.x-cam.X), int(n.y-cam.Y-15))
		vector.StrokeRect(screen, float32(n.x-cam.X), float32(n.y-cam.Y), float32(n.w), float32(n.h), 1, color.RGBA{255, 0, 255, 255}, false)
	}
}

// OnCollision satisfies the Entity interface for the CollisionCenter
func (n *NPCBiker) OnCollision(other Entity, grid *tiled.CollisionGrid) {
	switch other.(type) {
	case *Taxi:
		// Bounce back harder when hitting a taxi
		n.velX, n.velY = -n.velX*1.5, -n.velY*1.5
		n.StuckTimer += 10
	case *Player, *NPCBiker:
		// Slight bounce when hitting other bikers or the player
		n.velX, n.velY = -n.velX*0.5, -n.velY*0.5
	}
}
