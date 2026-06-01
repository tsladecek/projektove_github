package projektove

import "context"

type MockProjektove struct {
	ListIssuesFunc  func(ctx context.Context) ([]ProjektoveIssue, error)
	UpdateIssueFunc func(ctx context.Context, issueID int, obj ProjektoveIssueUpdate) error
}

func (m *MockProjektove) ListIssues(ctx context.Context) ([]ProjektoveIssue, error) {
	if m.ListIssuesFunc != nil {
		return m.ListIssuesFunc(ctx)
	}
	return nil, nil
}

func (m *MockProjektove) UpdateIssue(ctx context.Context, issueID int, obj ProjektoveIssueUpdate) error {
	if m.UpdateIssueFunc != nil {
		return m.UpdateIssueFunc(ctx, issueID, obj)
	}
	return nil
}

type MockGithub struct {
	ListIssuesFunc     func(ctx context.Context, repository string) ([]GithubIssue, error)
	CreateIssueFunc    func(ctx context.Context, repository string, issue GithubIssueCreate) (GithubIssue, error)
	UpdateIssueFunc    func(ctx context.Context, repository string, id int, issue GithubIssueUpdate) error
	GetPullRequestFunc func(ctx context.Context, repository string, id int) (GithubPullRequest, error)
}

func (m *MockGithub) ListIssues(ctx context.Context, repository string) ([]GithubIssue, error) {
	if m.ListIssuesFunc != nil {
		return m.ListIssuesFunc(ctx, repository)
	}
	return nil, nil
}

func (m *MockGithub) CreateIssue(ctx context.Context, repository string, issue GithubIssueCreate) (GithubIssue, error) {
	if m.CreateIssueFunc != nil {
		return m.CreateIssueFunc(ctx, repository, issue)
	}
	return GithubIssue{}, nil
}

func (m *MockGithub) UpdateIssue(ctx context.Context, repository string, id int, issue GithubIssueUpdate) error {
	if m.UpdateIssueFunc != nil {
		return m.UpdateIssueFunc(ctx, repository, id, issue)
	}
	return nil
}

func (m *MockGithub) GetPullRequest(ctx context.Context, repository string, id int) (GithubPullRequest, error) {
	if m.GetPullRequestFunc != nil {
		return m.GetPullRequestFunc(ctx, repository, id)
	}
	return GithubPullRequest{}, nil
}
