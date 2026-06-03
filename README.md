# projektove_github

A CLI tool that synchronizes issues between Projektove (issue tracker) and GitHub.

## How it works

Projektove is the source of truth. Issues marked with `GitHub Repository: owner/repo` in their description are synced to GitHub.

- **GitHub issue closes → Projektove issue closes**: When a GitHub issue is closed, the corresponding Projektove issue is also closed.
- **Projektove issue closes → GitHub description updated**: When a Projektove issue is closed, the GitHub issue is *not* closed (work may have started), but its description is appended with a note that the issue has been closed in Projektove.

## Usage

Download binary from the [releases](https://github.com/tsladecek/projektove_github/releases) page:

```sh
./projektove-github \
  --projektoveURL https://your-instance.projektove.cz \
  --projektoveToken YOUR_PROJEKTOVE_TOKEN \
  --githubToken YOUR_GITHUB_TOKEN \
  --users '[{"projektove": {"id": 1}, "github": {"login": "joegithub"}}]'
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
