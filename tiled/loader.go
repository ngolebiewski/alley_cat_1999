package tiled

import (
	"encoding/json"
	"os"
)

func LoadMap(path string) (*Map, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m Map
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	return &m, nil
}
