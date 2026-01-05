package sprites

type AsepriteSheet struct {
	Frames map[string]struct {
		Frame struct {
			X int `json:"x"`
			Y int `json:"y"`
			W int `json:"w"`
			H int `json:"h"`
		} `json:"frame"`
		Duration int `json:"duration"`
	} `json:"frames"`

	Meta struct {
		FrameTags []struct {
			Name string `json:"name"`
			From int    `json:"from"`
			To   int    `json:"to"`
		} `json:"frameTags"`
	} `json:"meta"`
}
