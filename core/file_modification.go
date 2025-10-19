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
	Data map[string]CodeLine   `json:"data"`
}

func (res *ScanResult)SaveFilenameMap(root string) error {
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
	for _, res := range res.filenameMap {
		total += len(res.Indexes)
	}

	// Fill in top-level struct
	topLevel := TopLevel{
		Meta: JsonMetadata{
			Timestamp: time.Now().Format(time.RFC3339),
			Author:    author,
			Frequency: total,
		},
		Data: res.filenameMap,
	}

	// Marshal with indentation
	data, err := json.MarshalIndent(topLevel, "", "  ")
	if err != nil {
		
		log.Printf("Unable to save into a jsonl for saveFilenamemap")
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		log.Fatal(err)
		return err
	}
	f.Sync()
	
	checkAddGitignore(root, ".gitaegis.jsonl")
	return err
}

func LoadFilenameMap(root string) (map[string]CodeLine, error) {
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

func checkAddGitignore(root string, filename string) error {
    ignorePath := filepath.Join(root, ".gitignore")
    var lines []string
    if data, err := os.ReadFile(ignorePath); err == nil {
        lines = strings.Split(string(data), "\n")
        for _, line := range lines {
            if strings.TrimSpace(line) == filename {
                return nil
            }
        }
    }

    f, err := os.OpenFile(ignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()

    if _, err := f.WriteString(filename + "\n"); err != nil {
        return err
    }

    return nil
}



// UpdateGitignore appends all saved file paths from the filename map into .gitignore.
// It ensures that previously detected files are ignored in Git.
func UpdateGitignore(blob map[string]CodeLine) error {

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
/*

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
	*/
