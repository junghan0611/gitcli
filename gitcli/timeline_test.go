package main

import (
	"testing"
)

func TestQueryTimeline(t *testing.T) {
	base := t.TempDir()

	createTestRepo(t, base, "proj", []struct{ date, msg string }{
		{"2025-10-10T09:00:00", "day1 morning"},
		{"2025-10-10T14:00:00", "day1 afternoon"},
		{"2025-10-11T10:00:00", "day2"},
		{"2025-10-13T08:00:00", "day4"},
	})

	repos := DiscoverRepos([]string{base})
	result := QueryTimeline(repos, "", "2025-10", "")

	if result.TotalCommits != 4 {
		t.Errorf("total_commits = %d, want 4", result.TotalCommits)
	}
	if result.ActiveDays != 3 {
		t.Errorf("active_days = %d, want 3", result.ActiveDays)
	}

	// Check first day has 2 commits
	if len(result.Daily) < 1 {
		t.Fatal("no daily entries")
	}
	found := false
	for _, d := range result.Daily {
		if d.Date == "2025-10-10" {
			found = true
			if d.Commits != 2 {
				t.Errorf("2025-10-10 commits = %d, want 2", d.Commits)
			}
			if len(d.Repos) != 1 || d.Repos[0] != "proj" {
				t.Errorf("repos = %v", d.Repos)
			}
		}
	}
	if !found {
		t.Error("2025-10-10 not found in daily")
	}
}

func TestQueryTimelineAuthorFilter(t *testing.T) {
	base := t.TempDir()

	createTestRepo(t, base, "proj", []struct{ date, msg string }{
		{"2025-10-10T09:00:00", "commit"},
	})

	repos := DiscoverRepos([]string{base})

	result := QueryTimeline(repos, "", "2025-10", "nobody")
	if result.TotalCommits != 0 {
		t.Errorf("expected 0 with wrong author, got %d", result.TotalCommits)
	}
}
