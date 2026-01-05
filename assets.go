package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Assets struct {
	TitleImage *ebiten.Image
	BikerImage *ebiten.Image

	// future
	// NYCSpriteSheet *sprites.AsepriteSheet
}

func LoadAssets() *Assets {
	title, err := loadImage("art/ac99_title.png")
	if err != nil {
		log.Fatal(err)
	}

	biker, err := loadImage("art/aseprite_files/biker.png")
	if err != nil {
		log.Fatal(err)
	}

	return &Assets{
		TitleImage: title,
		BikerImage: biker,
	}
}
