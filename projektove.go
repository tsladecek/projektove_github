package projektove

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type ProjektoveAPI struct {
	client  *http.Client
	baseURL *url.URL
	token   string
}

func NewProjektoveAPI(baseURL, token string, client *http.Client) (Projektove, error) {
	burl, err := url.Parse(baseURL)
	if err != nil {
		return ProjektoveAPI{}, fmt.Errorf("when parsing projektove url %q: %w", baseURL, err)
	}

	if client == nil {
		return ProjektoveAPI{}, fmt.Errorf("no http client provided")
	}
	return ProjektoveAPI{
		baseURL: burl,
		token:   token,
		client:  client,
	}, nil
}

// endpoint: GET /issues.json
func (p ProjektoveAPI) ListIssues(ctx context.Context) ([]ProjektoveIssue, error) {
	return nil, nil
}

// endpoint: PUT /issues/{issueID}.json
// body (json): {"issue": {"status_id": <status ID>}}
// Header: Content-Type: application/json
// Header: X-API-Authorization: <token>
func (p ProjektoveAPI) UpdateIssue(ctx context.Context, issueID int, obj ProjektoveIssueUpdate) error {
	return nil
}
