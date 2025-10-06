package frontend

import (
	"fmt"
	core "github.com/steverahardjo/GitAegis/core"
)

var (
	global_result         *core.ScanResult
	global_entLimit       float64 = 5.0
	global_logging        bool    = true
	global_git_integration bool   = false
	global_gitignore      bool    = true
	global_filemaxsize    int64   = 500 // default in KB
	sitter_path           string
	targetRegex           map[string]string
	lineExemptions        map[string]string
)

// SetGlobalLogging updates global logging flag
func SetGlobalLogging(enabled bool) {
	global_logging = enabled
	fmt.Printf("[Config] Logging set to %v\n", enabled)
}

// IntegrateTreeSitter safely sets TreeSitter source path if provided
func IntegrateTreeSitter(path string) {
	if path == "" {
		fmt.Println("[TreeSitter] No directory provided — skipping.")
		return
	}
	sitter_path = path
	fmt.Printf("[TreeSitter] Integrated from %s\n", sitter_path)
}

// SetUseGitignore toggles .gitignore usage
func SetUseGitignore(enable bool) {
	global_gitignore = enable
	fmt.Printf("[Config] Use .gitignore = %v\n", enable)
}

// SetEntropyLimit safely updates the entropy limit
func SetEntropyLimit(limit float64) {
	if limit <= 0 {
		limit = 5.0 // sensible default
	}
	global_entLimit = limit
	fmt.Printf("[Config] Entropy limit set to %.2f\n", limit)
}

// SetMaxFileSize updates maximum file size limit (in KB)
func SetMaxFileSize(size int64) {
	if size <= 0 {
		size = 500
	}
	global_filemaxsize = size
	fmt.Printf("[Config] Max file size set to %d KB\n", size)
}

// SetRegexFilters loads regex filters
func SetRegexFilters(filters map[string]string) {
	if len(filters) == 0 {
		return
	}
	targetRegex = filters
	fmt.Printf("[Config] %d regex filters loaded\n", len(filters))
}

// SetLineExemptions loads line exemption rules
func SetLineExemptions(exempt map[string]string) {
	if len(exempt) == 0 {
		return
	}
	lineExemptions = exempt
	fmt.Printf("[Config] %d line exemptions loaded\n", len(exempt))
}
