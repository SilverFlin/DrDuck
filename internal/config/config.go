package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AIProvider      string       `yaml:"ai_provider"`
	DocStorage      string       `yaml:"doc_storage"`
	ADRTemplate     string       `yaml:"adr_template"`
	Hooks           HooksConfig  `yaml:"hooks"`
	DocPath         string       `yaml:"doc_path"`
	SeparateRepoURL string       `yaml:"separate_repo_url,omitempty"`
	AISettings      AISettings   `yaml:"ai_settings"`
}

type HooksConfig struct {
	PreCommit bool `yaml:"pre_commit"`
	PrePush   bool `yaml:"pre_push"`
}

type AISettings struct {
	Persona           string   `yaml:"persona"`
	Sensitivity       string   `yaml:"sensitivity"`
	IgnorePatterns    []string `yaml:"ignore_patterns,omitempty"`
	RequireADRFor     []string `yaml:"require_adr_for,omitempty"`
	NeverRequireADRFor []string `yaml:"never_require_adr_for,omitempty"`
}

const (
	ConfigDir      = ".drduck"
	ConfigFile     = "config.yml"
	DefaultDocPath = "docs/adrs"
)

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		AIProvider:  "claude-code",
		DocStorage:  "same-repo",
		ADRTemplate: "madr",
		Hooks: HooksConfig{
			PreCommit: false,
			PrePush:   false,
		},
		DocPath: DefaultDocPath,
		AISettings: AISettings{
			Persona:     "drduck",
			Sensitivity: "moderate",
			IgnorePatterns: []string{
				"*.md",
				"*.txt", 
				"docs/*",
				"README*",
				"test/*",
				"tests/*",
				"*_test.go",
				"*.test.*",
			},
			RequireADRFor: []string{
				"database",
				"architecture",
				"api design",
				"security",
				"performance",
			},
			NeverRequireADRFor: []string{
				"typo",
				"formatting", 
				"comment",
				"log message",
				"debug",
			},
		},
	}
}

// GetConfigDir returns the path to the .drduck directory
func GetConfigDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Join(cwd, ConfigDir), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigFile), nil
}

// Load loads the configuration from the config file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Return default config if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Save saves the configuration to the config file
func (c *Config) Save() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, ConfigFile)

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Exists checks if the config file exists
func Exists() (bool, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// IsInitialized checks if the project is initialized (has .drduck directory)
func IsInitialized() (bool, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return false, err
	}

	info, err := os.Stat(configDir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}