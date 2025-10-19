package tests

import (
	"fmt"
	"path/filepath"
	"testing"
	runner "github.com/steverahardjo/GitAegis/frontend"
)

// Assume you have: func NewRuntimeConfig() *RuntimeValue
// and method: func (rv *RuntimeValue) Scan(path string) (*ScanResult, error)

func TestRuntimeScan(t *testing.T) {
	rv := runner.NewRuntimeConfig()

	// set regex filters
	rv.SetFilters(map[string]string{
		"password": `(?i)password\s*[:=]`,
		"apikey":   `(?i)api[_-]?key\s*[:=]`,
	})

	// test directory (change this to something local in your repo)
	testDir := filepath.Join("testdata", "dummy_project")

	res, err := rv.Scan(testDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// you can print results for inspection
	fmt.Printf("Scan result: %+v\n", res)
}
