# Projektove ↔ GitHub Issue Synchronizer

## Project Purpose

A CLI tool that synchronizes issues from **Projektove** (issue tracker) to **GitHub**. Projektove is the source of truth. Issues containing a `GitHub Repository: owner/repo` marker in their description (or subject) are synced. When a GitHub issue is closed, the corresponding Projektove issue is also closed. The sync is bidirectional but Projektove always wins for content; GitHub closing triggers a close in Projektove.

## Technology Stack

- **Language**: Go 1.26+
- **Dependencies**: stdlib only (`context`, `fmt`, `net/http`, `regexp`, `sync`, ...)
- **Testing**: `github.com/stretchr/testify` (assert + require), table-driven tests
- **No external HTTP routers or frameworks** — use `net/http` directly

## Project Structure

```
├── synchronize.go          # Core sync orchestration (MatchIssues, Synchronize)
├── synchronize_test.go     # Tests (table-driven)
├── model.go                # All data types & enums
├── interfaces.go           # Projektove + Github interface definitions
├── cmd/cli/main.go          # CLI entrypoint
├── data/                   # JSON fixtures for testing
│   ├── projektove_issue.json
│   └── projektove_enums.json
├── go.mod
└── AGENTS.md
```

## Build / Run / Test Commands

- **Build CLI**: `go build -o projektove-github ./cmd/cli/`
- **Run all tests**: `go test ./...`
- **Run specific test**: `go test -run TestMatchIssues ./...`
- **Lint**: `go vet ./...`

## Code Conventions

- Package name: `projektove`
- No external dependencies except testify (dev-only)
- All exported structs have JSON tags
- Use `context.Context` as first parameter in all interface methods
- Interface-based APIs to enable mocking in tests
- Enums use `iota` with explicit string representations where needed
- Error wrapping with `fmt.Errorf("context: %w", err)`
- Logging via `log` package or stdout; no logrus/zerolog

## Synchronization Logic

The `Synchronize` function operates in two phases:

1. **Phase 1 – Planning**: Fetch all Projektove issues, collect unique GitHub repos, fetch issues from each repo, call `MatchIssues`, and build a `map[string]func() error` of actions:
   - `create:owner/repo/PROJ_ID` — GitHub issue doesn't exist yet → create it
   - `close:owner/repo/PROJ_ID` — GitHub issue is closed but Projektove isn't → close in Projektove

2. **Phase 2 – Concurrent Execution**: Each action runs in its own goroutine. Print `[START]`, `[SUCCESS]`, or `[FAIL]` status lines (concurrency-safe via `sync.Mutex`). Wait for all with `sync.WaitGroup`.

### Issue Marker Format

Detected via regex from the Projektove issue description (or subject as fallback):

- **Repo marker**: `GitHub Repository: owner/repo` or `GitHub Repo: owner/repo`
- **Issue link**: `https://github.com/owner/repo/issues/N`

`model.go` exports `GithubRepository() string` and `GithubID() int` helper methods.

## Testing Approach

- **Table-driven tests** with testify
- Use `assert.Equal`, `assert.Nil`, `require.NoError`, etc.
- Mock `Projektove` and `Github` interfaces for unit tests
- Load fixtures from `data/*.json` where useful
- Test `MatchIssues` as a pure function (no IO)
- Test `Synchronize` with mock interfaces
- Each test function: `func TestXxx(t *testing.T) { ... }`
- Use subtests with `t.Run` for table cases
- In case mock data is needed, just create simple functions. Dont pull in additional deps
