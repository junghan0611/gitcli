# gitcli

Local git timeline CLI for AI agents. Queries commit history across multiple repositories.

## Install

```bash
./run.sh
```

## Commands

```
gitcli day [DATE]       — All commits for a specific date across repos
gitcli repos            — List all repos with stats
gitcli log <repo-name>  — Repo-specific commit log
gitcli timeline         — Activity overview for a period
```

## Examples

```bash
# Today's commits
gitcli day

# Specific date
gitcli day 2025-10-10

# 1 year ago today
gitcli day --years-ago 1

# All repos info
gitcli repos --repos ~/repos/gh,~/repos/work

# Recent 7 days of a repo
gitcli log pi-mono --days 7

# Monthly overview
gitcli timeline --month 2026-02
```

All output is JSON for agent consumption.
