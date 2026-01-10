package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Assets struct {
	TitleImage    *ebiten.Image
	TitleImageNYC *ebiten.Image
	BikerImage    *ebiten.Image
	PeopleImage   *ebiten.Image
	TilesetImage  *ebiten.Image

	// future
	// NYCSpriteSheet *sprites.AsepriteSheet //??? Have it as TilesetImage now
}

// NOTE: LOOK AT EMBED.GO to embed this files in so WASM works!
func LoadAssets() *Assets {
	title, err := loadImage("art/ac99_title.png")
	if err != nil {
		log.Fatal(err)
	}

	nycTitle, err := loadImage("art/aseprite_files/nyc_title.png")
	if err != nil {
		log.Fatal(err)
	}

	biker, err := loadImage("art/aseprite_files/biker.png")
	if err != nil {
		log.Fatal(err)
	}

	people, err := loadImage("art/aseprite_files/people.png")
	if err != nil {
		log.Fatal(err)
	}

	tileset, _, err := ebitenutil.NewImageFromFile("assets/NEW_nyc_spritesheet-Recovered.png")
	if err != nil {
		panic(err)
	}

	return &Assets{
		TitleImage:    title,
		TitleImageNYC: nycTitle,
		BikerImage:    biker,
		PeopleImage:   people,
		TilesetImage:  tileset,
	}
}
