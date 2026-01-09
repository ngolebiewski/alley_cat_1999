package main

import (
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// This just checks to see if the virtual joystick gets appended to the UI on touch screens/mobile for WASM in browser
// See input_desktop.go and input_mobile.go
var isMobile = false

func init() {
	isMobile = isMobileBrowser()
}

// Starting up a State Machine here, to switch between scenes in the game
// Title -> Race ( + stack for cutscenes, map, rooms, etc. ) -> End Scene
type Scene interface {
	Update() error
	Draw(screen *ebiten.Image)
}

type Game struct {
	scene  Scene
	assets *Assets // ALL game assets are embeded for WASM builds
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
	assets := LoadAssets()
	g := &Game{
		assets: assets,
	}
	g.scene = NewTitleScene(g)
	return g
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Alley Cat 1999")

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
