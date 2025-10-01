package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/jackchuka/gh-dep/internal/types"
)

// GetCachePath returns the path to the cache file
func GetCachePath() (string, error) {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cacheDir = filepath.Join(home, ".cache")
	}

	depCache := filepath.Join(cacheDir, "gh-dep")
	if err := os.MkdirAll(depCache, 0755); err != nil {
		return "", err
	}

	return filepath.Join(depCache, "groups.json"), nil
}

// Save writes the cache to disk
func Save(cache *types.Cache) error {
	path, err := GetCachePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Load reads the cache from disk
func Load() (*types.Cache, error) {
	path, err := GetCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cache types.Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}
