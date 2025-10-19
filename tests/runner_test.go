package tests

import (
	"path/filepath"
	"testing"
	"time"
	runner "github.com/steverahardjo/GitAegis/frontend"
)

func TestFolderScan_Basic(t *testing.T) {
	rv := runner.NewRuntimeConfig()
	rv.Result.Init()

	rv.SetFilters(map[string]string{
		"password": `(?i)password\s*[:=]`,
		"apikey":   `(?i)api[_-]?key\s*[:=]`,
	})

	testDir := filepath.Join("testdata", "dummy_project")
	found, err := rv.Scan(testDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !found {
		t.Errorf("expected sensitive patterns to be found in %s", testDir)
	}
}

func TestFolderScan_MultipleDirs(t *testing.T) {
	tests := map[string]bool{
		"testdata/dummy_project":        true,
		"testdata/clean_project":        false,
		"testdata/partial_sensitive":    true,
	}

	rv := runner.NewRuntimeConfig()
	rv.Result.Init()
	rv.SetFilters(map[string]string{
		"token": `(?i)token\s*[:=]`,
	})

	for dir, want := range tests {
		t.Run(dir, func(t *testing.T) {
			found, err := rv.Scan(filepath.FromSlash(dir))
			if err != nil {
				t.Fatalf("Scan failed for %s: %v", dir, err)
			}
			if found != want {
				t.Errorf("expected found=%v for %s, got %v", want, dir, found)
			}
		})
	}
}

func TestFileScan(t *testing.T) {
	rv := runner.NewRuntimeConfig()
	rv.Result.Init()

	rv.SetFilters(map[string]string{
		"password": `(?i)password\s*[:=]`,
		"apikey":   `(?i)api[_-]?key\s*[:=]`,
	})

	testFile := filepath.Join("testdata", "dummy_project", "config.txt")

	found, err := rv.Scan(testFile)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !found {
		t.Errorf("expected sensitive patterns to be found in file: %s", testFile)
	}
}

func TestBigFolderScan(t *testing.T) {
	rv := runner.NewRuntimeConfig()

	rv.SetFilters(map[string]string{
		"password": `(?i)password\s*[:=]`,
		"apikey":   `(?i)api[_-]?key\s*[:=]`,
	})

	// Point to a large directory (adjust path)
	bigDir := filepath.Join("/home/holyknight101/Documents/Projects/Personal/exp_site")

	start := time.Now()
	ok, err := rv.Scan(bigDir)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	t.Logf("Scan completed in %v", elapsed)

	if elapsed > 5*time.Second {
		t.Errorf("Scan took too long: %v", elapsed)
	}

	if !ok {
		t.Errorf("Expected successful scan result, got false")
	}
}
