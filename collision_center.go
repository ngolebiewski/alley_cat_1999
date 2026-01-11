package main

import (
	"github.com/ngolebiewski/alley_cat_1999/tiled"
)

type CollisionSystem struct {
	game *Game
}

// Update returns true if the player hit a taxi
func (cs *CollisionSystem) Update(player *Player, taxis []*Taxi, rivals []*NPCBiker, grid *tiled.CollisionGrid) bool {
	playerBounds := player.Bounds()
	playerHit := false

	// 1. Taxis are the primary "Hazard"
	for _, taxi := range taxis {
		if taxi.crashed {
			continue
		}
		taxiBounds := taxi.Bounds()

		// Player vs Taxi
		if playerBounds.Overlaps(taxiBounds) {
			player.OnCollision(taxi, grid)
			taxi.OnCollision(player, grid)
			playerHit = true
		}

		// Rivals vs Taxi
		for _, rival := range rivals {
			if rival.Finished {
				continue
			}
			if rival.Bounds().Overlaps(taxiBounds) {
				rival.OnCollision(taxi, grid)
				taxi.OnCollision(rival, grid)
			}
		}

		// Taxi vs Taxi
		for _, otherTaxi := range taxis {
			if taxi == otherTaxi || otherTaxi.crashed {
				continue
			}
			if taxiBounds.Overlaps(otherTaxi.Bounds()) {
				taxi.OnCollision(otherTaxi, grid)
				otherTaxi.OnCollision(taxi, grid)
			}
		}
	}

	// 2. Player vs Rivals (Bumping into each other)
	for _, rival := range rivals {
		if rival.Finished {
			continue
		}
		if playerBounds.Overlaps(rival.Bounds()) {
			player.OnCollision(rival, grid)
			rival.OnCollision(player, grid)
		}
	}

	return playerHit
}
