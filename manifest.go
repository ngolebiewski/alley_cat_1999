package main

import (
	"fmt"
	"image"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/ngolebiewski/alley_cat_1999/tiled"
)

type Checkpoint struct {
	Name         string // Stores the "Location" (e.g. White Space Invader)
	X, Y         float64
	IsComplete   bool
	IsFinishLine bool
	Client       *Person
}

type Person struct {
	Img        *ebiten.Image
	X, Y       float64
	StartX     float64
	Direction  int
	PaceDist   float64
	PauseTimer int
	BobTimer   float64
}

type Manifest struct {
	Checkpoints []*Checkpoint
}

func NewManifest(m *tiled.Map, peopleSheet *ebiten.Image, scale float64) *Manifest {
	fmt.Println("DEBUG: NewManifest: Extracting CHECKPOINTS from 'Spawns' layer...")

	// Call the updated extractor
	rawSpawns := tiled.ExtractManifestCheckpoints(m)

	if len(rawSpawns) == 0 {
		fmt.Println("DEBUG ERROR: Still no checkpoints found! Double check layer/object names.")
		return &Manifest{Checkpoints: []*Checkpoint{}}
	}

	fmt.Printf("DEBUG: Found %d potential checkpoints in JSON\n", len(rawSpawns))

	availablePeople := makePeople(peopleSheet)
	var allPossible []*Checkpoint

	for _, s := range rawSpawns {
		pImg := availablePeople[rand.Intn(len(availablePeople))]
		allPossible = append(allPossible, &Checkpoint{
			Name: s.Location, // Will now correctly be "White Space Invader", "Yagg Grafitti", etc.
			X:    s.X * scale,
			Y:    s.Y * scale,
			Client: &Person{
				Img:       pImg,
				X:         s.X * scale,
				Y:         s.Y * scale,
				StartX:    s.X * scale,
				Direction: 1,
				PaceDist:  50.0,
			},
		})
	}

	// Shuffle and pick mission length (prevents panic)
	rand.Shuffle(len(allPossible), func(i, j int) {
		allPossible[i], allPossible[j] = allPossible[j], allPossible[i]
	})

	numToKeep := len(allPossible)
	if len(allPossible) > 2 {
		numToKeep = rand.Intn(len(allPossible)-1) + 2
	}

	activeCPs := allPossible[:numToKeep]
	activeCPs[len(activeCPs)-1].IsFinishLine = true

	return &Manifest{Checkpoints: activeCPs}
}
func (p *Person) Update() {
	p.BobTimer += 0.05
	if p.PauseTimer > 0 {
		p.PauseTimer--
		return
	}
	p.X += float64(p.Direction) * 0.4
	if math.Abs(p.X-p.StartX) > p.PaceDist {
		p.Direction *= -1
		p.PauseTimer = rand.Intn(120) + 60
	}
}

func (cp *Checkpoint) Draw(screen *ebiten.Image, cam *Camera) {
	// 1. ALWAYS draw the Person (they don't disappear anymore)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(cp.Client.X-cam.X, cp.Client.Y-cam.Y)
	screen.DrawImage(cp.Client.Img, op)

	// 2. ONLY draw indicators if not complete
	if !cp.IsComplete {
		bob := math.Sin(cp.Client.BobTimer) * 4
		indicator := "CHECKPOINT"
		if cp.IsFinishLine {
			indicator = "FINISH" // Visual hint for the last stop
		}

		ebitenutil.DebugPrintAt(screen, indicator,
			int(cp.Client.X-cam.X)+8,
			int(cp.Client.Y-cam.Y)-20+int(bob),
		)
	}
}

func makePeople(spritesheet *ebiten.Image) []*ebiten.Image {
	var people []*ebiten.Image
	const size = 32
	if spritesheet == nil {
		return people
	}
	count := spritesheet.Bounds().Dx() / size
	for i := 0; i < count; i++ {
		rect := image.Rect(i*size, 0, (i+1)*size, size)
		people = append(people, spritesheet.SubImage(rect).(*ebiten.Image))
	}
	return people
}

func resetManifestCheckins(m *Manifest) {
	for i := range m.Checkpoints {
		m.Checkpoints[i].IsComplete = false
	}
}
