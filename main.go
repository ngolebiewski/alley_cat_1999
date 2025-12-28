package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	titleImage *ebiten.Image
}

func NewGame() *Game {
	img, err := loadImage("art/ac99_title.png")
	if err != nil {
		log.Fatal(err)
	}

	return &Game{
		titleImage: img,
	}
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Alley Cat 1999")
	op := &ebiten.DrawImageOptions{}

	// Center the image
	size := g.titleImage.Bounds().Size()
	op.GeoM.Translate(
		float64((320-size.X)/2),
		float64((240-size.Y)/2),
	)

	screen.DrawImage(g.titleImage, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Alley Cat 1999")

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
