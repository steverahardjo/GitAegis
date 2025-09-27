package core

import (
	"fmt"
	"os"
	"path/filepath"
	"log"
	"sync"

	gitignore "github.com/sabhiram/go-gitignore"
)

type ScanResult struct{
	filenameMap map[string][]CodeLine
	mutex sync.Mutex
	exempt []string
}

// language specific exemption
var Exempt = []string{"uv.lock", "pyproject.toml", "pnpm-lock.yaml", "package-lock.json", "yarn.lock", "go.sum", "deno.lock", "Cargo.lock", ".gitignore", ".python-version"}

func (res *ScanResult) Init() {
	res.filenameMap = make(map[string][]CodeLine)
	res.exempt = Exempt
}

func (res *ScanResult)AddExempt(file string) {
	for _, f := range Exempt {
		if f == file {
			fmt.Println("File is already exempted.")
		}
	}
	Exempt = append(res.exempt, file)
}


// Load .gitignore once {private}
func initGitIgnore() *gitignore.GitIgnore {
	ign, err := gitignore.CompileIgnoreFile(".gitignore")
	if err != nil {
		if os.IsNotExist(err) {
			// no .gitignore: create empty matcher
			return gitignore.CompileIgnoreLines()
		}
		log.Fatalf("Error loading .gitignore: %v", err)
	}
	return ign
}

//filenameMap check if emptu {public}
func IsFilenameMapEmpty(m map[string][]CodeLine) bool {
	return len(m) == 0
}

func isExempt(filename string) bool {
	for _, ex := range Exempt {
		if filepath.Base(filename) == ex {
			return true
		}
	}
	return false
}

// Main folder walker
func (res  *ScanResult)IterFolder(root string, filter LineFilter) (map[string][]CodeLine, error) {
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
				res.filenameMap[p] = results
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking folder: %w", err)
	}

	return res.filenameMap, nil
}


// Private function to check ignores
func ignoreFiles(path string, ign *gitignore.GitIgnore) bool {
	return ign.MatchesPath(path)
}


func (res *ScanResult)  PrettyPrintResults() {
	red := "\033[31m"
	green := "\033[32m"
	yellow := "\033[33m"
	reset := "\033[0m"

	fmt.Println(yellow + "GITAEGIS DETECTED THE FOLLOWING SECRETS\n===============================" + reset)
	for filename, lines := range res.filenameMap {
		fmt.Println(green + "File: " + filename + reset)
		if len(lines) <= 0 {
			continue
		}
		for _, line := range lines {
			fmt.Printf("%s \t |\n Index: %d\n Line: %s%s\n", red, line.Index, line.Line, reset)
		}
		fmt.Println(green + "------------------------------" + reset)
	}
}
