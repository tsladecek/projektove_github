package projektove

import (
	"context"
)

type Projektove interface {
	ListIssues(ctx context.Context) ([]ProjektoveIssue, error)
	UpdateIssue(ctx context.Context, issueID int, obj ProjektoveIssueUpdate) error
}

type Github interface {
	ListIssues(ctx context.Context, repository string) ([]GithubIssue, error)
	CreateIssue(ctx context.Context, repository string, issue GithubIssueCreate) (GithubIssue, error)
	UpdateIssue(ctx context.Context, repository string, id int, issue GithubIssueUpdate) error
	GetPullRequest(ctx context.Context, repository string, id int) (GithubPullRequest, error)
}
