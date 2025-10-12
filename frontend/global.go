package frontend

import (
	"fmt"

	core "github.com/steverahardjo/GitAegis/core"
)

var rv *RuntimeValue

// RuntimeValue encapsulates the runtime configuration and state
type RuntimeValue struct {
	Result         *core.ScanResult
	EntropyLimit   float64
	LoggingEnabled bool
	GitIntegration bool
	UseGitignore   bool
	MaxFileSize    int64
	TreeSitterPath string
	Filters        core.LineFilter
	GlobalResult core.ScanResult
}

// NewRuntimeConfig initializes a RuntimeValue with default values
func NewRuntimeConfig() *RuntimeValue {
	rv := &RuntimeValue{
		Result:         &core.ScanResult{},
		EntropyLimit:   5.0,
		LoggingEnabled: true,
		GitIntegration: false,
		UseGitignore:   true,
		MaxFileSize:    500,
	}
	rv.Result.Init()
	return rv
}

// SetLogging updates the logging flag
func (rv *RuntimeValue) SetLogging(enabled bool) {
	rv.LoggingEnabled = enabled
	if rv.LoggingEnabled {
		fmt.Printf("[Config] Logging enabled\n")
	} else {
		fmt.Printf("[Config] Logging disabled\n")
	}
}

// SetTreeSitterPath safely sets TreeSitter source path if provided
func (rv *RuntimeValue) SetTreeSitterPath(path string) {
	if path == "" {
		return
	}
	rv.TreeSitterPath = path
	fmt.Printf("[Config] TreeSitter path set to %s\n", path)
}

// SetUseGitignore toggles .gitignore usage
func (rv *RuntimeValue) SetUseGitignore(enable bool) {
	rv.UseGitignore = enable
	fmt.Printf("[Config] Use .gitignore = %v\n", enable)
}

// SetEntropyLimit safely updates the entropy limit
func (rv *RuntimeValue) SetEntropyLimit(limit float64) {
	if limit <= 0 {
		limit = 5.0
	}
	rv.EntropyLimit = limit
	fmt.Printf("[Config] Entropy limit set to %.2f\n", limit)
}

// SetMaxFileSize updates the maximum file size limit (in KB)
func (rv *RuntimeValue) SetMaxFileSize(size int64) {
	if size <= 0 {
		size = 500
	}
	rv.MaxFileSize = size
	fmt.Printf("[Config] Max file size set to %d KB\n", size)
}

// SetFilters builds and applies regex + entropy filters
func (rv *RuntimeValue) SetFilters(regexes map[string]string) {
	var filters []core.LineFilter
	for k, v := range regexes {
		filters = append(filters, core.AddTargetRegexPattern(k, v))
	}
	filters = append(filters, core.BasicFilter())
	filters = append(filters, core.EntropyFilter(rv.EntropyLimit))
	rv.Filters = core.AllFilters(filters...)
	fmt.Println("[Config] Filters initialized")
}
