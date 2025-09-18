package core

import (
	"fmt"
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

var filenameMap = make(map[string][]CodeLine)

// language specific exemption
var exempt = []string{"uv.lock", "pyproject.toml", "pnpm-lock.yaml", "package-lock.json", "yarn.lock", "go.sum", "deno.lock", "Cargo.lock"}

// Load .gitignore once
func initGitIgnore() *gitignore.GitIgnore {
	ign, err := gitignore.CompileIgnoreFile(".gitignore")
	if err != nil {
		if os.IsNotExist(err) {
			// no .gitignore: create empty matcher
			return gitignore.CompileIgnoreLines()
		}
		panic(err)
	}
	return ign
}

func isExempt(filename string) bool {
	for _, ex := range exempt {
		if filepath.Base(filename) == ex {
			return true
		}
	}
	return false
}

// Private function to check ignores
func ignoreFiles(path string, ign *gitignore.GitIgnore) bool {
	return ign.MatchesPath(path)
}

// Main folder walker
func IterFolder(root string, filter LineFilter) (map[string][]CodeLine, error) {
	ign := initGitIgnore()

	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip ignored files
		if ignoreFiles(p, ign) || isExempt(p) {
			return nil
		}

		if !d.IsDir() {
			// Parse file with tree-sitter
			tree, code, parseErr := createTree(p)
			if parseErr != nil {
				fmt.Printf("Skipping %s (parse error: %v)\n", p, parseErr)
				return nil
			}
			defer tree.Close()

			// Walk AST and collect CodeLine results
			rootNode := tree.RootNode()
			results := walkParse(rootNode, filter, code)

			// Save results in global map
			if len(results) > 0 {
				filenameMap[p] = results
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking folder: %w", err)
	}

	return filenameMap, nil
}

func PrettyPrintResults(results map[string][]CodeLine) {
	red := "\033[31m"
	green := "\033[32m"
	yellow := "\033[33m"
	reset := "\033[0m"

	fmt.Println(yellow + "GITAEGIS DETECTED THE FOLLOWING SECRETS\n===============================" + reset)
	for filename, lines := range results {
		fmt.Println(green + "File: " + filename + reset)
		for _, line := range lines {
			fmt.Printf("%s|\n Index: %d\n Line: %s%s\n", red, line.Index, line.Line, reset)
		}
		fmt.Println(green + "------------------------------" + reset)
	}
}
