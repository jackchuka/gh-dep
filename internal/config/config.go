package config

import (
	"strings"

	"github.com/cli/go-gh/v2/pkg/config"
)

// Config holds all configuration values from gh config
type Config struct {
	Repos    []string // dep.repo (comma-separated)
	Patterns []string // dep.patterns (comma-separated regex patterns)
}

// Load reads configuration from gh config
// Returns a Config with zero values if no config is set
func Load() (*Config, error) {
	ghCfg, err := config.Read(nil)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	if repos, err := ghCfg.Get([]string{"dep.repo"}); err == nil && repos != "" {
		parts := strings.SplitSeq(repos, ",")
		for part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				cfg.Repos = append(cfg.Repos, part)
			}
		}
	}

	if patterns, err := ghCfg.Get([]string{"dep.patterns"}); err == nil && patterns != "" {
		parts := strings.SplitSeq(patterns, ",")
		for part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				cfg.Patterns = append(cfg.Patterns, part)
			}
		}
	}

	return cfg, nil
}

// GetRepos returns the configured repos or nil if not set
func (c *Config) GetRepos() []string {
	return c.Repos
}

// GetPatterns returns the configured patterns or nil if not set
func (c *Config) GetPatterns() []string {
	return c.Patterns
}
