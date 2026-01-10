package tiled

type CollisionGrid struct {
	Width  int
	Height int
	Solid  [][]bool
}

func BuildCollisionGrid(m *Map) *CollisionGrid {
	grid := &CollisionGrid{
		Width:  m.Width,
		Height: m.Height,
		Solid:  make([][]bool, m.Height),
	}

	for y := range m.Height {
		grid.Solid[y] = make([]bool, m.Width)
	}

	var walk func(layers []Layer)
	walk = func(layers []Layer) {
		for _, layer := range layers {
			if layer.Type == "group" {
				walk(layer.Layers)
				continue
			}

			if layer.Type != "tilelayer" {
				continue
			}

			if !IsCollideLayer(layer.Name) {
				continue
			}

			for i, gid := range layer.Data {
				if gid == 0 {
					continue
				}

				x := i % m.Width
				y := i / m.Width
				grid.Solid[y][x] = true
			}
		}
	}

	walk(m.Layers)
	return grid
}
