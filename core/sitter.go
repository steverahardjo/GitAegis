package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	sitter "github.com/smacker/go-tree-sitter"
)

// GrammarConfig defines the structure of sitter.json
type GrammarConfig struct {
	Extensions map[string]string `json:"extensions"`
	Filenames  map[string]string `json:"filenames"`
}

var (
	sitterInit    sync.Once
	sitterInitErr error
	SitterMap     *GrammarConfig
	sitter_path   string
	langCache     sync.Map
)

func IntegrateTreeSitter(homePath string) error {
	sitterInit.Do(func() {
		if homePath == "" {
			sitterInitErr = errors.New("tree-sitter path cannot be empty")
			return
		}

		// Clean and make absolute
		absPath, err := filepath.Abs(filepath.Clean(homePath))
		if err != nil {
			sitterInitErr = err
			return
		}

		// Check if the path exists and is a directory
		info, err := os.Stat(absPath)
		if err != nil {
			sitterInitErr = errors.New("tree-sitter path does not exist")
			return
		}
		if !info.IsDir() {
			sitterInitErr = errors.New("tree-sitter path must be a directory")
			return
		}
		sitter_path = absPath
		SitterMap = &GrammarConfig{}
	})

	return sitterInitErr
}

// loadExtMap fetches and unmarshals the sitter.json file
func loadExtMap() (*GrammarConfig, error) {
	resp, err := http.Get("https://raw.githubusercontent.com/steverahardjo/gitaegis/cli-enabled/core/sitter.json")
	if err != nil {
		return nil, fmt.Errorf("[TreeSitter] failed to read sitter.json: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cfg GrammarConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, errors.New("[TreeSitter] unable to laod json helper")
	}
	return &cfg, nil
}

// getExtMap lazily loads sitter.json once and caches it
func getExtMap() (*GrammarConfig, error) {
	sitterInit.Do(func() {
		SitterMap, sitterInitErr = loadExtMap()
	})
	return SitterMap, sitterInitErr
}

// initGrammar initializes and returns a Tree-sitter parser for a given file.
func initGrammar(filename string) *sitter.Parser {
	cfg, err := getExtMap()
	if err != nil {
		fmt.Println("[TreeSitter] Error loading config:", err)
		return nil
	}

	ext := filepath.Ext(filename)
	langFile, ok := cfg.Extensions[ext]
	if !ok {
		base := filepath.Base(filename)
		langFile, ok = cfg.Filenames[base]
	}
	if !ok {
		return nil
	}

	// --- Cache only the *language* ---
	val, ok := langCache.Load(langFile)
	var lang *sitter.Language
	if ok {
		lang = val.(*sitter.Language)
	} else {
		langBase := strings.TrimSuffix(langFile, ".so")
		soPath := filepath.Join(sitter_path, langFile)

		lib, dlErr := purego.Dlopen(soPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if dlErr != nil {
			fmt.Printf("[TreeSitter] Error loading grammar %s: %v\n", soPath, dlErr)
			return nil
		}

		var sym func() uintptr
		symbolName := "tree_sitter_" + langBase
		purego.RegisterLibFunc(&sym, lib, symbolName)

		lang = sitter.NewLanguage(unsafe.Pointer(sym()))
		if lang == nil {
			fmt.Println("[TreeSitter] Error creating language for:", langBase)
			return nil
		}
		langCache.Store(langFile, lang)
	}

	// --- Always create a new parser ---
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	return parser
}

// createTree parses a file and returns both the syntax tree and file content
func CreateTree(filename string) (*sitter.Tree, []byte, error) {
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

// Run a DFS to walk through the tree and get leaf node
func walkParse(root *sitter.Node, filter LineFilter, code []byte) *CodeLine {
	var lines []string
	var indexes, columns []int
	var extracted []Payload

	stack := []*sitter.Node{root}

	for len(stack) > 0 {
		// Pop from stack
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// If leaf node, process it
		if n.ChildCount() == 0 {
			content := n.Content(code)
			if pl, ok := filter(content); ok {
				start := n.StartPoint()
				lines = append(lines, content)
				indexes = append(indexes, int(start.Row)+1)
				columns = append(columns, int(start.Column)+1)
				extracted = append(extracted, pl)
			}
			continue
		}

		// Push children onto stack in reverse order for DFS
		for i := int(n.ChildCount()) - 1; i >= 0; i-- {
			stack = append(stack, n.Child(i))
		}
	}

	return &CodeLine{
		Lines:     lines,
		Indexes:   indexes,
		Columns:   columns,
		Extracted: extracted,
	}
}
