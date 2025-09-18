package main

import (
	"encoding/gob"
	"os"
	"path"
)

func SaveFilenameMap(root string, filenameMap map[string][]CodeLine) error {
	path := path.Join(root, ".gitaegis")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	gob.Register(CodeLine{})
	return enc.Encode(filenameMap)
}

func LoadFilenameMap(root string) (map[string][]CodeLine, error) {
	path := path.Join(root, ".gitaegis")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty map if file does not exist
			return make(map[string][]CodeLine), nil
		}
		return nil, err
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	gob.Register(CodeLine{})

	filenameMap := make(map[string][]CodeLine)
	if err := dec.Decode(&filenameMap); err != nil {
		return nil, err
	}

	return filenameMap, err
}
