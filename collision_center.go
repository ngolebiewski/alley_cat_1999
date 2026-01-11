package main

type CollisionSystem struct {
	game *Game
}

// Update returns true if the player hit a taxi
func (cs *CollisionSystem) Update(player *Player, taxis []*Taxi) bool {
	playerBounds := player.Bounds()
	hit := false

	for _, taxi := range taxis {
		if taxi.crashed {
			continue
		}

		// Check Player vs Taxi
		if playerBounds.Overlaps(taxi.Bounds()) {
			player.OnCollision(taxi)
			taxi.OnCollision(player)
			hit = true // Mark that a collision occurred
		}

		// Check Taxi vs Taxi (Optional Chaos)
		for _, other := range taxis {
			if taxi == other || other.crashed {
				continue
			}
			if taxi.Bounds().Overlaps(other.Bounds()) {
				taxi.OnCollision(other)
				other.OnCollision(taxi)
			}
		}
	}
	return hit
}
