// config.go — user identity config for author filtering
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AuthorConfig holds known author identities for the user.
type AuthorConfig struct {
	Authors []string `json:"authors"` // all known author patterns
}

// DefaultAuthorPatterns returns the user's known git author patterns.
// Reads from ~/.config/gitcli/authors (one per line) if exists,
// otherwise returns empty (no filtering).
func DefaultAuthorPatterns() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	configPath := filepath.Join(home, ".config", "gitcli", "authors")
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %s not found — --me filter disabled (all commits included)\n", configPath)
		return nil
	}

	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}
	return patterns
}

// IsMyCommit checks if a commit author matches any known pattern.
// Uses case-insensitive substring matching (same as git --author).
func IsMyCommit(author string, patterns []string) bool {
	if len(patterns) == 0 {
		return true // no filter = include all
	}
	authorLower := strings.ToLower(author)
	for _, p := range patterns {
		if strings.Contains(authorLower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}
