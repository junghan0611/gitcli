// log_cmd.go — gitcli log: repo-specific commit log
package main

import (
	"strconv"
	"strings"
	"time"
)

type LogResult struct {
	Repo    string      `json:"repo"`
	Period  string      `json:"period"`
	Total   int         `json:"total"`
	Commits []LogCommit `json:"commits"`
}

type LogCommit struct {
	Hash    string `json:"hash"`
	Date    string `json:"date"`
	Time    string `json:"time"`
	Message string `json:"message"`
	Author  string `json:"author"`
}

// QueryLog returns commits for a specific repo within a period.
func QueryLog(repoPath, repoName, daysStr, from, to, author string) LogResult {
	result := LogResult{Repo: repoName}

	args := []string{
		"--format=%h|%ad|%s|%an",
		"--date=format:%Y-%m-%d %H:%M",
	}

	if from != "" && to != "" {
		args = append(args, "--since="+from+"T00:00:00", "--until="+to+"T23:59:59")
		result.Period = from + " ~ " + to
	} else if daysStr != "" {
		n, _ := strconv.Atoi(daysStr)
		if n <= 0 {
			n = 7
		}
		since := time.Now().AddDate(0, 0, -n).Format("2006-01-02")
		today := time.Now().Format("2006-01-02")
		args = append(args, "--since="+since+"T00:00:00")
		result.Period = since + " ~ " + today
	} else {
		// Default: 7 days
		since := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		today := time.Now().Format("2006-01-02")
		args = append(args, "--since="+since+"T00:00:00")
		result.Period = since + " ~ " + today
	}

	if author != "" {
		args = append(args, "--author="+author)
	}

	lines := gitLog(repoPath, args...)
	for _, line := range lines {
		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 4 {
			continue
		}
		datetime := strings.SplitN(parts[1], " ", 2)
		d, t := "", ""
		if len(datetime) == 2 {
			d = datetime[0]
			t = datetime[1]
		}
		result.Commits = append(result.Commits, LogCommit{
			Hash:    parts[0],
			Date:    d,
			Time:    t,
			Message: parts[2],
			Author:  parts[3],
		})
	}
	result.Total = len(result.Commits)
	return result
}
