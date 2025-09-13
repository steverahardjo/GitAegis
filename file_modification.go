package main

import (
	"encoding/gob"
	"fmt"
	"os"
)

// Global map to hold data
var PrevfilenameMap = make(map[string][]CodeLine)

func SaveFilenameMap() error {
	f, err := os.Create(".gitaegis")
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	gob.Register(CodeLine{})
	return enc.Encode(filenameMap)
}

func LoadFilenameMap(root string) error {
	f, err := os.Open(".gitaegis")
	if err != nil {
		if os.IsNotExist(err) {
			PrevfilenameMap = make(map[string][]CodeLine)
			return nil
		}
		return err
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	gob.Register(CodeLine{})
	return dec.Decode(&filenameMap)
}

func main() {
	// Call iterFolder, assumed defined in another file in same package
	result, err := iterFolder("/home/holyknight101/Documents/Projects/Personal/patent-analyser-fyp")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for file, lines := range result {
		if len(lines) > 0 {
			fmt.Printf("File: %s (total %d lines)\n", file, len(lines))
		}
	}

	// Save the map after traversal
	if err := SaveFilenameMap(); err != nil {
		fmt.Println("Error saving:", err)
	}
}
