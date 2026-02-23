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
func QueryDayMe(repos []Repo, date string) DayResult {
	patterns := DefaultAuthorPatterns()
	result := DayResult{
		Date:      date,
		DayOfWeek: dayOfWeek(date),
	}

	sinceDate := date + "T00:00:00"
	untilDate := date + "T23:59:59"
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
func QueryDay(repos []Repo, date string, authorFilter string) DayResult {
	result := DayResult{
		Date:      date,
		DayOfWeek: dayOfWeek(date),
	}

	// git log format: hash|time|message|author
	// --numstat for file stats
	logFmt := "%h|%H|%ad|%s|%an"
	sinceDate := date + "T00:00:00"
	untilDate := date + "T23:59:59"

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
