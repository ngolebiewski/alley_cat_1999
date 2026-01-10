package main

import "image"

func tileRectFromID(
	tileID int,
	tileSize int,
	tilesetWidth int,
) image.Rectangle {
	tilesPerRow := tilesetWidth / tileSize

	x := (tileID % tilesPerRow) * tileSize
	y := (tileID / tilesPerRow) * tileSize

	return image.Rect(
		x,
		y,
		x+tileSize,
		y+tileSize,
	)
}
