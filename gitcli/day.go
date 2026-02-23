// day.go — gitcli day: all commits for a specific date across repos
package main

import (
	"fmt"
	"strings"
)

type DayResult struct {
	Date        string      `json:"date"`
	DayOfWeek   string      `json:"day_of_week"`
	TotalCommits int        `json:"total_commits"`
	Repos       []DayRepo   `json:"repos"`
	Summary     DaySummary  `json:"summary"`
}

type DayRepo struct {
	Name    string      `json:"name"`
	Path    string      `json:"path"`
	Commits []DayCommit `json:"commits"`
}

type DayCommit struct {
	Hash         string `json:"hash"`
	Time         string `json:"time"`
	Message      string `json:"message"`
	FilesChanged int    `json:"files_changed"`
	Insertions   int    `json:"insertions"`
	Deletions    int    `json:"deletions"`
	Author       string `json:"author"`
}

type DaySummary struct {
	ActiveRepos int    `json:"active_repos"`
	FirstCommit string `json:"first_commit"`
	LastCommit  string `json:"last_commit"`
	ActiveHours float64 `json:"active_hours"`
}

// QueryDayMe queries using ~/.config/gitcli/authors patterns.
// tz is an optional timezone offset like "+09:00". Empty means local time.
func QueryDayMe(repos []Repo, date string, tz string) DayResult {
	patterns := DefaultAuthorPatterns()
	result := DayResult{
		Date:      date,
		DayOfWeek: dayOfWeek(date),
		Repos:     []DayRepo{},
	}

	sinceDate, untilDate := dayRange(date, tz)
	logFmt := "%h|%H|%ad|%s|%an"
	var allTimes []string

	for _, repo := range repos {
		args := []string{
			"--format=" + logFmt,
			"--date=format:%H:%M",
			"--since=" + sinceDate,
			"--until=" + untilDate,
			"--all",
		}

		lines := gitLog(repo.Path, args...)
		if len(lines) == 0 {
			continue
		}

		dayRepo := DayRepo{Name: repo.Name, Path: repo.Path}

		for _, line := range lines {
			parts := strings.SplitN(line, "|", 5)
			if len(parts) < 5 {
				continue
			}
			author := parts[4]
			if !IsMyCommit(author, patterns) {
				continue
			}

			fc, ins, del := commitStats(repo.Path, parts[0])
			dayRepo.Commits = append(dayRepo.Commits, DayCommit{
				Hash: parts[0], Time: parts[2], Message: parts[3],
				FilesChanged: fc, Insertions: ins, Deletions: del, Author: author,
			})
			allTimes = append(allTimes, parts[2])
		}

		if len(dayRepo.Commits) > 0 {
			result.Repos = append(result.Repos, dayRepo)
			result.TotalCommits += len(dayRepo.Commits)
		}
	}

	result.Summary.ActiveRepos = len(result.Repos)
	if len(allTimes) > 0 {
		first, last := allTimes[len(allTimes)-1], allTimes[0]
		result.Summary.FirstCommit = first
		result.Summary.LastCommit = last
		result.Summary.ActiveHours = calcActiveHours(first, last)
	}
	return result
}

