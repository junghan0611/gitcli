package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// createTestRepo creates a git repo with commits on specific dates.
func createTestRepo(t *testing.T, base, name string, commits []struct{ date, msg string }) string {
	t.Helper()
	repoPath := filepath.Join(base, name)
	os.MkdirAll(repoPath, 0755)

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=testuser",
			"GIT_COMMITTER_NAME=testuser",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %s\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "testuser")

	for i, c := range commits {
		fname := filepath.Join(repoPath, "file.txt")
		os.WriteFile(fname, []byte(c.msg+"\n"), 0644)
		run("add", ".")
		cmd := exec.Command("git", "-C", repoPath, "commit", "-m", c.msg, "--allow-empty")
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=testuser",
			"GIT_COMMITTER_NAME=testuser",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_EMAIL=test@test.com",
			"GIT_AUTHOR_DATE="+c.date,
			"GIT_COMMITTER_DATE="+c.date,
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("commit %d failed: %s\n%s", i, err, out)
		}
	}

	return repoPath
}

func TestQueryDay(t *testing.T) {
	base := t.TempDir()

	createTestRepo(t, base, "alpha", []struct{ date, msg string }{
		{"2025-10-10T09:00:00", "morning work"},
		{"2025-10-10T14:30:00", "afternoon fix"},
		{"2025-10-11T10:00:00", "next day"},
	})

	createTestRepo(t, base, "beta", []struct{ date, msg string }{
		{"2025-10-10T11:00:00", "beta feature"},
		{"2025-10-12T08:00:00", "other day"},
	})

	repos := DiscoverRepos([]string{base})
	result := QueryDay(repos, "2025-10-10", "", "")

	if result.Date != "2025-10-10" {
		t.Errorf("date = %q", result.Date)
	}
	if result.DayOfWeek != "Friday" {
		t.Errorf("day_of_week = %q", result.DayOfWeek)
	}
	if result.TotalCommits != 3 {
		t.Errorf("total_commits = %d, want 3", result.TotalCommits)
	}
	if result.Summary.ActiveRepos != 2 {
		t.Errorf("active_repos = %d, want 2", result.Summary.ActiveRepos)
	}
}

func TestQueryDayAuthorFilter(t *testing.T) {
	base := t.TempDir()

	createTestRepo(t, base, "mixed", []struct{ date, msg string }{
		{"2025-10-10T09:00:00", "my commit"},
	})

	repos := DiscoverRepos([]string{base})

	// Filter by matching author
	result := QueryDay(repos, "2025-10-10", "testuser", "")
	if result.TotalCommits != 1 {
		t.Errorf("with matching author: total=%d, want 1", result.TotalCommits)
	}

	// Filter by non-matching author
	result = QueryDay(repos, "2025-10-10", "nobody", "")
	if result.TotalCommits != 0 {
		t.Errorf("with non-matching author: total=%d, want 0", result.TotalCommits)
	}
}

func TestQueryDayMe(t *testing.T) {
	base := t.TempDir()

	// Set up config for --me
	configDir := filepath.Join(base, ".config", "gitcli")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "authors"), []byte("testuser\n"), 0644)
	t.Setenv("HOME", base)

	repoDir := filepath.Join(base, "repos")
	createTestRepo(t, repoDir, "myrepo", []struct{ date, msg string }{
		{"2025-10-10T09:00:00", "my work"},
	})

	repos := DiscoverRepos([]string{repoDir})
	result := QueryDayMe(repos, "2025-10-10", "")
	if result.TotalCommits != 1 {
		t.Errorf("expected 1 with --me, got %d", result.TotalCommits)
	}
}

func TestQueryDayMeFiltersForks(t *testing.T) {
	base := t.TempDir()

	// Config only matches "testuser"
	configDir := filepath.Join(base, ".config", "gitcli")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "authors"), []byte("testuser\n"), 0644)
	t.Setenv("HOME", base)

	// Create repo with foreign-author commits by manipulating git env
	repoDir := filepath.Join(base, "repos")
	repoPath := filepath.Join(repoDir, "forked")
	os.MkdirAll(repoPath, 0755)

	runGit := func(env []string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
		cmd.Env = append(os.Environ(), env...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %s\n%s", args, err, out)
		}
	}

	runGit(nil, "init")
	runGit(nil, "config", "user.email", "test@test.com")
	runGit(nil, "config", "user.name", "testuser")

	// Commit as testuser (me)
	os.WriteFile(filepath.Join(repoPath, "f.txt"), []byte("mine"), 0644)
	runGit([]string{
		"GIT_AUTHOR_NAME=testuser", "GIT_COMMITTER_NAME=testuser",
		"GIT_AUTHOR_EMAIL=test@test.com", "GIT_COMMITTER_EMAIL=test@test.com",
		"GIT_AUTHOR_DATE=2025-10-10T09:00:00", "GIT_COMMITTER_DATE=2025-10-10T09:00:00",
	}, "add", ".")
	runGit([]string{
		"GIT_AUTHOR_NAME=testuser", "GIT_COMMITTER_NAME=testuser",
		"GIT_AUTHOR_EMAIL=test@test.com", "GIT_COMMITTER_EMAIL=test@test.com",
		"GIT_AUTHOR_DATE=2025-10-10T09:00:00", "GIT_COMMITTER_DATE=2025-10-10T09:00:00",
	}, "commit", "-m", "my commit")

	// Commit as someone else
	os.WriteFile(filepath.Join(repoPath, "f.txt"), []byte("theirs"), 0644)
	runGit([]string{
		"GIT_AUTHOR_NAME=karlvoit", "GIT_COMMITTER_NAME=karlvoit",
		"GIT_AUTHOR_EMAIL=karl@voit.at", "GIT_COMMITTER_EMAIL=karl@voit.at",
		"GIT_AUTHOR_DATE=2025-10-10T10:00:00", "GIT_COMMITTER_DATE=2025-10-10T10:00:00",
	}, "add", ".")
	runGit([]string{
		"GIT_AUTHOR_NAME=karlvoit", "GIT_COMMITTER_NAME=karlvoit",
		"GIT_AUTHOR_EMAIL=karl@voit.at", "GIT_COMMITTER_EMAIL=karl@voit.at",
		"GIT_AUTHOR_DATE=2025-10-10T10:00:00", "GIT_COMMITTER_DATE=2025-10-10T10:00:00",
	}, "commit", "-m", "their commit")

	repos := DiscoverRepos([]string{repoDir})

	// QueryDayMe should only return testuser's commit
	result := QueryDayMe(repos, "2025-10-10", "")
	if result.TotalCommits != 1 {
		t.Errorf("expected 1 (only my commit), got %d", result.TotalCommits)
	}
	if len(result.Repos) > 0 && len(result.Repos[0].Commits) > 0 {
		if result.Repos[0].Commits[0].Author != "testuser" {
			t.Errorf("expected author testuser, got %s", result.Repos[0].Commits[0].Author)
		}
	}
}

