package frontend

import (
	"fmt"
	"log"
	"os"
	"sync"

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
func (c *Config) IntegrateConfig(){

	rv.SetLogging(c.Logging)
	rv.SetTreeSitterPath(c.TreeSitterDir)
	rv.SetUseGitignore(c.UseGitignore)
	rv.SetEntropyLimit(c.Filter.EntLimit)
	rv.SetMaxFileSize(int64(c.Filter.MaxFileSize))
	rv.SetFilters(c.Filter.TargetRegex)
}







