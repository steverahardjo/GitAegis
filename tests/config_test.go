package tests

import(
	"testing"
	"path/filepath"
	frontend "github.com/steverahardjo/GitAegis/frontend"
)

func TestLoadConfig(t *testing.T){
	cfgPath, err := filepath.Abs("aegis.config.toml")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	cfg, err := frontend.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	// spot-check key fields
	if !cfg.Logging {
		t.Errorf("expected Logging=true, got false")
	}
	if cfg.Filter.MaxFileSize != 1024 {
		t.Errorf("expected MaxFileSize=1024, got %d", cfg.Filter.MaxFileSize)
	}
	if cfg.Filter.TargetRegex["go"] != ".*\\.go$" {
		t.Errorf("expected regex for go files, got %v", cfg.Filter.TargetRegex["go"])
	}
}