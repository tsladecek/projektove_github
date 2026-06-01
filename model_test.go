package projektove

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjektoveIssueGithubRepository(t *testing.T) {
	cases := []struct {
		name       string
		issue      ProjektoveIssue
		repository string
		issueID    int
	}{
		{name: "should match", issue: ProjektoveIssue{Description: "this is a description\nGitHub Repository: org/repo"}, repository: "org/repo"},
		{name: "should match with github id", issue: ProjektoveIssue{Description: "this is a description\nGitHub Repository: org/repo\nGitHub Issue ID: 123"}, repository: "org/repo", issueID: 123},
		{name: "should match with underscores and hyphens", issue: ProjektoveIssue{Description: "this is a description\nGitHub Repository: org_1-2/repo-2_3"}, repository: "org_1-2/repo-2_3"},
		{name: "real world issue", issue: ProjektoveIssue{Description: "<p>GitHub Repository: org/repo</p>"}, repository: "org/repo"},
		{name: "should match without space", issue: ProjektoveIssue{Description: "this is a description\nGitHub Repository:org/repo"}, repository: "org/repo"},
		{name: "shouldnot match", issue: ProjektoveIssue{Description: "this is a description"}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.repository, tt.issue.GithubRepository())
			assert.Equal(t, tt.issueID, tt.issue.GithubID())
		})
	}
}

func TestProjektoveIssueUpdateOmit(t *testing.T) {
	p := ProjektoveIssueUpdate{Description: "desc"}
	m, err := json.Marshal(p)
	require.NoError(t, err)
	assert.Equal(t, `{"description":"desc"}`, string(m))
}
