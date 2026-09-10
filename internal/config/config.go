package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const FileName = ".gh-tuissue.json"

type Config struct {
	ProjectNumber int           `json:"project_number,omitempty"`
	HiddenColumns []string      `json:"hidden_columns,omitempty"`
	Sort          string        `json:"sort,omitempty"`
	Agents        []AgentAction `json:"agents,omitempty"`
}

// AgentAction binds a key to a Claude Code skill. Pressing the key on an
// issue launches a background Claude Code session running the skill.
type AgentAction struct {
	// Key is the single character that triggers the action.
	Key string `json:"key"`
	// Skill is the slash command to run, with or without the leading "/".
	Skill string `json:"skill"`
	// Label is shown in the help view; defaults to the skill name.
	Label string `json:"label,omitempty"`
	// Args is the argument template passed to the skill. Supports
	// {{number}}, {{title}} and {{url}}. Defaults to "{{number}}".
	Args string `json:"args,omitempty"`
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
