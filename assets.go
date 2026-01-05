package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Assets struct {
	TitleImage *ebiten.Image

	// future
	// BikerImage *ebiten.Image
	// NYCSpriteSheet *sprites.AsepriteSheet
}

func LoadAssets() *Assets {
	title, err := loadImage("art/ac99_title.png")
	if err != nil {
		log.Fatal(err)
	}

	return &Assets{
		TitleImage: title,
	}
}
