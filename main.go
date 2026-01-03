package main

import (
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var test_img *ebiten.Image

func init() {
	var err error
	test_img, _, err = ebitenutil.NewImageFromFile("art/aseprite_files/random_test_img.png")
	if err != nil {
		log.Fatal(err)
	}
}

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

	ti_op := &ebiten.DrawImageOptions{}
	ti_op.GeoM.Translate(200.0, 50.0)
	screen.DrawImage(test_img, ti_op)
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
