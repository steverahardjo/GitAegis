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
	global_filemaxsize    int64   = 500
	sitter_path           string
	global_filters 		  core.LineFilter
)

// SetGlobalLogging updates global logging flag
func SetGlobalLogging(enabled bool) {
	global_logging = enabled
	fmt.Printf("[Config] Logging set to %v\n", enabled)
}

// IntegrateTreeSitter safely sets TreeSitter source path if provided
func IntegrateTreeSitter(path string) {
	if path == "" {
		return
	}
	sitter_path = path
}

// SetUseGitignore toggles .gitignore usage
func SetUseGitignore(enable bool) {
	global_gitignore = enable
	fmt.Printf("[Config] Use .gitignore = %v\n", enable)
}

// SetEntropyLimit safely updates the entropy limit
func SetEntropyLimit(limit float64) {
	if limit <= 0 {
		limit = 5.0
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

func SetGlobalFilters(regexes map[string]string) {
	var filters []core.LineFilter

	for k, v := range regexes {
		filter := core.AddTargetRegexPattern(k, v)
		filters = append(filters, filter)
	}
	filters = append(filters, core.BasicFilter())
	filters = append(filters, core.EntropyFilter(global_entLimit))
	// Combine all filters into one global LineFilter
	global_filters = core.AllFilters(filters...)
}

