package frontend

import (
	"os"
	"sync"
	"testing"
)

func TestLoadConfig_Valid(t *testing.T) {
	tmpFile := "test_config.toml"
	defer os.Remove(tmpFile)

	content := `
logging = true
treesitter_source = "/path/to/treesitter"
output_format = ["json", "txt"]
use_gitignore = true
use_gitdiff = true


[filter]
ent_limit = 4.5
max_file_size = 1024
target_regex = { aws = "AKIA.*", github = "ghp_.*" }
`
	os.WriteFile(tmpFile, []byte(content), 0644)

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Logging != true {
		t.Error("expected logging = true")
	}
	if cfg.TreeSitterDir != "/path/to/treesitter" {
		t.Errorf("expected treesitter_source = /path/to/treesitter, got %s", cfg.TreeSitterDir)
	}
	if cfg.UseGitignore != true {
		t.Error("expected use_gitignore = true")
	}
	if cfg.GitDiffOpt != true{
		t.Error("expected use_gitdiff = true")
	}
	if cfg.Filter.EntLimit != 4.5 {
		t.Errorf("expected ent_limit = 4.5, got %f", cfg.Filter.EntLimit)
	}
	if cfg.Filter.MaxFileSize != 1024 {
		t.Errorf("expected max_file_size = 1024, got %d", cfg.Filter.MaxFileSize)
	}
	if len(cfg.Filter.TargetRegex) != 2 {
		t.Errorf("expected 2 target_regex patterns, got %d", len(cfg.Filter.TargetRegex))
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.toml")
	if err == nil {
		t.Error("expected error for missing config")
	}
}

func TestLoadConfig_InvalidTOML(t *testing.T) {
	tmpFile := "invalid_config.toml"
	defer os.Remove(tmpFile)

	content := `
logging = true
[filter
ent_limit = 4.5
`
	os.WriteFile(tmpFile, []byte(content), 0644)

	_, err := LoadConfig(tmpFile)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

func TestLoadConfig_Minimal(t *testing.T) {
	tmpFile := "minimal_config.toml"
	defer os.Remove(tmpFile)

	content := `
logging = false
use_gitignore = false
use_gitdiff = true

[filter]
ent_limit = 3.0
`
	os.WriteFile(tmpFile, []byte(content), 0644)

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Logging != false {
		t.Error("expected logging = false")
	}
	if cfg.Filter.EntLimit != 3.0 {
		t.Errorf("expected ent_limit = 3.0, got %f", cfg.Filter.EntLimit)
	}
}

func TestLazyInitConfig_NoConfigFile(t *testing.T) {
	globalConfig = nil
	configOnce = sync.Once{}

	cfg := LazyInitConfig()
	if cfg != nil {
		t.Error("expected nil config when file doesn't exist")
	}
}

func TestRuntimeValue_NewRuntimeConfig(t *testing.T) {
	rv := NewRuntimeConfig()

	if rv == nil {
		t.Fatal("NewRuntimeConfig returned nil")
	}
	if rv.Result == nil {
		t.Error("Result should be initialized")
	}
	if rv.EntropyLimit != 4.5 {
		t.Errorf("expected default EntropyLimit = 4.5, got %f", rv.EntropyLimit)
	}
	if rv.LoggingEnabled != true {
		t.Error("expected default LoggingEnabled = true")
	}
	if rv.UseGitignore != true {
		t.Error("expected default UseGitignore = true")
	}
	if rv.MaxFileSize != 500 {
		t.Errorf("expected default MaxFileSize = 500, got %d", rv.MaxFileSize)
	}
}

func TestRuntimeValue_SetLogging(t *testing.T) {
	rv := NewRuntimeConfig()

	rv.SetLogging(false)
	if rv.LoggingEnabled != false {
		t.Error("SetLogging(false) failed")
	}

	rv.SetLogging(true)
	if rv.LoggingEnabled != true {
		t.Error("SetLogging(true) failed")
	}
}

func TestRuntimeValue_SetGitDiffOpt(t *testing.T) {
	rv := NewRuntimeConfig()

	rv.SetGitDiffOpt(false)
	if rv.GitDiffScan != false {
		t.Error("SetGitDiffOpt(false) failed")
	}

	rv.SetGitDiffOpt(true)
	if rv.GitDiffScan != true {
		t.Error("SetGitDiffOpt(true) failed")
	}
}

func TestRuntimeValue_SetTreeSitterPath(t *testing.T) {
	rv := NewRuntimeConfig()

	rv.SetTreeSitterPath("")
	if rv.TreeSitterPath != "" {
		t.Error("empty path should not be set")
	}

	rv.SetTreeSitterPath("/test/path")
	if rv.TreeSitterPath != "/test/path" {
		t.Errorf("expected TreeSitterPath = /test/path, got %s", rv.TreeSitterPath)
	}
}

func TestRuntimeValue_SetUseGitignore(t *testing.T) {
	rv := NewRuntimeConfig()

	rv.SetUseGitignore(false)
	if rv.UseGitignore != false {
		t.Error("SetUseGitignore(false) failed")
	}
}

func TestRuntimeValue_SetEntropyLimit(t *testing.T) {
	rv := NewRuntimeConfig()

	rv.SetEntropyLimit(5.0)
	if rv.EntropyLimit != 5.0 {
		t.Errorf("expected EntropyLimit = 5.0, got %f", rv.EntropyLimit)
	}
}

func TestRuntimeValue_SetMaxFileSize(t *testing.T) {
	rv := NewRuntimeConfig()

	rv.SetMaxFileSize(2048)
	if rv.MaxFileSize != 2048 {
		t.Errorf("expected MaxFileSize = 2048, got %d", rv.MaxFileSize)
	}

	rv.SetMaxFileSize(0)
	if rv.MaxFileSize != 500 {
		t.Error("zero size should default to 500")
	}

	rv.SetMaxFileSize(-100)
	if rv.MaxFileSize != 500 {
		t.Error("negative size should default to 500")
	}
}

func TestRuntimeValue_SetFilters(t *testing.T) {
	rv := NewRuntimeConfig()

	regexes := map[string]string{
		"aws":   "AKIA.*",
		"github": "ghp_.*",
	}

	rv.SetFilters(regexes)

	if rv.Filters == nil {
		t.Error("Filters should be initialized")
	}
}

func TestRuntimeValue_SetFilters_Empty(t *testing.T) {
	rv := NewRuntimeConfig()

	rv.SetFilters(map[string]string{})

	if rv.Filters == nil {
		t.Error("Filters should be initialized even with empty regexes")
	}
}
