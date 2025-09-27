package core

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

// SaveFilenameMap saves the filename map (results from a scan) into a binary file.
// The map stores file paths with their detected CodeLine entries.
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

// LoadFilenameMap loads a previously saved filename map from disk.
// If no file exists, it returns an empty map and no error.
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

// UpdateGitignore appends all saved file paths from the filename map into .gitignore.
// It ensures that previously detected files are ignored in Git.
func UpdateGitignore() error {
	blob, err := LoadFilenameMap(".")
	if err != nil {
		return fmt.Errorf("unable to load filename map: %w", err)
	}

	if _, err := os.Stat(".gitignore"); os.IsNotExist(err) {
		fmt.Println(".gitignore does not exist, creating it...")
	}

	ignoreFile, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open .gitignore file: %w", err)
	}
	defer ignoreFile.Close()

	for filename := range blob {
		if _, err := ignoreFile.WriteString(filename + "\n"); err != nil {
			return fmt.Errorf("failed to write to .gitignore: %w", err)
		}
		fmt.Printf("Added %s to .gitignore\n", filename)
	}

	fmt.Println("Updated .gitignore successfully.")
	return nil
}

// obfuscatePerLine replaces a specific line in a file with a warning marker.
// This is used to "mask" secrets without deleting the file entirely.
func obfuscatePerLine(filename string, line CodeLine) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	buf := strings.Split(string(content), "\n")
	if line.Index < 0 || line.Index >= len(buf) {
		return fmt.Errorf("line index %d out of range for file %s", line.Index, filename)
	}

	buf[line.Index] = "===== GitAegis detected a secret: OBFUSCATED ====="
	output := strings.Join(buf, "\n")
	return os.WriteFile(filename, []byte(output), 0644)
}

// LoadObfuscation obfuscates all lines previously detected and saved in the filename map.
// This will overwrite files in-place with warnings on secret lines.
func LoadObfuscation(root string) error {
	x, err := LoadFilenameMap(root)
	if err != nil {
		return fmt.Errorf("unable to load filename map: %w", err)
	}
	for k, v := range x {
		for _, line := range v {
			if err := obfuscatePerLine(k, line); err != nil {
				return fmt.Errorf("failed to obfuscate file %s line %d: %v", k, line.Index, err)
			}
		}
	}
	return nil
}

// undoObfuscate restores all obfuscated lines back to their original content
// using the saved filename map.
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