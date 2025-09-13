package main

import (
	"encoding/gob"
	"fmt"
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

func main() {
	root := "/home/holyknight101/Documents/Projects/Personal/e-form"

	// Run folder traversal
	result, err := iterFolder(root)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Save the map after traversal
	if err := SaveFilenameMap(root, result); err != nil {
		fmt.Println("Error saving:", err)
	}

	// Load the map back
	loadedMap, err := LoadFilenameMap(root)
	if err != nil {
		fmt.Println("Error loading:", err)
		return
	}

	// Iterate over loaded data
	for file, lines := range loadedMap {
		if len(lines) > 0 {
			for l := range lines {
				fmt.Printf("File: %s, Line %d: %s\n", file, lines[l].Index, lines[l].Line)
			}
		}
	}
}