func TestQueryDayEmpty(t *testing.T) {
	base := t.TempDir()
	createTestRepo(t, base, "empty", []struct{ date, msg string }{
		{"2025-10-11T10:00:00", "other day"},
	})

	repos := DiscoverRepos([]string{base})
	result := QueryDay(repos, "2025-10-10", "", "")
	if result.TotalCommits != 0 {
		t.Errorf("expected 0, got %d", result.TotalCommits)
	}
}

func TestDayRange(t *testing.T) {
	tests := []struct {
		date, tz, wantSince, wantUntil string
	}{
		{"2025-10-10", "", "2025-10-10T00:00:00", "2025-10-10T23:59:59"},
		{"2025-10-10", "+09:00", "2025-10-10T00:00:00+09:00", "2025-10-10T23:59:59+09:00"},
		{"2026-02-17", "+09:00", "2026-02-17T00:00:00+09:00", "2026-02-17T23:59:59+09:00"},
		{"2026-02-17", "-05:00", "2026-02-17T00:00:00-05:00", "2026-02-17T23:59:59-05:00"},
	}
	for _, tt := range tests {
		since, until := dayRange(tt.date, tt.tz)
		if since != tt.wantSince || until != tt.wantUntil {
			t.Errorf("dayRange(%q, %q) = (%q, %q), want (%q, %q)",
				tt.date, tt.tz, since, until, tt.wantSince, tt.wantUntil)
		}
	}
}

func TestToDaySummary(t *testing.T) {
	result := DayResult{
		Date:      "2025-10-10",
		DayOfWeek: "Friday",
		Repos: []DayRepo{
			{Name: "repo-a", Commits: []DayCommit{{Hash: "abc"}, {Hash: "def"}}},
			{Name: "repo-b", Commits: []DayCommit{{Hash: "ghi"}}},
		},
		TotalCommits: 3,
		Summary: DaySummary{ActiveRepos: 2, FirstCommit: "09:00", LastCommit: "17:00", ActiveHours: 8.0},
	}

	s := result.ToDaySummary()

	if s.TotalCommits != 3 {
		t.Errorf("TotalCommits = %d, want 3", s.TotalCommits)
	}
	if len(s.ReposSummary) != 2 {
		t.Fatalf("ReposSummary len = %d, want 2", len(s.ReposSummary))
	}
	if s.ReposSummary[0].Name != "repo-a" || s.ReposSummary[0].Commits != 2 {
		t.Errorf("repo-a: got %+v", s.ReposSummary[0])
	}
	if s.ReposSummary[1].Name != "repo-b" || s.ReposSummary[1].Commits != 1 {
		t.Errorf("repo-b: got %+v", s.ReposSummary[1])
	}
}

func TestToDaySummaryEmpty(t *testing.T) {
	result := DayResult{
		Date:      "2020-01-01",
		DayOfWeek: "Wednesday",
		Repos:     []DayRepo{},
	}
	s := result.ToDaySummary()
	if len(s.ReposSummary) != 0 {
		t.Errorf("ReposSummary should be empty, got %d", len(s.ReposSummary))
	}
}

func TestTruncateDay(t *testing.T) {
	r := DayResult{
		TotalCommits: 10,
		Repos: []DayRepo{
			{Name: "a", Commits: []DayCommit{{Hash: "1"}, {Hash: "2"}, {Hash: "3"}}},
			{Name: "b", Commits: []DayCommit{{Hash: "4"}, {Hash: "5"}, {Hash: "6"}, {Hash: "7"}}},
			{Name: "c", Commits: []DayCommit{{Hash: "8"}, {Hash: "9"}, {Hash: "10"}}},
		},
	}

	// Truncate to 5
	tr := truncateDay(r, 5)
	total := 0
	for _, repo := range tr.Repos {
		total += len(repo.Commits)
	}
	if total != 5 {
		t.Errorf("truncated to %d commits, want 5", total)
	}
	if tr.Summary.Truncated != 5 {
		t.Errorf("Truncated = %d, want 5", tr.Summary.Truncated)
	}

	// No truncation needed
	tr2 := truncateDay(r, 100)
	if tr2.Summary.Truncated != 0 {
		t.Errorf("should not truncate, got %d", tr2.Summary.Truncated)
	}
}
