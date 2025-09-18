package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unsafe"

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

// initGrammar initializes and returns a Tree-sitter parser for a given file.
func initGrammar(filename string) *sitter.Parser {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home dir:", err)
		return nil
	}

	// load grammar config
	cfg, err := loadExtMap("sitter.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return nil
	}

	root := ".private/helix/runtime/grammars"

	parser := sitter.NewParser()

	// lookup by extension
	ext := filepath.Ext(filename)
	langname, ok := cfg.Extensions[ext]
	if !ok {
		// fallback: check filenames
		base := filepath.Base(filename)
		langname, ok = cfg.Filenames[base]
	}
	if !ok {
		fmt.Println("No grammar found for:", filename)
		return nil
	}

	// ensure langname doesn’t include ".so"
	langBase := strings.TrimSuffix(langname, ".so")

	// build .so path and load
	soPath := filepath.Join(homeDir, root, langBase+".so")
	lib, dlErr := purego.Dlopen(soPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if dlErr != nil {
		fmt.Println("Error loading grammar:", dlErr)
		return nil
	}

	// bind tree_sitter_<langname> symbol
	var sym func() uintptr
	symbolName := "tree_sitter_" + langBase
	purego.RegisterLibFunc(&sym, lib, symbolName)

	// construct Language
	lang := sitter.NewLanguage(unsafe.Pointer(sym()))
	if lang == nil {
		fmt.Println("Error creating language")
		return nil
	}

	parser.SetLanguage(lang)
	return parser
}

// createTree parses a file and returns both the syntax tree and file content
func createTree(filename string) (*sitter.Tree, []byte, error) {
	parser := initGrammar(filename)
	if parser == nil {
		return nil, nil, fmt.Errorf("failed to initialize grammar")
	}

	// Read full file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	// Parse the whole file content
	tree, err := parser.ParseCtx(context.Background(), nil, data)
	if err != nil {
		return nil, nil, err
	}

	return tree, data, nil
}

// walkParse recursively walks the AST without depth limit
var apiKeyRegex = regexp.MustCompile(`[a-zA-Z0-9_.+/~$-][a-zA-Z0-9_.+/~$=!%:-]{10,1000}[a-zA-Z0-9_.+/=~$!%-]`)

// walkParse recursively walks the AST without depth limit
func walkParse(node *sitter.Node, code []byte) []CodeLine {
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

		// --- Apply filters ---
		if len(content) > 0 && len(content) <= 2048 { // limit content length
			if calcEntropy(content) > 5.0 && apiKeyRegex.MatchString(content) {
				start := child.StartPoint()
				results = append(results, CodeLine{
					Line:   content,               // snippet (string literal / identifier)
					Index:  int(start.Row) + 1,    // 1-based line number
					Column: int(start.Column) + 1, // 1-based column number
				})
			}
		}

		// Recurse
		results = append(results, walkParse(child, code)...)
	}

	return results
}

// --------------------
// Helpers
// --------------------

// safePreview truncates long strings for printing
func safePreview(s string) string {
	if len(s) > 50 {
		return s[:47] + "..."
	}
	return s
}

func filter_sitter(filename string) {
	tree, code, err := createTree(filename)
	if err != nil {
		fmt.Println("Failed to parse file:", err)
		return
	}
	defer tree.Close()

	fmt.Printf("File parsed successfully: %s\n", filename)
	// Get root node of AST
	rootNode := tree.RootNode()
	fmt.Printf("Root node type: %s\n", rootNode.Type())
	// Walk the tree
	fmt.Println("Walking AST...")
	walkParse(rootNode, code)

}
