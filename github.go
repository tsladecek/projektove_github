package projektove

import (
	"context"
	"net/http"
)

type GithubAPI struct {
	client *http.Client
	token  string
}

func NewGithubAPI(baseURL, token string, client *http.Client) (Github, error) {
	return GithubAPI{
		token:  token,
		client: client,
	}, nil
}

//	curl -L \
//	  -H "Accept: application/vnd.github+json" \
//	  -H "Authorization: Bearer <YOUR-TOKEN>" \
//	  -H "X-GitHub-Api-Version: 2026-03-10" \
//	  https://api.github.com/repos/OWNER/REPO/issues
func (p GithubAPI) ListIssues(ctx context.Context, repository string) ([]GithubIssue, error) {
	return nil, nil
}

//	curl -L \
//	  -X POST \
//	  -H "Accept: application/vnd.github+json" \
//	  -H "Authorization: Bearer <YOUR-TOKEN>" \
//	  -H "X-GitHub-Api-Version: 2026-03-10" \
//	  https://api.github.com/repos/OWNER/REPO/issues \
//	  -d '{"title":"Found a bug","body":"I'\''m having a problem with this.","assignees":["octocat"],"milestone":1,"labels":["bug"]}'
func (p GithubAPI) CreateIssue(ctx context.Context, repository string, issue GithubIssueCreate) (GithubIssue, error) {
	return GithubIssue{}, nil
}

//	curl -L \
//	  -X PATCH \
//	  -H "Accept: application/vnd.github+json" \
//	  -H "Authorization: Bearer <YOUR-TOKEN>" \
//	  -H "X-GitHub-Api-Version: 2026-03-10" \
//	  https://api.github.com/repos/OWNER/REPO/issues/ISSUE_NUMBER \
//	  -d '{"title":"Found a bug","body":"I'\''m having a problem with this.","assignees":["octocat"],"milestone":1,"state":"open","labels":["bug"]}'
func (p GithubAPI) UpdateIssue(ctx context.Context, repository string, id int, issue GithubIssueUpdate) error {
	return nil
}

//	curl -L \
//	  -H "Accept: application/vnd.github+json" \
//	  -H "Authorization: Bearer <YOUR-TOKEN>" \
//	  -H "X-GitHub-Api-Version: 2026-03-10" \
//	  https://api.github.com/repos/OWNER/REPO/pulls/PULL_NUMBER
func (p GithubAPI) GetPullRequest(ctx context.Context, repository string, id int) (GithubPullRequest, error) {
	return GithubPullRequest{}, nil
}
