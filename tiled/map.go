package tiled

type Map struct {
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	TileWidth  int       `json:"tilewidth"`
	TileHeight int       `json:"tileheight"`
	Layers     []Layer   `json:"layers"`
	Tilesets   []Tileset `json:"tilesets"`
}

type Layer struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"` // tilelayer | group
	Data    []uint32 `json:"data"`
	Layers  []Layer  `json:"layers"` // for group layers
	Visible bool     `json:"visible"`
}

type Tileset struct {
	FirstGID    int    `json:"firstgid"`
	Image       string `json:"image"`
	ImageWidth  int    `json:"imagewidth"`
	ImageHeight int    `json:"imageheight"`
	TileWidth   int    `json:"tilewidth"`
	TileHeight  int    `json:"tileheight"`
}
