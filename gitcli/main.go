// main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const Version = "0.2.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "day":
		cmdDay()
	case "repos":
		cmdRepos()
	case "log":
		cmdLog()
	case "timeline":
		cmdTimeline()
	case "-V", "--version", "version":
		fmt.Println("gitcli " + Version)
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func cmdDay() {
	args := os.Args[2:]
	date := ""
	if len(os.Args) >= 3 && !strings.HasPrefix(os.Args[2], "--") {
		date = os.Args[2]
		args = os.Args[3:]
	}
	reposStr := getFlag(args, "--repos", "~/repos/gh,~/repos/work")
	author := getFlag(args, "--author", "")
	me := hasFlag(args, "--me")
	yearsAgo := getFlag(args, "--years-ago", "")
	daysAgo := getFlag(args, "--days-ago", "")
	tz := getFlag(args, "--tz", "")
	summary := hasFlag(args, "--summary")

	resolved, err := resolveDate(date, yearsAgo, daysAgo)
	if err != nil {
		fatal(err.Error())
	}

	dirs := strings.Split(reposStr, ",")
	repos := DiscoverRepos(dirs)

	var result DayResult
	if me {
		result = QueryDayMe(repos, resolved, tz)
	} else {
		result = QueryDay(repos, resolved, author, tz)
	}

	if summary {
		printJSON(result.ToDaySummary())
	} else {
		printJSON(result)
	}
}

func cmdRepos() {
	args := os.Args[2:]
	reposStr := getFlag(args, "--repos", "~/repos/gh,~/repos/work")

	dirs := strings.Split(reposStr, ",")
	repos := DiscoverRepos(dirs)
	result := QueryReposInfo(repos)
	printJSON(result)
}

func cmdLog() {
	if len(os.Args) < 3 {
		fatal("usage: gitcli log <repo-name> [--repos DIR,...] [--days N] [--from DATE] [--to DATE] [--author AUTHOR]")
	}
	repoName := os.Args[2]
	args := os.Args[3:]
	reposStr := getFlag(args, "--repos", "~/repos/gh,~/repos/work")
	daysStr := getFlag(args, "--days", "")
	from := getFlag(args, "--from", "")
	to := getFlag(args, "--to", "")
	author := getFlag(args, "--author", "")

	dirs := strings.Split(reposStr, ",")
	repos := DiscoverRepos(dirs)

	repoPath := ""
	for _, r := range repos {
		if r.Name == repoName {
			repoPath = r.Path
			break
		}
	}
	if repoPath == "" {
		fatal("repo not found: " + repoName)
	}

	result := QueryLog(repoPath, repoName, daysStr, from, to, author)
	printJSON(result)
}

func cmdTimeline() {
	args := os.Args[2:]
	reposStr := getFlag(args, "--repos", "~/repos/gh,~/repos/work")
	daysStr := getFlag(args, "--days", "30")
	month := getFlag(args, "--month", "")
	author := getFlag(args, "--author", "")

	dirs := strings.Split(reposStr, ",")
	repos := DiscoverRepos(dirs)
	result := QueryTimeline(repos, daysStr, month, author)
	printJSON(result)
}

func hasFlag(args []string, name string) bool {
	for _, arg := range args {
		if arg == name {
			return true
		}
	}
	return false
}

func getFlag(args []string, name string, def string) string {
	for i, arg := range args {
		if arg == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return def
}

func printJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	enc.Encode(v)
}

func fatal(msg string) {
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, `gitcli %s — Local git timeline CLI for AI agents

Usage:
  gitcli day [DATE] [--repos DIR,...] [--author AUTHOR] [--me] [--years-ago N] [--days-ago N] [--tz OFFSET] [--summary]
  gitcli repos [--repos DIR,...]
  gitcli log <repo-name> [--repos DIR,...] [--days N] [--from DATE] [--to DATE] [--author AUTHOR]
  gitcli timeline [--repos DIR,...] [--days N] [--month YYYY-MM] [--author AUTHOR]

Date formats:
  2025-10-10    Specific date
  20251010      Denote ID compatible
  --years-ago N N years ago today
  --days-ago N  N days ago today

Options:
  --repos DIR,...   Search directories (default: ~/repos/gh,~/repos/work)
  --author AUTHOR   Filter by author name/email
  --me              Filter using ~/.config/gitcli/authors patterns
  --tz OFFSET       Timezone offset for day boundary (e.g. +09:00 for KST)
  --summary         Compact output: repo names + commit counts only (no details)
`, Version)
}
