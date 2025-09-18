package core

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

// SaveFilenameMap saves the current run hash map into a binary file
func SaveFilenameMap(root string, filenameMap map[string][]CodeLine) error {
	filePath := path.Join(root, ".gitaegis")
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	gob.Register(CodeLine{})
	return enc.Encode(filenameMap)
}

// LoadFilenameMap loads the previous run hash map from binary
func LoadFilenameMap(root string) (map[string][]CodeLine, error) {
	filePath := path.Join(root, ".gitaegis")
	f, err := os.Open(filePath)
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

	return filenameMap, nil
}

// obfuscate replaces the specified line with a warning
func obfuscate(filename string, line CodeLine) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	buf := strings.Split(string(content), "\n")
	if line.Index < 0 || line.Index >= len(buf) {
		return fmt.Errorf("line index %d out of range for file %s", line.Index, filename)
	}

	buf[line.Index] = "=====GitAegis Detect a Secret, OBFUSCATED===="
	output := strings.Join(buf, "\n")
	return os.WriteFile(filename, []byte(output), 0644)
}

// undoObfuscate restores all obfuscated lines from the saved map
func undoObfuscate(root string) error {
	blob, err := LoadFilenameMap(root)
	if err != nil {
		return fmt.Errorf("unable to load filename map: %w", err)
	}

	for filename, lines := range blob {
		content, err := os.ReadFile(filename)
		if err != nil {
			log.Printf("failed to read file %s: %v", filename, err)
			continue
		}

		buf := strings.Split(string(content), "\n")
		for _, line := range lines {
			if line.Index >= 0 && line.Index < len(buf) {
				buf[line.Index] = line.Line
			}
		}

		output := strings.Join(buf, "\n")
		if err := os.WriteFile(filename, []byte(output), 0644); err != nil {
			log.Printf("failed to write file %s: %v", filename, err)
		}
	}

	return nil
}

