// repos_cmd.go — gitcli repos: list all repos with stats
package main

type ReposResult struct {
	TotalRepos int        `json:"total_repos"`
	Repos      []RepoInfo `json:"repos"`
}

type RepoInfo struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	FirstCommit   string `json:"first_commit"`
	LastCommit    string `json:"last_commit"`
	TotalCommits  int    `json:"total_commits"`
	CurrentBranch string `json:"current_branch"`
}

// QueryReposInfo returns info about all discovered repos.
func QueryReposInfo(repos []Repo) ReposResult {
	result := ReposResult{}

	for _, repo := range repos {
		info := RepoInfo{
			Name:          repo.Name,
			Path:          repo.Path,
			CurrentBranch: gitCurrentBranch(repo.Path),
		}

		// Total commits
		info.TotalCommits = gitRevListCount(repo.Path, "HEAD")

		// First commit date (oldest)
		lines := gitLog(repo.Path, "--format=%ad", "--date=short", "--diff-filter=A", "--follow", "--reverse")
		if len(lines) > 0 {
			info.FirstCommit = lines[0]
		} else {
			// Fallback: rev-list to get root commit
			firstLines := gitLog(repo.Path, "--format=%ad", "--date=short", "--reverse")
			if len(firstLines) > 0 {
				info.FirstCommit = firstLines[0]
			}
		}

		// Last commit date (newest)
		lines = gitLog(repo.Path, "--format=%ad", "--date=short", "-1")
		if len(lines) > 0 {
			info.LastCommit = lines[0]
		}

		result.Repos = append(result.Repos, info)
	}

	result.TotalRepos = len(result.Repos)
	return result
}
