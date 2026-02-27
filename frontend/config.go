package frontend

import (
	"fmt"
	"log"
	"os"
	"sync"
	"encoding/json"

	toml "github.com/BurntSushi/toml"
)

// Config represents the structure of the TOML configuration file
type Config struct {
	Logging       bool     `toml:"logging"`
	TreeSitterDir string   `toml:"treesitter_source"`
	OutputFormat  []string `toml:"output_format"`
	UseGitignore  bool     `toml:"use_gitignore"`
	Filter        Filter   `toml:"filter"`
	GitDiffOpt	  bool     `toml:"use_gitdiff"`
}

// Filter section of the TOML config
type Filter struct {
	EntLimit    float64           `toml:"ent_limit"`
	MaxFileSize int               `toml:"max_file_size"`
	TargetRegex map[string]string `toml:"target_regex"`
}

var (
	configOnce sync.Once
	globalConfig *Config
	defaultCfgPath = "aegis.config.toml"
)

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
//use sync.once to init config efficiently only once when aegis.toml is changed
func LazyInitConfig() *Config {
	configOnce.Do(func() {
		if _, err := os.Stat(defaultCfgPath); os.IsNotExist(err) {
			log.Printf("[Config] %s not found — skipping initialization", defaultCfgPath)
			return
		}

		cfg, err := LoadConfig(defaultCfgPath)
		if err != nil {
			log.Printf("[Config] failed to load: %v", err)
			return
		}

		globalConfig = cfg
		globalConfig.IntegrateConfig()
		log.Printf("[Config] loaded successfully from %s", defaultCfgPath)
	})

	return globalConfig
}
// IntegrateConfig applies loaded configuration to global state
func (c *Config) IntegrateConfig() {
    
    // --- ADDED: Print the loaded configuration ---
    // Use json.MarshalIndent for a clean, readable printout
    configData, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        log.Printf("[Config] Error marshaling config for printing: %v", err)
    } else {
        log.Printf("[Config] Applying loaded config:\n%s", string(configData))
    }

    if rv == nil {
        return
    }
    if c.Logging {
        rv.SetLogging(c.Logging)
    }
	if c.GitDiffOpt{
		rv.SetGitDiffOpt(true)
	}
    if c.TreeSitterDir != "" {
        rv.SetTreeSitterPath(c.TreeSitterDir)
    }
    rv.SetUseGitignore(c.UseGitignore)
    if c.Filter.EntLimit > 0 {
        rv.SetEntropyLimit(c.Filter.EntLimit)
    }
    if c.Filter.MaxFileSize > 0 {
        rv.SetMaxFileSize(int64(c.Filter.MaxFileSize))
    }
    if len(c.Filter.TargetRegex) > 0 {
        rv.SetFilters(c.Filter.TargetRegex)
    } else {
        fmt.Println("No config is found, use default of 5.0 entrophy limit")
    }
}
