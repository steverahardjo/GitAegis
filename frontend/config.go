package frontend

import (
	"fmt"
	"os"

	toml "github.com/BurntSushi/toml"
)

// Config represents the structure of the TOML configuration file
type Config struct {
	Logging       bool     `toml:"logging"`
	TreeSitterDir string   `toml:"treesitter_source"`
	OutputFormat  []string `toml:"output_format"`
	UseGitignore  bool     `toml:"use_gitignore"`
	Filter        Filter   `toml:"filter"`
}

// Filter section of the TOML config
type Filter struct {
	EntLimit    float64           `toml:"ent_limit"`
	MaxFileSize int               `toml:"max_file_size"`
	TargetRegex map[string]string `toml:"target_regex"`
}

// LoadConfig reads and decodes a TOML file into Config
func LoadConfig(path string) (*Config, error) {
	var cfg Config

	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &cfg, nil
}

// IntegrateConfig applies loaded configuration to global state
func (c *Config) IntegrateConfig() {
	SetGlobalLogging(c.Logging)
	IntegrateTreeSitter(c.TreeSitterDir)
	SetUseGitignore(c.UseGitignore)
	SetEntropyLimit(c.Filter.EntLimit)
	SetMaxFileSize(int64(c.Filter.MaxFileSize))
	SetGlobalFilters(c.Filter.TargetRegex)
}




