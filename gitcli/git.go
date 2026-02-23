// git.go — git command wrappers and repo discovery
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Repo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// DiscoverRepos finds all git repos under given directories (1-depth).
func DiscoverRepos(dirs []string) []Repo {
	var repos []Repo
	for _, dir := range dirs {
		dir = expandHome(dir)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			p := filepath.Join(dir, e.Name())
			if isGitRepo(p) {
				repos = append(repos, Repo{Name: e.Name(), Path: p})
			}
		}
	}
	return repos
}

func isGitRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		return false
	}
	return info.IsDir()
}

// gitLog runs git log with given format and args, returns lines.
func gitLog(repoPath string, extraArgs ...string) []string {
	args := append([]string{"-C", repoPath, "log"}, extraArgs...)
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// gitRevList returns line count (commit count).
func gitRevListCount(repoPath string, extraArgs ...string) int {
	args := append([]string{"-C", repoPath, "rev-list", "--count"}, extraArgs...)
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	s := strings.TrimSpace(string(out))
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

// gitCurrentBranch returns the current branch name.
func gitCurrentBranch(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[2:])
	}
	return p
}
