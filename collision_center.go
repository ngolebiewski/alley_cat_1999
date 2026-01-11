package main

import (
	"github.com/ngolebiewski/alley_cat_1999/tiled"
)

type CollisionSystem struct {
	game *Game
}

// Update returns true if the player hit a taxi
func (cs *CollisionSystem) Update(player *Player, taxis []*Taxi, grid *tiled.CollisionGrid) bool {
	playerBounds := player.Bounds()
	hit := false

	for _, taxi := range taxis {
		if taxi.crashed {
			continue
		}

		// 1. Check Player vs Taxi
		if playerBounds.Overlaps(taxi.Bounds()) {
			// Pass the grid so the Player can check buildings during ejection
			player.OnCollision(taxi, grid)
			taxi.OnCollision(player, grid)
			hit = true
		}

		// 2. Check Taxi vs Taxi (Optional Chaos)
		for _, other := range taxis {
			if taxi == other || other.crashed {
				continue
			}
			if taxi.Bounds().Overlaps(other.Bounds()) {
				// Even if taxis don't use the grid, they must accept it
				// to satisfy the new Entity interface
				taxi.OnCollision(other, grid)
				other.OnCollision(taxi, grid)
			}
		}
	}
	return hit
}
