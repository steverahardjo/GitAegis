package main

import (
	"fmt"
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

var filenameMap = make(map[string][]CodeLine)

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
	entropyFilter(5.0),
)

// Private function to check ignores
func ignoreFiles(path string, ign *gitignore.GitIgnore) bool {
	if ign == nil {
		return false
	}
	return ign.MatchesPath(path)
}

// Main folder walker
func iterFolder(path string) (map[string][]CodeLine, error) {
	ign := initGitIgnore()

	err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if ignoreFiles(p, ign) {
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

func main() {
	result, err := iterFolder(".")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for file, lines := range result {
		fmt.Printf("File: %s (total %d lines)\n", file, len(lines))
		for _, l := range lines {
			fmt.Printf("  %d: %s\n", l.index, l.line)
		}
	}
}
