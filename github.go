package projektove

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const defaultGithubURL = "https://api.github.com"

type GithubAPI struct {
	client  *Client
	baseURL string
	token   string
}

func NewGithubAPI(baseURL, token string, client *Client) (Github, error) {
	if client == nil {
		return nil, fmt.Errorf("no http client provided")
	}
	if baseURL == "" {
		baseURL = defaultGithubURL
	}
	return GithubAPI{
		baseURL: baseURL,
		token:   token,
		client:  client,
	}, nil
}

func (g GithubAPI) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
}

func (g GithubAPI) repoURL(repository, endpoint string) string {
	return g.baseURL + "/repos/" + repository + "/" + endpoint
}

func (g GithubAPI) ListIssues(ctx context.Context, repository string) ([]GithubIssue, error) {
	u := g.repoURL(repository, "issues") + "?state=all"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create list issues request: %w", err)
	}
	g.setHeaders(req)

	var issues []GithubIssue
	if err := g.client.DoRaw(req, &issues); err != nil {
		return nil, fmt.Errorf("list issues for %s: %w", repository, err)
	}

	return issues, nil
}

func (g GithubAPI) CreateIssue(ctx context.Context, repository string, issue GithubIssueCreate) (GithubIssue, error) {
	u := g.repoURL(repository, "issues")

	data, err := json.Marshal(issue)
	if err != nil {
		return GithubIssue{}, fmt.Errorf("marshal create body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return GithubIssue{}, fmt.Errorf("create create issue request: %w", err)
	}
	g.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	var created GithubIssue
	if err := g.client.DoRaw(req, &created); err != nil {
		return GithubIssue{}, fmt.Errorf("create issue in %s: %w", repository, err)
	}

	return created, nil
}

func (g GithubAPI) UpdateIssue(ctx context.Context, repository string, id int, issue GithubIssueUpdate) error {
	u := g.repoURL(repository, fmt.Sprintf("issues/%d", id))

	data, err := json.Marshal(issue)
	if err != nil {
		return fmt.Errorf("marshal update body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create update issue request: %w", err)
	}
	g.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	if err := g.client.DoRaw(req, nil); err != nil {
		return fmt.Errorf("update issue #%d in %s: %w", id, repository, err)
	}

	return nil
}

func (g GithubAPI) GetPullRequest(ctx context.Context, repository string, id int) (GithubPullRequest, error) {
	u := g.repoURL(repository, fmt.Sprintf("pulls/%d", id))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return GithubPullRequest{}, fmt.Errorf("create get pull request: %w", err)
	}
	g.setHeaders(req)

	var pr GithubPullRequest
	if err := g.client.DoRaw(req, &pr); err != nil {
		return GithubPullRequest{}, fmt.Errorf("get pull request #%d from %s: %w", id, repository, err)
	}

	return pr, nil
}
