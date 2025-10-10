package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"strings"
	gitignore "github.com/sabhiram/go-gitignore"
)

type ScanResult struct {
	filenameMap map[string][]CodeLine
	mutex       sync.Mutex
	exempt      map[string]struct{} // use set instead of slice
}

// language specific exemption (defaults)
var DefaultExempt = []string{
	"uv.lock", "pyproject.toml", "pnpm-lock.yaml", "package-lock.json",
	"yarn.lock", "go.sum", "deno.lock", "Cargo.lock",
	".gitignore", ".python-version", "LICENSE", ".gitaegis.jsonl",
}

func (res *ScanResult) Init() {
	res.filenameMap = make(map[string][]CodeLine)
	res.exempt = make(map[string]struct{})
	for _, f := range DefaultExempt {
		res.exempt[f] = struct{}{}
	}
}

// AddExempt adds a new exempt file to the set.
func (res *ScanResult) AddExempt(file string) {
	if _, exists := res.exempt[file]; exists {
		fmt.Println("File is already exempted.")
		return
	}
	res.exempt[file] = struct{}{}
}

// Load .gitignore once
func initGitIgnore() *gitignore.GitIgnore {
	ign, err := gitignore.CompileIgnoreFile(".gitignore")
	if err != nil {
		if os.IsNotExist(err) {
			return gitignore.CompileIgnoreLines()
		}
		log.Fatalf("Error loading .gitignore: %v", err)
	}
	return ign
}

// IsFilenameMapEmpty returns true if no results are stored
func (res *ScanResult) IsFilenameMapEmpty() bool {
	return len(res.filenameMap) == 0
}

func (res *ScanResult) isExempt(filename string) bool {
	_, ok := res.exempt[filepath.Base(filename)]
	return ok
}

func isExecutable(filename string) bool {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return false
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return false
	}
	// Check if any execute bit is set (owner, group, or other)
	return info.Mode().Perm()&0111 != 0
}
// IterFolder walks a directory and processes files in parallel
func (res *ScanResult) IterFolder(root string, filter LineFilter, is_gitignore bool, max_filesize int64) error {
	var ign *gitignore.GitIgnore
	if is_gitignore == true{
		ign = initGitIgnore()
	}

	// Collect files first
	var files []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if ignoreFiles(p, ign) || res.isExempt(p) || isExecutable(p){
			return nil
		}
		info, err := os.Stat(p)
		if err != nil{
			log.Fatal("Unable to get file size inside of core.IterFolder")
		}
		if info.Size() > max_filesize{
			return nil
		}

		files = append(files, p)
		return nil
	})
	if err != nil {
		return err
	}

	// Worker pool
	numWorkers := runtime.NumCPU()
	fileCh := make(chan string, len(files))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filename := range fileCh {
				tree, code, err := createTree(filename)
				if err != nil {
					res.PerLineScan(filename, filter)
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

	for _, f := range files {
		fileCh <- f
	}
	close(fileCh)
	wg.Wait()

	return nil
}

func (res *ScanResult) ClearMap() {
	res.filenameMap = make(map[string][]CodeLine)
}

func ignoreFiles(path string, ign *gitignore.GitIgnore) bool {
	return ign.MatchesPath(path)
}

func (res *ScanResult) GetFilenameMap() map[string][]CodeLine {
	return res.filenameMap
}
///fallback if we can't use treesitter
func (res *ScanResult) PerLineScan(filename string, filter LineFilter) []CodeLine {
    data, err := os.ReadFile(filename)
    if err != nil {
        log.Fatal(err)
    }

    lines := strings.Split(string(data), "\n")
    var matched []CodeLine

    for i, line := range lines {
        // Split line by spaces
        sublines := strings.Fields(line)

        for _, sub := range sublines {
			pl, dec := filter(sub)
            if dec {
                matched = append(matched, CodeLine{
                    Line:    sub,
                    Index:   i + 1,
					Column: i+1,
					Extracted: pl,
                })
            }
        }
    }

    return matched
}


// PrettyPrintResults prints results nicely with colors
func (res *ScanResult) PrettyPrintResults() {
	const (
		red    = "\033[31m"
		green  = "\033[32m"
		yellow = "\033[33m"
		reset  = "\033[0m"
	)

	fmt.Printf("%sGITAEGIS DETECTED THE FOLLOWING SECRETS%s\n", yellow, reset)
	fmt.Printf("%s=======================================%s\n", yellow, reset)

	for filename, lines := range res.filenameMap {
		fmt.Printf("\n%sFile:%s %s\n", green, reset, filename)

		for _, line := range lines {
			fmt.Printf("%s---------------------------------------%s\n", yellow, reset)
			fmt.Printf("%sIndex:%s   %d\n", red, reset, line.Index)
			fmt.Printf("%sLine:%s    %s\n", red, reset, line.Line)
			for k, v := range line.Extracted{
				fmt.Printf("%s%s:%s %.4f\n", red, k, reset, v)
			}
		}
	}
}
