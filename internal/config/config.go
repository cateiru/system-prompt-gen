package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	InputDir     string   `json:"inputDir"`
	OutputFiles  []string `json:"outputFiles"`
	ExcludeFiles []string `json:"excludeFiles"`
	Header       string   `json:"header"`
	Footer       string   `json:"footer"`
}

func DefaultConfig() *Config {
	return &Config{
		InputDir:    ".system_prompt",
		OutputFiles: []string{"CLAUDE.md", ".clinerules"},
		ExcludeFiles: []string{},
		Header:      "# System Prompt\n\n",
		Footer:      "",
	}
}

func LoadConfig(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Save(configPath string) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}