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

// CodeLine stores matching lines in a file
type CodeLine struct {
	Lines     []string
	Indexes   []int
	Columns   []int
	Extracted []Payload
}

// ScanResult holds scanned files and exemptions
type ScanResult struct {
	filenameMap map[string]CodeLine
	mutex       sync.RWMutex
	exempt      map[string]struct{}
}

// DefaultExempt files that are skipped
var DefaultExempt = []string{
	"uv.lock", "pyproject.toml", "pnpm-lock.yaml", "package-lock.json",
	"yarn.lock", "go.sum", "deno.lock", "Cargo.lock",
	".gitignore", ".python-version", "LICENSE", ".gitaegis.jsonl",
	".git/", "gitaegis/",
}

// Init initializes ScanResult
func (res *ScanResult) Init() {
	res.filenameMap = make(map[string]CodeLine)
	res.exempt = make(map[string]struct{}, len(DefaultExempt))
	for _, f := range DefaultExempt {
		res.exempt[f] = struct{}{}
	}
}

// AddExempt adds a file to exemption list
func (res *ScanResult) AddExempt(file string) {
	res.mutex.Lock()
	defer res.mutex.Unlock()
	res.exempt[file] = struct{}{}
}

// initGitIgnore loads .gitignore safely
func initGitIgnore() *gitignore.GitIgnore {
	ign, err := gitignore.CompileIgnoreFile(".gitignore")
	if err != nil {
		if os.IsNotExist(err) {
			return gitignore.CompileIgnoreLines()
		}
		log.Printf("[core.analyzer] failed to load .gitignore (%v)", err)
		return gitignore.CompileIgnoreLines()
	}
	return ign
}

// isExempt checks if a filename is in exemptions
func (res *ScanResult) isExempt(filename string) bool {
	base := filepath.Base(filename)
	if _, ok := res.exempt[base]; ok {
		return true
	}
	// Check for directory exemptions in the path
	for exempt := range res.exempt {
		if strings.Contains(filename, string(filepath.Separator)+exempt) || strings.HasPrefix(filename, exempt) {
			return true
		}
	}
	return false
}

// isExecutable checks if file is executable
func isExecutable(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return info.Mode().Perm()&0111 != 0
}

// ignoreFiles checks if a file is ignored by gitignore
func ignoreFiles(path string, ign *gitignore.GitIgnore) bool {
	return ign != nil && ign.MatchesPath(path)
}

// IsFilenameMapEmpty returns true if no files have been scanned
func (res *ScanResult) IsFilenameMapEmpty() bool {
	res.mutex.RLock()
	defer res.mutex.RUnlock()
	return len(res.filenameMap) == 0
}

// ClearMap clears all scan results
func (res *ScanResult) ClearMap() {
	res.mutex.Lock()
	defer res.mutex.Unlock()
	res.filenameMap = make(map[string]CodeLine)
}

// IterFolder scans a folder recursively
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
				var lines *CodeLine
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[core.analyzer] Recovered panic scanning %s: %v", filename, r)
					}
				}()

				tree, code, err := CreateTree(filename)
				if err != nil {
					lines = res.PerLineScan(filename, filter)
				} else {
					lines = walkParse(tree.RootNode(), filter, code)
				}

				if lines != nil && len(lines.Lines) > 0 {
					res.mutex.Lock()
					res.filenameMap[filename] = *lines
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


func (res *ScanResult) IterFiles(files []string, filter LineFilter, maxFileSize int64) error {
	numWorkers := runtime.NumCPU()
	fileCh := make(chan string, numWorkers*2)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filename := range fileCh {
				var lines *CodeLine
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[core.analyzer] Recovered panic scanning %s: %v", filename, r)
					}
				}()

				tree, code, err := CreateTree(filename)
				if err != nil {
					lines = res.PerLineScan(filename, filter)
				} else {
					lines = walkParse(tree.RootNode(), filter, code)
				}

				if lines != nil && len(lines.Lines) > 0 {
					res.mutex.Lock()
					res.filenameMap[filename] = *lines
					res.mutex.Unlock()
				}
			}
		}()
	}

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil || info.IsDir() || info.Size() > maxFileSize {
			continue
		}
		fileCh <- f
	}
	close(fileCh)
	wg.Wait()
	return nil
}
// PerLineScan scans file line by line as a fallback
func (res *ScanResult) PerLineScan(filename string, filter LineFilter) *CodeLine {
	if filter == nil {
		return nil
	}

	f, err := os.Open(filename)
	if err != nil {
		log.Printf("[core.analyzer] Error reading file %s: %v", filename, err)
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
			if ok && pl != nil {
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

	return &CodeLine{
		Lines:     lines,
		Indexes:   indexes,
		Columns:   columns,
		Extracted: extracted,
	}
}

// PrettyPrintResults prints results with colors
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
	b.Grow(4096)

	b.WriteString(yellow)
	b.WriteString("gitaegis DETECTED THE FOLLOWING SECRETS\n")
	b.WriteString("=======================================\n")
	b.WriteString(reset)

	for filename, lines := range res.filenameMap {
		b.WriteString("\n")
		b.WriteString(green)
		b.WriteString("File: ")
		b.WriteString(reset)
		b.WriteString(filename)
		b.WriteByte('\n')

		for i, l := range lines.Lines {
			b.WriteString(yellow)
			b.WriteString("---------------------------------------\n")
			b.WriteString(reset)

			fmt.Fprintf(&b, "%sLine %d (Col %d):%s %s\n",
				red, lines.Indexes[i], lines.Columns[i], reset, l)

			if i < len(lines.Extracted) && lines.Extracted[i] != nil {
				for k, v := range lines.Extracted[i] {
					fmt.Fprintf(&b, "  %s%s:%s %v\n", red, k, reset, v)
				}
			}
		}
	}

	os.Stdout.WriteString(b.String())
}
