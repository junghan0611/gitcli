package main

import (
	"testing"
)

func TestQueryLog(t *testing.T) {
	base := t.TempDir()

	repoPath := createTestRepo(t, base, "myrepo", []struct{ date, msg string }{
		{"2025-10-10T09:00:00", "first"},
		{"2025-10-11T10:00:00", "second"},
		{"2025-10-12T11:00:00", "third"},
	})

	result := QueryLog(repoPath, "myrepo", "", "2025-10-10", "2025-10-12", "")

	if result.Repo != "myrepo" {
		t.Errorf("repo = %q", result.Repo)
	}
	if result.Total != 3 {
		t.Errorf("total = %d, want 3", result.Total)
	}
}

func TestQueryLogDays(t *testing.T) {
	base := t.TempDir()

	repoPath := createTestRepo(t, base, "myrepo", []struct{ date, msg string }{
		{"2025-10-10T09:00:00", "old commit"},
	})

	// --days 1 should not include a commit from months ago
	result := QueryLog(repoPath, "myrepo", "1", "", "", "")
	if result.Total != 0 {
		t.Errorf("expected 0 recent commits, got %d", result.Total)
	}
}

func TestQueryLogAuthor(t *testing.T) {
	base := t.TempDir()

	repoPath := createTestRepo(t, base, "myrepo", []struct{ date, msg string }{
		{"2025-10-10T09:00:00", "commit"},
	})

	result := QueryLog(repoPath, "myrepo", "", "2025-10-10", "2025-10-10", "testuser")
	if result.Total != 1 {
		t.Errorf("with author testuser: total = %d, want 1", result.Total)
	}

	result = QueryLog(repoPath, "myrepo", "", "2025-10-10", "2025-10-10", "nobody")
	if result.Total != 0 {
		t.Errorf("with author nobody: total = %d, want 0", result.Total)
	}
}
