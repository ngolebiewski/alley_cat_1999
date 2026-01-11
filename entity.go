package main

import (
	"image"

	"github.com/ngolebiewski/alley_cat_1999/tiled"
)

// Entity is anything that can collide in the world
type Entity interface {
	Bounds() image.Rectangle
	OnCollision(other Entity, grid *tiled.CollisionGrid)
}