// QueryDay queries all repos for commits on a specific date.
// tz is an optional timezone offset like "+09:00". Empty means local time.
func QueryDay(repos []Repo, date string, authorFilter string, tz string) DayResult {
	result := DayResult{
		Date:      date,
		DayOfWeek: dayOfWeek(date),
		Repos:     []DayRepo{},
	}

	logFmt := "%h|%H|%ad|%s|%an"
	sinceDate, untilDate := dayRange(date, tz)

	var allTimes []string

	for _, repo := range repos {
		args := []string{
			"--format=" + logFmt,
			"--date=format:%H:%M",
			"--since=" + sinceDate,
			"--until=" + untilDate,
			"--all",
		}
		if authorFilter != "" {
			args = append(args, "--author="+authorFilter)
		}

		lines := gitLog(repo.Path, args...)
		if len(lines) == 0 {
			continue
		}

		dayRepo := DayRepo{
			Name: repo.Name,
			Path: repo.Path,
		}

		for _, line := range lines {
			parts := strings.SplitN(line, "|", 5)
			if len(parts) < 5 {
				continue
			}
			shortHash := parts[0]
			// fullHash := parts[1] // available if needed
			commitTime := parts[2]
			message := parts[3]
			author := parts[4]

			// Get file stats for this commit
			fc, ins, del := commitStats(repo.Path, shortHash)

			dayRepo.Commits = append(dayRepo.Commits, DayCommit{
				Hash:         shortHash,
				Time:         commitTime,
				Message:      message,
				FilesChanged: fc,
				Insertions:   ins,
				Deletions:    del,
				Author:       author,
			})

			allTimes = append(allTimes, commitTime)
		}

		if len(dayRepo.Commits) > 0 {
			result.Repos = append(result.Repos, dayRepo)
			result.TotalCommits += len(dayRepo.Commits)
		}
	}

	// Summary
	result.Summary.ActiveRepos = len(result.Repos)
	if len(allTimes) > 0 {
		first, last := allTimes[len(allTimes)-1], allTimes[0] // git log is newest-first
		result.Summary.FirstCommit = first
		result.Summary.LastCommit = last
		result.Summary.ActiveHours = calcActiveHours(first, last)
	}

	return result
}

// DaySummaryResult is a compact version of DayResult without individual commits.
type DaySummaryResult struct {
	Date         string            `json:"date"`
	DayOfWeek    string            `json:"day_of_week"`
	TotalCommits int               `json:"total_commits"`
	ReposSummary []RepoSummaryItem `json:"repos_summary"`
	Summary      DaySummary        `json:"summary"`
}

type RepoSummaryItem struct {
	Name    string `json:"name"`
	Commits int    `json:"commits"`
}

// ToDaySummary converts a full DayResult to a compact summary.
func (r DayResult) ToDaySummary() DaySummaryResult {
	s := DaySummaryResult{
		Date:         r.Date,
		DayOfWeek:    r.DayOfWeek,
		TotalCommits: r.TotalCommits,
		ReposSummary: make([]RepoSummaryItem, len(r.Repos)),
		Summary:      r.Summary,
	}
	for i, repo := range r.Repos {
		s.ReposSummary[i] = RepoSummaryItem{Name: repo.Name, Commits: len(repo.Commits)}
	}
	return s
}

// dayRange returns since/until strings for git log.
// If tz is set (e.g. "+09:00"), uses ISO 8601 offset so git interprets
// the day boundary in that timezone.
func dayRange(date, tz string) (string, string) {
	if tz == "" {
		return date + "T00:00:00", date + "T23:59:59"
	}
	return date + "T00:00:00" + tz, date + "T23:59:59" + tz
}

// commitStats returns files changed, insertions, deletions for a commit.
func commitStats(repoPath, hash string) (int, int, int) {
	lines := gitLog(repoPath, hash, "-1", "--format=", "--numstat")
	files, ins, del := 0, 0, 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		files++
		if parts[0] != "-" {
			n := 0
			for _, c := range parts[0] {
				if c >= '0' && c <= '9' {
					n = n*10 + int(c-'0')
				}
			}
			ins += n
		}
		if parts[1] != "-" {
			n := 0
			for _, c := range parts[1] {
				if c >= '0' && c <= '9' {
					n = n*10 + int(c-'0')
				}
			}
			del += n
		}
	}
	return files, ins, del
}

// calcActiveHours calculates hours between first and last commit time (HH:MM).
func calcActiveHours(first, last string) float64 {
	t1 := parseHHMM(first)
	t2 := parseHHMM(last)
	diff := t2 - t1
	if diff < 0 {
		diff += 24 * 60
	}
	return float64(diff) / 60.0
}

func parseHHMM(s string) int {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0
	}
	h, m := 0, 0
	fmt.Sscanf(parts[0], "%d", &h)
	fmt.Sscanf(parts[1], "%d", &m)
	return h*60 + m
}
