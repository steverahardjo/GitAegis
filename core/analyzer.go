package core

import (
	"fmt"
	"os"
	"path/filepath"
	"log"
	"sync"
	"runtime"

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
func (res *ScanResult)IsFilenameMapEmpty() bool {
	if res.filenameMap == nil {
		return true
	}
	return false
}

func isExempt(filename string) bool {
	for _, ex := range Exempt {
		if filepath.Base(filename) == ex {
			return true
		}
	}
	return false
}

// Main folder walker (parallelized) {public}
func (res *ScanResult) IterFolder(root string, filter LineFilter) (error) {
	ign := initGitIgnore()

	// Collect all file paths first
	var files []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ignoreFiles(p, ign) || isExempt(p) {
			return nil
		}
		if !d.IsDir() {
			files = append(files, p)
		}
		return nil
	})

	// Worker pool for parallel parsing
	numWorkers := runtime.NumCPU()
	fileCh := make(chan string, len(files))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++{
		wg.Add(1)
		go func(){
			defer wg.Done()
			for filename := range fileCh {
				tree, code, err := createTree(filename)
				if err != nil {
					log.Printf("Error parsing %s: %v", filename, err)
					continue
				}
				lines := walkParse(tree.RootNode(), filter, code)
				if len(lines) > 0 {
					res.mutex.Lock()
					res.filenameMap[filename] = lines
					res.mutex.Unlock()
				}
			}
		}()
	}
	close(fileCh)
	wg.Wait()
	return err
}



// Private function to check ignores
func ignoreFiles(path string, ign *gitignore.GitIgnore) bool {
	return ign.MatchesPath(path)
}

func (res *ScanResult) Get_filenameMap() map[string][]CodeLine {
	return res.filenameMap
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
