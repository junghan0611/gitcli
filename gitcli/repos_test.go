package main

import (
	"testing"
)

func TestQueryReposInfo(t *testing.T) {
	base := t.TempDir()

	createTestRepo(t, base, "proj1", []struct{ date, msg string }{
		{"2024-01-15T10:00:00", "init"},
		{"2025-06-20T14:00:00", "update"},
	})

	repos := DiscoverRepos([]string{base})
	result := QueryReposInfo(repos)

	if result.TotalRepos != 1 {
		t.Errorf("total_repos = %d, want 1", result.TotalRepos)
	}

	r := result.Repos[0]
	if r.Name != "proj1" {
		t.Errorf("name = %q", r.Name)
	}
	if r.TotalCommits != 2 {
		t.Errorf("total_commits = %d, want 2", r.TotalCommits)
	}
	if r.FirstCommit != "2024-01-15" {
		t.Errorf("first_commit = %q, want 2024-01-15", r.FirstCommit)
	}
	if r.LastCommit != "2025-06-20" {
		t.Errorf("last_commit = %q, want 2025-06-20", r.LastCommit)
	}
	if r.CurrentBranch == "" {
		t.Error("current_branch should not be empty")
	}
}
