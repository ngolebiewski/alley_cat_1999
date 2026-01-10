package tiled

import (
	"encoding/json"
	"io/fs"
)

func LoadMapFS(fsys fs.FS, path string) (*Map, error) {
	b, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}

	var m Map
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	return &m, nil
}
