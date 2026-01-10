package main

import "image"

// Entity is anything that can collide in the world
type Entity interface {
	Bounds() image.Rectangle
	OnCollision(other Entity)
}
