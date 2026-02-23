package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := expandHome("~/repos")
	want := filepath.Join(home, "repos")
	if got != want {
		t.Errorf("expandHome(~/repos) = %q, want %q", got, want)
	}

	// absolute path unchanged
	got = expandHome("/tmp/foo")
	if got != "/tmp/foo" {
		t.Errorf("expandHome(/tmp/foo) = %q", got)
	}
}

func TestIsGitRepo(t *testing.T) {
	// Create a temp dir with .git
	tmp := t.TempDir()
	os.Mkdir(filepath.Join(tmp, ".git"), 0755)
	if !isGitRepo(tmp) {
		t.Error("expected true for dir with .git")
	}

	// Dir without .git
	tmp2 := t.TempDir()
	if isGitRepo(tmp2) {
		t.Error("expected false for dir without .git")
	}
}

func TestDiscoverRepos(t *testing.T) {
	// Create temp structure: base/repo1/.git, base/repo2/.git, base/notrepo/
	base := t.TempDir()
	os.MkdirAll(filepath.Join(base, "repo1", ".git"), 0755)
	os.MkdirAll(filepath.Join(base, "repo2", ".git"), 0755)
	os.MkdirAll(filepath.Join(base, "notrepo"), 0755)

	repos := DiscoverRepos([]string{base})
	if len(repos) != 2 {
		t.Errorf("expected 2 repos, got %d", len(repos))
	}

	names := map[string]bool{}
	for _, r := range repos {
		names[r.Name] = true
	}
	if !names["repo1"] || !names["repo2"] {
		t.Errorf("expected repo1 and repo2, got %v", names)
	}
}
