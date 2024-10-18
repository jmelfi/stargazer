package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration settings.
type Config struct {
	GithubUser    string   `yaml:"github_user"`      // GitHub username
	GithubToken   string   `yaml:"github_token"`     // GitHub access token
	OutputFile    string   `yaml:"output_file"`      // Path to the output file
	OutputFormat  string   `yaml:"output_format"`    // Format of the output (e.g., "list" or "table")
	IgnoreRepos   []string `yaml:"ignore_repos"`     // List of repositories to ignore
	WithTOC       bool     `yaml:"with_toc"`         // Whether to include a table of contents
	WithStars     bool     `yaml:"with_stars"`       // Whether to include star counts
	WithLicense   bool     `yaml:"with_license"`     // Whether to include license information
	WithBackToTop bool     `yaml:"with_back_to_top"` // Whether to include "back to top" links
	Test          bool     `yaml:"test"`             // Whether to use test data
	RateLimit     int      `yaml:"rate_limit"`       // Number of API requests per second
}

// LoadConfig loads the configuration from a YAML file.
// If the file doesn't exist, it returns a default configuration.
// If the file exists but can't be read or parsed, it returns an error.
func LoadConfig(filename string) (*Config, error) {
	config := &Config{
		OutputFile:   "README.md",
		OutputFormat: "list",
		WithTOC:      true,
		WithStars:    true,
		WithLicense:  true,
		Test:         false,
		RateLimit:    5,
	}

	// Check if config file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		logger.WithField("filename", filename).Info("Config file not found. Using default configuration.")
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return config, nil
}

// Save writes the configuration to a YAML file.
// It returns an error if the configuration can't be marshaled or written to the file.
func (c *Config) Save(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}
