package main

import (
	"testing"
	"time"
)

func TestIsIgnored(t *testing.T) {
	tests := []struct {
		name     string
		ignored  []string
		input    string
		expected bool
	}{
		{"Empty ignored list", []string{}, "repo", false},
		{"Ignored repo", []string{"repo1", "repo2"}, "repo1", true},
		{"Not ignored repo", []string{"repo1", "repo2"}, "repo3", false},
		{"Case insensitive", []string{"Repo1"}, "repo1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ignored = tt.ignored
			if got := isIgnored(tt.input); got != tt.expected {
				t.Errorf("isIgnored() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTestStars(t *testing.T) {
	stars, total := testStars()

	if total != 4 {
		t.Errorf("Expected total of 4, got %d", total)
	}

	expectedLangs := []string{"go", "markdown", "C#", "C++"}
	for _, lang := range expectedLangs {
		if _, ok := stars[lang]; !ok {
			t.Errorf("Expected language %s not found in stars", lang)
		}
	}

	if len(stars["go"]) != 1 || stars["go"][0].Name != "stargazer" {
		t.Errorf("Expected 'stargazer' in 'go' category")
	}

	if len(stars["markdown"]) != 1 || stars["markdown"][0].Name != "stars" {
		t.Errorf("Expected 'stars' in 'markdown' category")
	}
}

// Mock for DefaultFetchStars function
func mockFetchStars(user, token string, rateLimit int) (map[string][]Star, int, error) {
	stars := make(map[string][]Star)
	stars["go"] = []Star{
		{
			Url:           "https://github.com/user/repo1",
			Name:          "repo1",
			NameWithOwner: "user/repo1",
			Description:   "Test repo 1",
			License:       "MIT",
			Stars:         10,
			Archived:      false,
			StarredAt:     time.Now(),
		},
	}
	return stars, 1, nil
}

func TestFetchAndProcessStars(t *testing.T) {
	// Save the original DefaultFetchStars function and restore it after the test
	originalFetchStars := DefaultFetchStars
	defer func() { DefaultFetchStars = originalFetchStars }()

	// Replace DefaultFetchStars with our mock function
	DefaultFetchStars = mockFetchStars

	config := &Config{
		GithubUser:  "testuser",
		GithubToken: "testtoken",
		Test:        false,
		RateLimit:   5,
	}

	stars, total, err := fetchAndProcessStars(config)

	if err != nil {
		t.Fatalf("fetchAndProcessStars() returned an error: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected total of 1, got %d", total)
	}

	if len(stars["go"]) != 1 || stars["go"][0].Name != "repo1" {
		t.Errorf("Expected 'repo1' in 'go' category")
	}
}
