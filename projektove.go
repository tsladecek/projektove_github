package projektove

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type ProjektoveAPI struct {
	client  *Client
	baseURL *url.URL
	token   string
}

type projektoveListResponse struct {
	Issues []ProjektoveIssue `json:"issues"`
}

func (plr *projektoveListResponse) fixDescriptions() {
	for i := range plr.Issues {
		plr.Issues[i].Description = strings.ReplaceAll(plr.Issues[i].Description, " ", " ")
	}
}

type projektoveUpdateBody struct {
	Issue ProjektoveIssueUpdate `json:"issue"`
}

func NewProjektoveAPI(baseURL, token string, client *Client) (Projektove, error) {
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

func (p ProjektoveAPI) listIssues(ctx context.Context, status *ProjektoveStatus) ([]ProjektoveIssue, error) {
	u := p.baseURL.JoinPath("issues.json").String()

	if status != nil {
		u = u + "?status_id=" + strconv.Itoa(int(*status))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create list issues request: %w", err)
	}
	req.Header.Set("X-API-Authorization", p.token)

	var resp projektoveListResponse
	if _, err := p.client.Do(req, &resp); err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}

	resp.fixDescriptions()

	return resp.Issues, nil
}

func (p ProjektoveAPI) ListIssues(ctx context.Context) ([]ProjektoveIssue, error) {
	open, err := p.listIssues(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("when listing open projektove issues: %w", err)
	}
	closed, err := p.listIssues(ctx, new(ProjektoveStatusClosed))
	if err != nil {
		return nil, fmt.Errorf("when listing closed projektove issues: %w", err)
	}

	return append(open, closed...), nil
}

func (p ProjektoveAPI) UpdateIssue(ctx context.Context, issueID int, obj ProjektoveIssueUpdate) error {
	u := p.baseURL.JoinPath(fmt.Sprintf("issues/%d.json", issueID)).String()

	body := projektoveUpdateBody{
		Issue: obj,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal update body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create update issue request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Authorization", p.token)

	if _, err := p.client.Do(req, nil); err != nil {
		return fmt.Errorf("update issue #%d: %w", issueID, err)
	}

	return nil
}
