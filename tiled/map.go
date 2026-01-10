package tiled

// Map represents a Tiled map
type Map struct {
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	TileWidth  int       `json:"tilewidth"`
	TileHeight int       `json:"tileheight"`
	Layers     []Layer   `json:"layers"`
	Tilesets   []Tileset `json:"tilesets"`
}

// Layer represents a layer in Tiled. It can be a tile layer, group, or object layer.
type Layer struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`   // "tilelayer" | "group" | "objectgroup"
	Data    []uint32 `json:"data"`   // only for tile layers
	Layers  []Layer  `json:"layers"` // for group layers
	Visible bool     `json:"visible"`

	// Only for object layers
	Objects []Object `json:"objects,omitempty"`
}

// Tileset represents a Tiled tileset
type Tileset struct {
	FirstGID    int    `json:"firstgid"`
	Image       string `json:"image"`
	ImageWidth  int    `json:"imagewidth"`
	ImageHeight int    `json:"imageheight"`
	TileWidth   int    `json:"tilewidth"`
	TileHeight  int    `json:"tileheight"`
}

// Object represents an individual object in an object layer
type Object struct {
	ID         int              `json:"id"`
	Name       string           `json:"name"`
	Type       string           `json:"type"` // "taxi", "player", "checkpoint", etc.
	X          float64          `json:"x"`
	Y          float64          `json:"y"`
	Properties []ObjectProperty `json:"properties,omitempty"`
}

// ObjectProperty represents a custom property attached to an object
type ObjectProperty struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"` // "string", "int", "float", "bool"
	Value interface{} `json:"value"`
}

// Spawn is a simplified struct used by your game code
// after extracting object layer info
type Spawn struct {
	X, Y      float64
	Type      string // "taxi", "player", "checkpoint", etc.
	Direction string // optional, e.g., for taxis
}

// Helper: get string property from an object
func (o *Object) GetStringProperty(name, fallback string) string {
	for _, p := range o.Properties {
		if p.Name == name {
			if str, ok := p.Value.(string); ok {
				return str
			}
		}
	}
	return fallback
}
