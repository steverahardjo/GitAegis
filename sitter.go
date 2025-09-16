package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"
	// example import for Go grammar
	"github.com/smacker/go-tree-sitter/golang"
)

// GrammarConfig defines the structure of sitter_filecase.json
type GrammarConfig struct {
	Extensions map[string]string `json:"extensions"`
	Filenames  map[string]string `json:"filenames"`
}

// --------------------
// Globals
// --------------------

// grammars must be preloaded here
var grammars = map[string]*sitter.Language{
	"go": golang.GetLanguage(), // preload Go grammar
}

// --------------------
// Grammar Loader
// --------------------

// loadExtMap loads extension/filename mappings from a JSON config
func loadExtMap(path string) (*GrammarConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg GrammarConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// initGrammar selects the grammar for a given filename
func initGrammar(filename string) *sitter.Language {
	jsonPath := "sitter.json"
	cfg, err := loadExtMap(jsonPath)
	if err != nil {
		fmt.Printf("Error loading extension map from %s: %v\n", jsonPath, err)
		return nil
	}

	suffix := filepath.Ext(filename) // includes the dot
	base := filepath.Base(filename)

	var langName string
	var ok bool

	if suffix != "" {
		langName, ok = cfg.Extensions[suffix]
	}
	if !ok { // try special filenames
		langName, ok = cfg.Filenames[base]
	}

	if !ok {
		fmt.Printf("No grammar mapping for file: %s (ext=%s)\n", filename, suffix)
		return nil
	}

	lang, ok := grammars[langName]
	if !ok {
		fmt.Printf("Grammar %s not loaded (filename=%s)\n", langName, filename)
		return nil
	}

	return lang
}

// --------------------
// Example usage
// --------------------

func main() {
	// Example JSON config (sitter_filecase.json):
	// {
	//   "extensions": {".go": "go"},
	//   "filenames": {}
	// }

	filename := "example.go"
	lang := initGrammar(filename)
	if lang == nil {
		fmt.Println("Failed to initialize grammar.")
		return
	}
	fmt.Println("Grammar loaded successfully for:", filename)
}
