package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanResult_Init(t *testing.T) {
	result := &ScanResult{}
	result.Init()

	if result.filenameMap == nil {
		t.Error("filenameMap should be initialized")
	}
	if len(result.exempt) == 0 {
		t.Error("exempt map should be initialized with default exemptions")
	}
}

func TestScanResult_AddExempt(t *testing.T) {
	result := &ScanResult{}
	result.Init()

	result.AddExempt("test.txt")

	if _, ok := result.exempt["test.txt"]; !ok {
		t.Error("test.txt should be in exempt map")
	}
}

func TestScanResult_IsExempt(t *testing.T) {
	result := &ScanResult{}
	result.Init()

	tests := []struct {
		name     string
		filename string
		wantExempt bool
	}{
		{"lock file", "go.sum", true},
		{"gitignore", ".gitignore", true},
		{"regular file", "main.go", false},
		{"git directory", ".git/config", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := result.isExempt(tt.filename)
			if got != tt.wantExempt {
				t.Errorf("isExempt(%q) = %v, want %v", tt.filename, got, tt.wantExempt)
			}
		})
	}
}

func TestScanResult_IsFilenameMapEmpty(t *testing.T) {
	result := &ScanResult{}
	result.Init()

	if !result.IsFilenameMapEmpty() {
		t.Error("new ScanResult should have empty filenameMap")
	}

	result.filenameMap["test.go"] = CodeLine{Lines: []string{"test"}}

	if result.IsFilenameMapEmpty() {
		t.Error("filenameMap should not be empty after adding entry")
	}
}

func TestScanResult_ClearMap(t *testing.T) {
	result := &ScanResult{}
	result.Init()
	result.filenameMap["test.go"] = CodeLine{Lines: []string{"test"}}

	result.ClearMap()

	if !result.IsFilenameMapEmpty() {
		t.Error("ClearMap should empty the filenameMap")
	}
}

func TestScanResult_IterFolder(t *testing.T) {
	tmpDir := t.TempDir()

	secretFile := filepath.Join(tmpDir, "config.go")
	os.WriteFile(secretFile, []byte(`
		package main
		apiKey := "aB3$kL9@mX2#pQ5!rT8&nV1^wY4*uI7"
	`), 0644)

	cleanFile := filepath.Join(tmpDir, "main.go")
	os.WriteFile(cleanFile, []byte(`
		package main
		func main() { println("hello") }
	`), 0644)

	result := &ScanResult{}
	result.Init()

	filter := AnyFilters(
		EntropyFilter(4.0),
		BasicFilter(),
	)

	err := result.IterFolder(tmpDir, filter, false, 1024*1024)
	if err != nil {
		t.Fatalf("IterFolder failed: %v", err)
	}

	if len(result.filenameMap) == 0 {
		t.Error("expected secrets to be detected in config.go")
	}
}

func TestScanResult_IterFolder_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	result := &ScanResult{}
	result.Init()

	filter := EntropyFilter(4.0)
	err := result.IterFolder(tmpDir, filter, false, 1024*1024)
	if err != nil {
		t.Fatalf("IterFolder failed on empty dir: %v", err)
	}

	if !result.IsFilenameMapEmpty() {
		t.Error("empty directory should have no results")
	}
}

func TestScanResult_IterFolder_UseGitIgnore(t *testing.T) {
	tmpDir := t.TempDir()

	gitignore := filepath.Join(tmpDir, ".gitignore")
	os.WriteFile(gitignore, []byte("*.secret\n"), 0644)

	secretFile := filepath.Join(tmpDir, "test.secret")
	os.WriteFile(secretFile, []byte("secret data"), 0644)

	cleanFile := filepath.Join(tmpDir, "main.go")
	os.WriteFile(cleanFile, []byte("package main"), 0644)

	result := &ScanResult{}
	result.Init()

	filter := EntropyFilter(5.0)
	err := result.IterFolder(tmpDir, filter, true, 1024*1024)
	if err != nil {
		t.Fatalf("IterFolder failed: %v", err)
	}

	if len(result.filenameMap) != 0 {
		t.Error(".secret file should be ignored via gitignore")
	}
}

func TestScanResult_IterFolder_MaxFileSize(t *testing.T) {
	tmpDir := t.TempDir()

	largeFile := filepath.Join(tmpDir, "large.bin")
	os.WriteFile(largeFile, make([]byte, 2048), 0644)

	result := &ScanResult{}
	result.Init()

	filter := EntropyFilter(4.0)
	err := result.IterFolder(tmpDir, filter, false, 1024)
	if err != nil {
		t.Fatalf("IterFolder failed: %v", err)
	}

	if !result.IsFilenameMapEmpty() {
		t.Error("large file should be skipped due to max file size")
	}
}

func TestScanResult_PerLineScan(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(tmpFile, []byte("aB3$kL9@mX2#pQ5!"), 0644)

	result := &ScanResult{}
	result.Init()

	filter := EntropyFilter(3.5)
	lines := result.PerLineScan(tmpFile, filter)

	if lines == nil {
		t.Error("expected lines to be scanned")
	}
	if len(lines.Lines) == 0 {
		t.Error("expected at least one matching line")
	}
}

func TestScanResult_PerLineScan_NoMatch(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(tmpFile, []byte("hello world"), 0644)

	result := &ScanResult{}
	result.Init()

	filter := EntropyFilter(5.0)
	lines := result.PerLineScan(tmpFile, filter)

	if lines != nil {
		t.Error("expected nil for no matches")
	}
}

func TestScanResult_PerLineScan_NilFilter(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	result := &ScanResult{}
	result.Init()

	lines := result.PerLineScan(tmpFile, nil)
	if lines != nil {
		t.Error("nil filter should return nil")
	}
}

func TestScanResult_PerLineScan_FileNotFound(t *testing.T) {
	result := &ScanResult{}
	result.Init()

	filter := EntropyFilter(4.0)
	lines := result.PerLineScan("/nonexistent/file.txt", filter)

	if lines != nil {
		t.Error("missing file should return nil")
	}
}

func TestScanResult_PrettyPrintResults(t *testing.T) {
	result := &ScanResult{}
	result.Init()

	result.filenameMap["test.go"] = CodeLine{
		Lines:   []string{"secret_key = abc123"},
		Indexes: []int{10},
		Columns: []int{5},
		Extracted: []Payload{
			{"type": "api_key", "value": "abc123"},
		},
	}

	result.PrettyPrintResults()
}

func TestScanResult_PrettyPrintResults_Empty(t *testing.T) {
	result := &ScanResult{}
	result.Init()

	result.PrettyPrintResults()
}
