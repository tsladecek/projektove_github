# Implementation Plan - GitHub & Projektove Issue Synchronization

This plan describes how we will implement the synchronize functionality in `./synchronize.go` and update the models and interfaces.

## 1. Interface Updates (`interfaces.go`)
- Update `Github` interface:
  - Modify `CreateIssue` to return `(GithubIssue, error)`.
  - Add `IssueLink(repository string, id int) string` to return the issue's URL.

## 2. Model Updates (`model.go`)
- Add `ProjektoveStatus` iota enum:
  ```go
  type ProjektoveStatus int
  const (
      ProjektoveStatusNew ProjektoveStatus = iota
      ProjektoveStatusInProgress
      ProjektoveStatusResolved
      ProjektoveStatusFeedback
      ProjektoveStatusClosed
  )
  ```
- Update `ProjektoveIssueUpdate` struct to include `StatusID int json:"status_id"`.
- Update `GithubPullRequest` struct to include `ClosedBy string json:"closed_by"`.
- Implement `GithubID() int` to parse the GitHub issue ID from the Projektove issue description via regex.
- Implement `GithubRepository() string` to parse the GitHub repository from the description/subject.

## 3. Synchronize Functionality (`synchronize.go`)
- Signature: `func Synchronize(ctx context.Context, projektove Projektove, github Github, usernameMap map[string]int) error`
- **Phase 1: Planning**: Build a `plan map[string]func() error` where each key represents a synchronization task.
- **Phase 2: Concurrent Execution**: Concurrently execute each planned action in goroutines, printing `[START]`, `[SUCCESS]`, or `[FAIL]` status updates to stdout, protecting concurrency with a `sync.Mutex` and `sync.WaitGroup`.
