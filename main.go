package main

import (
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var test_img *ebiten.Image

// Starting up a State Machine here, to switch between scenes in the game
// Title -> Race ( + stack for cutscenes, map, rooms, etc. ) -> End Scene
type Scene interface {
	Update() error
	Draw(screen *ebiten.Image)
}

func init() {
	var err error
	test_img, _, err = ebitenutil.NewImageFromFile("art/aseprite_files/random_test_img.png")
	if err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	scene      Scene
	titleImage *ebiten.Image
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	return g.scene.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.scene.Draw(screen)
}

func NewGame() *Game {
	g := &Game{}
	g.scene = NewTitleScene(g)
	return g
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Alley Cat 1999\nPress Space to Start")

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
