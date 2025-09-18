package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"
	"strings"
	"github.com/ebitengine/purego"
	sitter "github.com/smacker/go-tree-sitter"
)

// GrammarConfig defines the structure of sitter.json
type GrammarConfig struct {
	Extensions map[string]string `json:"extensions"`
	Filenames  map[string]string `json:"filenames"`
}

// --------------------
// Grammar Loader
// --------------------

// loadExtMap loads extension/filename mappings from a JSON config
func loadExtMap() (*GrammarConfig, error) {
	configPath := "core/sitter.json"

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sitter.json: %w", err)
	}

	var cfg GrammarConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid sitter.json: %w", err)
	}
	return &cfg, nil
}

// platformExt returns the correct shared object extension for the OS
func platformExt() string {
	switch runtime.GOOS {
	case "darwin":
		return ".dylib"
	case "windows":
		return ".dll"
	default:
		return ".so"
	}
}

// initGrammar initializes and returns a Tree-sitter parser for a given file.
func initGrammar(filename string) *sitter.Parser {
	cfg, err := loadExtMap()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return nil
	}

	// Always resolve grammar dir from HOME to avoid relative path issues
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error resolving home directory:", err)
		return nil
	}
	grammarDir := filepath.Join(home, ".private", "helix", "runtime", "grammars")

	parser := sitter.NewParser()

	// lookup by extension
	ext := filepath.Ext(filename)
	langFile, ok := cfg.Extensions[ext]
	if !ok {
		// fallback: check filenames
		base := filepath.Base(filename)
		langFile, ok = cfg.Filenames[base]
	}
	if !ok {
		fmt.Println("No grammar found for:", filename)
		return nil
	}

	// Strip .so to get clean base name for symbol lookup
	langBase := strings.TrimSuffix(langFile, ".so")

	// Build full .so path
	soPath := filepath.Join(grammarDir, langFile)

	lib, dlErr := purego.Dlopen(soPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if dlErr != nil {
		fmt.Printf("Error loading grammar %s: %v\n", soPath, dlErr)
		return nil
	}

	// Bind tree_sitter_<langname> symbol
	var sym func() uintptr
	symbolName := "tree_sitter_" + langBase
	purego.RegisterLibFunc(&sym, lib, symbolName)

	// Construct Language
	lang := sitter.NewLanguage(unsafe.Pointer(sym()))
	if lang == nil {
		fmt.Println("Error creating language for:", langBase)
		return nil
	}

	parser.SetLanguage(lang)
	return parser
}

// --------------------
// File parsing
// --------------------

// createTree parses a file and returns both the syntax tree and file content
func createTree(filename string) (*sitter.Tree, []byte, error) {
	parser := initGrammar(filename)
	if parser == nil {
		return nil, nil, fmt.Errorf("failed to initialize grammar")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	tree, err := parser.ParseCtx(context.Background(), nil, data)
	if err != nil {
		return nil, nil, err
	}

	return tree, data, nil
}

// walkParse recursively walks the AST without depth limit
func walkParse(node *sitter.Node, filter LineFilter, code []byte) []CodeLine {
	results := []CodeLine{}
	if node == nil {
		return results
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}

		content := child.Content(code)

		if len(content) > 0 && len(content) <= 2048 && child.Type() == "identifier" {
			if filter(content) {
				start := child.StartPoint()
				results = append(results, CodeLine{
					Line:   content,
					Index:  int(start.Row) + 1,
					Column: int(start.Column) + 1,
				})
			}
		}
		// Recurse
		results = append(results, walkParse(child, filter, code)...)
	}

	return results
}

// safePreview truncates long strings for printing
func safePreview(s string) string {
	if len(s) > 50 {
		return s[:47] + "..."
	}
	return s
}
