package main

// This is where we embed the images and spritesheets for the game, so when it is published as WASM on the web
// we'll be ready to go, as the web can't (easily?) access the file system like a desktop version of the game can.

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed art/ac99_title.png
var embeddedAssets embed.FS

func loadImage(path string) (*ebiten.Image, error) {
	data, err := embeddedAssets.ReadFile(path)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}
