# projektove_github

A CLI tool that synchronizes issues between Projektove (issue tracker) and GitHub.

## How it works

Projektove is the source of truth. Issues marked with `GitHub Repository: owner/repo` in their description are synced to GitHub. When a GitHub issue is closed, the corresponding Projektove issue is also closed.

The sync is bidirectional for the close action, but Projektove always wins for content.

## Usage

```sh
go build -o projektove-github ./cmd/cli/

./projektove-github \
  --projektoveURL https://your-instance.projektove.cz \
  --projektoveToken YOUR_PROJEKTOVE_TOKEN \
  --githubToken YOUR_GITHUB_TOKEN \
  --users '[{"projektove": {"id": 1, "name": "joe"}, "github": {"id": 42, "login": "joegithub"}}]'
```

### Flags

| Flag | Description | Required |
|------|-------------|----------|
| `--projektoveURL` | Projektove base URL | yes |
| `--projektoveToken` | Projektove API token | yes |
| `--githubToken` | GitHub API token | yes |
| `--githubURL` | GitHub API base URL (default `https://api.github.com`) | no |
| `--users` | JSON list mapping Projektove users to GitHub users | no |
| `--dryRun` | Log actions without executing them | no |

## Issue marker format

To link a Projektove issue to GitHub, add these lines to its description:

```
GitHub Repository: owner/repo
```

After the first sync, the tool appends the GitHub issue URL and ID:

```
GitHub Issue URL: https://github.com/owner/repo/issues/123
GitHub Issue ID: 123
```

## Build

```
go build -o projektove-github ./cmd/cli/
```

## Test

```
go test ./...
```
