package core

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	gitignore "github.com/sabhiram/go-gitignore"
)

type CodeLine struct {
	Lines     []string
	Indexes   []int
	Columns   []int
	Extracted []Payload
}

type ScanResult struct {
	filenameMap map[string][]CodeLine
	mutex       sync.RWMutex
	exempt      map[string]struct{}
}

var DefaultExempt = []string{
	"uv.lock", "pyproject.toml", "pnpm-lock.yaml", "package-lock.json",
	"yarn.lock", "go.sum", "deno.lock", "Cargo.lock",
	".gitignore", ".python-version", "LICENSE", ".gitaegis.jsonl",
}

func (res *ScanResult) Init() {
	res.filenameMap = make(map[string][]CodeLine)
	res.exempt = make(map[string]struct{}, len(DefaultExempt))
	for _, f := range DefaultExempt {
		res.exempt[f] = struct{}{}
	}
}

func (res *ScanResult) AddExempt(file string) {
	res.mutex.Lock()
	defer res.mutex.Unlock()
	res.exempt[file] = struct{}{}
}

func initGitIgnore() *gitignore.GitIgnore {
	ign, err := gitignore.CompileIgnoreFile(".gitignore")
	if err != nil {
		if os.IsNotExist(err) {
			return gitignore.CompileIgnoreLines()
		}
		log.Printf("Warning: failed to load .gitignore (%v)", err)
		return gitignore.CompileIgnoreLines()
	}
	return ign
}

func (res *ScanResult) isExempt(filename string) bool {
	_, ok := res.exempt[filepath.Base(filename)]
	return ok
}

func isExecutable(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return info.Mode().Perm()&0111 != 0
}

func ignoreFiles(path string, ign *gitignore.GitIgnore) bool {
	return ign != nil && ign.MatchesPath(path)
}

func (res *ScanResult) IsFilenameMapEmpty() bool {
	res.mutex.RLock()
	defer res.mutex.RUnlock()
	return len(res.filenameMap) == 0
}

func (res *ScanResult) ClearMap() {
	res.mutex.Lock()
	defer res.mutex.Unlock()
	res.filenameMap = make(map[string][]CodeLine)
}

func (res *ScanResult) GetFilenameMap() map[string][]CodeLine {
	res.mutex.RLock()
	defer res.mutex.RUnlock()
	cpy := make(map[string][]CodeLine, len(res.filenameMap))
	for k, v := range res.filenameMap {
		cpy[k] = v
	}
	return cpy
}

// IterFolder walks through files concurrently and scans them
func (res *ScanResult) IterFolder(root string, filter LineFilter, useGitIgnore bool, maxFileSize int64) error {
	var ign *gitignore.GitIgnore
	if useGitIgnore {
		ign = initGitIgnore()
	}

	files := make([]string, 0, 512)
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if ignoreFiles(p, ign) || res.isExempt(p) || isExecutable(p) {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() > maxFileSize {
			return nil
		}
		files = append(files, p)
		return nil
	})
	if err != nil {
		return err
	}

	numWorkers := runtime.NumCPU()
	fileCh := make(chan string, numWorkers*2)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filename := range fileCh {
				tree, code, err := createTree(filename)
				if err != nil {
					lines := res.PerLineScan(filename, filter)
					if len(lines) > 0 {
						res.mutex.Lock()
						res.filenameMap[filename] = lines
						res.mutex.Unlock()
					}
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

// Fallback scan when tree-sitter not available
func (res *ScanResult) PerLineScan(filename string, filter LineFilter) []CodeLine {
	f, err := os.Open(filename)
	if err != nil {
		log.Printf("Error reading file %s: %v", filename, err)
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	var indexes, columns []int
	var extracted []Payload
	lineNum := 1

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		for col, token := range parts {
			pl, ok := filter(token)
			if ok {
				lines = append(lines, token)
				indexes = append(indexes, lineNum)
				columns = append(columns, col+1)
				extracted = append(extracted, pl)
			}
		}
		lineNum++
	}

	if len(lines) == 0 {
		return nil
	}

	return []CodeLine{{
		Lines:     lines,
		Indexes:   indexes,
		Columns:   columns,
		Extracted: extracted,
	}}
}

// PrettyPrintResults prints results nicely with ANSI colors
func (res *ScanResult) PrettyPrintResults() {
	const (
		red    = "\033[31m"
		green  = "\033[32m"
		yellow = "\033[33m"
		reset  = "\033[0m"
	)

	if len(res.filenameMap) == 0 {
		fmt.Println("No secrets detected.")
		return
	}

	var b strings.Builder
	b.Grow(4096) // Preallocate a reasonable buffer to reduce reallocations

	b.WriteString(yellow)
	b.WriteString("GITAEGIS DETECTED THE FOLLOWING SECRETS\n")
	b.WriteString("=======================================\n")
	b.WriteString(reset)

	for filename, lines := range res.filenameMap {
		b.WriteString("\n")
		b.WriteString(green)
		b.WriteString("File: ")
		b.WriteString(reset)
		b.WriteString(filename)
		b.WriteByte('\n')

		for _, line := range lines {
			for i, l := range line.Lines {
				b.WriteString(yellow)
				b.WriteString("---------------------------------------\n")
				b.WriteString(reset)

				// Print line info
				fmt.Fprintf(&b, "%sLine %d (Col %d):%s %s\n",
					red, line.Indexes[i], line.Columns[i], reset, l)

				// Print extracted payload
				for k, v := range line.Extracted[i] {
					fmt.Fprintf(&b, "  %s%s:%s %.4f\n", red, k, reset, v)
				}
			}
		}
	}

	os.Stdout.WriteString(b.String())
}
