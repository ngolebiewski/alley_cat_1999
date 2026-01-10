package tiled

// ExtractTaxiSpawns scans the map and returns all taxi spawns
func ExtractTaxiSpawns(m *Map) []Spawn {
	var spawns []Spawn

	var walkLayers func(layers []Layer)
	walkLayers = func(layers []Layer) {
		for _, layer := range layers {
			if !layer.Visible {
				continue
			}

			switch layer.Type {
			case "group":
				walkLayers(layer.Layers)
			case "objectgroup":
				for _, obj := range layer.Objects {
					// Use obj.Name, because in your Tiled JSON "name":"taxi"
					if obj.Name == "taxi" {
						spawns = append(spawns, Spawn{
							X:         obj.X,
							Y:         obj.Y,
							Type:      "taxi",
							Direction: obj.GetStringProperty("direction", "RIGHT"),
						})
					}
				}
			}
		}
	}

	walkLayers(m.Layers)
	return spawns
}

// ExtractSpawns scans the map and returns all objects from a given object layer name
func ExtractSpawns(m *Map, layerName string) []Spawn {
	var spawns []Spawn

	var walkLayers func(layers []Layer)
	walkLayers = func(layers []Layer) {
		for _, layer := range layers {
			if !layer.Visible {
				continue
			}

			switch layer.Type {
			case "group":
				walkLayers(layer.Layers)
			case "objectgroup":
				if layer.Name != layerName {
					continue
				}

				for _, obj := range layer.Objects {
					spawns = append(spawns, Spawn{
						X:         obj.X,
						Y:         obj.Y,
						Type:      obj.Name,                               // name in Tiled becomes the type
						Direction: obj.GetStringProperty("direction", ""), // optional
					})
				}
			}
		}
	}

	walkLayers(m.Layers)
	return spawns
}
