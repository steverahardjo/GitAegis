package main

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

var filters = allFilters(
	entropyFilter(4.8),
)

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
func iterFolder(root string) (map[string][]CodeLine, error) {
	ign := initGitIgnore()

	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if ignoreFiles(p, ign) {
			return nil
		}

		if isExempt(p) {
			return nil
		}

		if !d.IsDir() {
			lines, err := readAndCalc(p, filters)
			filenameMap[p] = lines

			if err != nil {
				panic("something is wrong while going through files")
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking folder: %w", err)
	}
	return filenameMap, nil
}
