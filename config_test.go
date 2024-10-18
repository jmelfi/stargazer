package main

import (
	"os"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test loading a non-existent file
	t.Run("Non-existent file", func(t *testing.T) {
		config, err := LoadConfig("non_existent.yml")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if config == nil {
			t.Error("Expected default config, got nil")
		}
		if config.OutputFile != "README.md" {
			t.Errorf("Expected default output file 'README.md', got %s", config.OutputFile)
		}
		if config.RateLimit != 5 {
			t.Errorf("Expected default rate limit 5, got %d", config.RateLimit)
		}
	})

	// Test loading a valid config file
	t.Run("Valid config file", func(t *testing.T) {
		testConfig := `
github_user: testuser
github_token: testtoken
output_file: test_output.md
output_format: table
ignore_repos:
  - repo1
  - repo2
with_toc: false
with_stars: true
with_license: false
with_back_to_top: true
test: true
rate_limit: 10
`
		tmpfile, err := os.CreateTemp("", "test_config*.yml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write([]byte(testConfig)); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatalf("Failed to close temp file: %v", err)
		}

		config, err := LoadConfig(tmpfile.Name())
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		expected := &Config{
			GithubUser:    "testuser",
			GithubToken:   "testtoken",
			OutputFile:    "test_output.md",
			OutputFormat:  "table",
			IgnoreRepos:   []string{"repo1", "repo2"},
			WithTOC:       false,
			WithStars:     true,
			WithLicense:   false,
			WithBackToTop: true,
			Test:          true,
			RateLimit:     10,
		}

		if !reflect.DeepEqual(config, expected) {
			t.Errorf("Loaded config does not match expected. Got %+v, want %+v", config, expected)
		}
	})
}

func TestSaveConfig(t *testing.T) {
	config := &Config{
		GithubUser:    "testuser",
		GithubToken:   "testtoken",
		OutputFile:    "test_output.md",
		OutputFormat:  "table",
		IgnoreRepos:   []string{"repo1", "repo2"},
		WithTOC:       false,
		WithStars:     true,
		WithLicense:   false,
		WithBackToTop: true,
		Test:          true,
		RateLimit:     10,
	}

	tmpfile, err := os.CreateTemp("", "test_config_save*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	err = config.Save(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedConfig, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if !reflect.DeepEqual(config, loadedConfig) {
		t.Errorf("Loaded config does not match saved config. Got %+v, want %+v", loadedConfig, config)
	}
}
