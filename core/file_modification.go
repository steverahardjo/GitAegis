package core

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"time"
	"fmt"
	"strings"
	"os/user"
	"path/filepath"
)

type JsonMetadata struct {
	Timestamp string `json:"timestamp"`
	Author    string `json:"author"`
	Frequency int    `json:"freq"`
}

// Global struct for saving/loading
type TopLevel struct {
	Meta JsonMetadata            `json:"meta"`
	Data map[string][]CodeLine   `json:"data"`
}

func SaveFilenameMap(root string, filenameMap map[string][]CodeLine) error {
	filePath := filepath.Join(root, ".gitaegis.jsonl")

	u, err := user.Current()
	if err != nil {
		return err
	}
	author := u.Username

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Count total CodeLine entries
	total := 0
	for _, lines := range filenameMap {
		total += len(lines)
	}

	// Fill in top-level struct
	topLevel := TopLevel{
		Meta: JsonMetadata{
			Timestamp: time.Now().Format(time.RFC3339),
			Author:    author,
			Frequency: total,
		},
		Data: filenameMap,
	}

	// Marshal with indentation
	data, err := json.MarshalIndent(topLevel, "", "  ")
	if err != nil {
		log.Fatal(err)
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		log.Fatal(err)
		return err
	}
	f.Sync()
	
	CheckAddGitignore(root, ".gitaegis.jsonl")
	return err
}

func LoadFilenameMap(root string) (map[string][]CodeLine, error) {
	filePath := path.Join(root, ".gitaegis.jsonl")

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var topLevel TopLevel
	if err := json.NewDecoder(f).Decode(&topLevel); err != nil {
		return nil, err
	}
	return topLevel.Data, nil
}

func CheckAddGitignore(root string, filename string) error {
    ignorePath := filepath.Join(root, ".gitignore")

    // Read existing .gitignore if it exists
    var lines []string
    if data, err := os.ReadFile(ignorePath); err == nil {
        lines = strings.Split(string(data), "\n")
        for _, line := range lines {
            if strings.TrimSpace(line) == filename {
                // already present
                return nil
            }
        }
    }

    // Open (or create) .gitignore for appending
    f, err := os.OpenFile(ignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()

    // Append with newline
    if _, err := f.WriteString(filename + "\n"); err != nil {
        return err
    }

    return nil
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
	ignoreFile.Sync()
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

// obfuscates all lines previously detected and saved in the filename map.
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
