package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsMyCommit(t *testing.T) {
	patterns := []string{"junghan", "jhkim2"}

	tests := []struct {
		author string
		want   bool
	}{
		{"junghan <junghanacs@gmail.com>", true},
		{"junghan0611 <31724164+junghan0611@users.noreply.github.com>", true},
		{"Jung Han <junghanacs@gmail.com>", true},
		{"Junghan Kim <jhkim2@goqual.com>", true},
		{"Karl Voit <git@Karl-Voit.at>", false},
		{"novoid <novoid@users.noreply.github.com>", false},
	}

	for _, tt := range tests {
		got := IsMyCommit(tt.author, patterns)
		if got != tt.want {
			t.Errorf("IsMyCommit(%q) = %v, want %v", tt.author, got, tt.want)
		}
	}
}

func TestIsMyCommitNoPatterns(t *testing.T) {
	// Empty patterns = include all
	if !IsMyCommit("anyone", nil) {
		t.Error("nil patterns should include all")
	}
	if !IsMyCommit("anyone", []string{}) {
		t.Error("empty patterns should include all")
	}
}

func TestDefaultAuthorPatterns(t *testing.T) {
	// Create temp config
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	configDir := filepath.Join(tmp, ".config", "gitcli")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "authors"), []byte("junghan\n# comment\njhkim2\n"), 0644)

	patterns := DefaultAuthorPatterns()
	if len(patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
	if patterns[0] != "junghan" || patterns[1] != "jhkim2" {
		t.Errorf("unexpected patterns: %v", patterns)
	}
}
