// timeline.go — gitcli timeline: activity overview for a period
package main

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

type TimelineResult struct {
	Period       string         `json:"period"`
	TotalCommits int           `json:"total_commits"`
	ActiveDays   int           `json:"active_days"`
	Daily        []TimelineDay `json:"daily"`
}

type TimelineDay struct {
	Date    string   `json:"date"`
	Commits int      `json:"commits"`
	Repos   []string `json:"repos"`
	Hours   string   `json:"hours"`
}

// QueryTimeline returns daily commit activity for a period.
func QueryTimeline(repos []Repo, daysStr, month, author string) TimelineResult {
	result := TimelineResult{}

	var since, until string

	if month != "" {
		// Parse YYYY-MM
		t, err := time.ParseInLocation("2006-01", month, time.Local)
		if err != nil {
			fatal("invalid --month format: " + month + " (use YYYY-MM)")
		}
		since = t.Format("2006-01-02")
		until = t.AddDate(0, 1, -1).Format("2006-01-02")
	} else {
		n, _ := strconv.Atoi(daysStr)
		if n <= 0 {
			n = 30
		}
		since = time.Now().AddDate(0, 0, -n).Format("2006-01-02")
		until = time.Now().Format("2006-01-02")
	}

	result.Period = since + " ~ " + until

	// date → {repos: set, times: []string}
	type dayData struct {
		repos map[string]bool
		times []string
	}
	dayMap := make(map[string]*dayData)

	logFmt := "%ad|%an|%s"
	for _, repo := range repos {
		args := []string{
			"--format=" + logFmt,
			"--date=format:%Y-%m-%d %H:%M",
			"--since=" + since + "T00:00:00",
			"--until=" + until + "T23:59:59",
		}
		if author != "" {
			args = append(args, "--author="+author)
		}

		lines := gitLog(repo.Path, args...)
		for _, line := range lines {
			parts := strings.SplitN(line, "|", 3)
			if len(parts) < 3 {
				continue
			}
			datetime := strings.SplitN(parts[0], " ", 2)
			if len(datetime) < 2 {
				continue
			}
			date := datetime[0]
			timeStr := datetime[1]

			dd, ok := dayMap[date]
			if !ok {
				dd = &dayData{repos: make(map[string]bool)}
				dayMap[date] = dd
			}
			dd.repos[repo.Name] = true
			dd.times = append(dd.times, timeStr)
		}
	}

	// Sort dates
	dates := make([]string, 0, len(dayMap))
	for d := range dayMap {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	for _, date := range dates {
		dd := dayMap[date]
		repoNames := make([]string, 0, len(dd.repos))
		for r := range dd.repos {
			repoNames = append(repoNames, r)
		}
		sort.Strings(repoNames)

		// Find earliest and latest times
		sort.Strings(dd.times)
		hours := ""
		if len(dd.times) > 0 {
			hours = dd.times[0] + "~" + dd.times[len(dd.times)-1]
		}

		result.Daily = append(result.Daily, TimelineDay{
			Date:    date,
			Commits: len(dd.times),
			Repos:   repoNames,
			Hours:   hours,
		})
		result.TotalCommits += len(dd.times)
	}

	result.ActiveDays = len(result.Daily)
	return result
}
