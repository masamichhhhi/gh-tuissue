package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const FileName = ".gh-tuissue.json"

type Config struct {
	ProjectNumber int `json:"project_number,omitempty"`
}

// Load reads the config file from the given repository root.
// Returns nil, nil if the file does not exist (first run).
func Load(repoRoot string) (*Config, error) {
	path := filepath.Join(repoRoot, FileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to the repository root.
func Save(repoRoot string, cfg Config) error {
	path := filepath.Join(repoRoot, FileName)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0644)
}
