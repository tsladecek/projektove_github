package projektove

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchIssues(t *testing.T) {
	tests := []struct {
		name      string
		projektove []ProjektoveIssue
		github     map[string][]GithubIssue
		want       []ProjektoveGithub
	}{
		{
			name:       "empty input",
			projektove: nil,
			github:     nil,
			want:       nil,
		},
		{
			name: "no repo marker in description or subject",
			projektove: []ProjektoveIssue{
				{ID: 1, Subject: "task", Description: "just a normal issue"},
				{ID: 2, Subject: "another", Description: "GitHub is not mentioned here"},
			},
			github: nil,
			want:   nil,
		},
		{
			name: "repo marker but no github issue yet",
			projektove: []ProjektoveIssue{
				{
					ID:          42,
					Subject:     "Implement login",
					Description: "GitHub Repository: owner/my-repo\n\nNeed to add login page",
				},
			},
			github: nil,
			want: []ProjektoveGithub{
				{
					Projektove: ProjektoveIssue{
						ID:          42,
						Subject:     "Implement login",
						Description: "GitHub Repository: owner/my-repo\n\nNeed to add login page",
					},
					Github: nil,
				},
			},
		},
		{
			name: "repo marker with matching github issue",
			projektove: []ProjektoveIssue{
				{
					ID:          10,
					Subject:     "Fix bug",
					Description: "GitHub Repo: owner/repo\nhttps://github.com/owner/repo/issues/99\n\nFix the crashing bug",
					AssignedTo:  ProjectoveUser{ID: 1, Name: "tester"},
				},
			},
			github: map[string][]GithubIssue{
				"owner/repo": {
					{ID: 99, Title: "Fix bug", State: GithubIssueStateOpen},
					{ID: 100, Title: "Other issue", State: GithubIssueStateClosed},
				},
			},
			want: []ProjektoveGithub{
				{
					Projektove: ProjektoveIssue{
						ID:          10,
						Subject:     "Fix bug",
						Description: "GitHub Repo: owner/repo\nhttps://github.com/owner/repo/issues/99\n\nFix the crashing bug",
						AssignedTo:  ProjectoveUser{ID: 1, Name: "tester"},
					},
					Github: &GithubIssue{
						ID:    99,
						Title: "Fix bug",
						State: GithubIssueStateOpen,
					},
				},
			},
		},
		{
			name: "repo marker but github id does not match any issue in repo",
			projektove: []ProjektoveIssue{
				{
					ID:          5,
					Subject:     "Missing issue",
					Description: "GitHub Repository: org/other\nhttps://github.com/org/other/issues/999",
				},
			},
			github: map[string][]GithubIssue{
				"org/other": {
					{ID: 1, Title: "Existing", State: GithubIssueStateOpen},
				},
			},
			want: []ProjektoveGithub{
				{
					Projektove: ProjektoveIssue{
						ID:          5,
						Subject:     "Missing issue",
						Description: "GitHub Repository: org/other\nhttps://github.com/org/other/issues/999",
					},
					Github: nil,
				},
			},
		},
		{
			name: "repo marker in subject as fallback",
			projektove: []ProjektoveIssue{
				{
					ID:          7,
					Subject:     "GitHub Repo: fallback/repo",
					Description: "no marker here, but subject has it",
				},
			},
			github: nil,
			want: []ProjektoveGithub{
				{
					Projektove: ProjektoveIssue{
						ID:          7,
						Subject:     "GitHub Repo: fallback/repo",
						Description: "no marker here, but subject has it",
					},
					Github: nil,
				},
			},
		},
		{
			name: "multiple issues across different repos",
			projektove: []ProjektoveIssue{
				{
					ID:          1,
					Subject:     "A",
					Description: "GitHub Repository: a/repo\nhttps://github.com/a/repo/issues/10",
				},
				{
					ID:          2,
					Subject:     "B",
					Description: "GitHub Repo: b/repo\nhttps://github.com/b/repo/issues/20",
				},
				{
					ID:          3,
					Subject:     "skip me",
					Description: "no marker",
				},
			},
			github: map[string][]GithubIssue{
				"a/repo": {
					{ID: 10, Title: "A", State: GithubIssueStateOpen},
				},
				"b/repo": {},
			},
			want: []ProjektoveGithub{
				{
					Projektove: ProjektoveIssue{ID: 1, Subject: "A", Description: "GitHub Repository: a/repo\nhttps://github.com/a/repo/issues/10"},
					Github:     &GithubIssue{ID: 10, Title: "A", State: GithubIssueStateOpen},
				},
				{
					Projektove: ProjektoveIssue{ID: 2, Subject: "B", Description: "GitHub Repo: b/repo\nhttps://github.com/b/repo/issues/20"},
					Github:     nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchIssues(tt.projektove, tt.github)

			if tt.want == nil {
				assert.Nil(t, got, "expected nil result")
				return
			}

			assert.Len(t, got, len(tt.want))
			for i := range tt.want {
				w := tt.want[i]
				g := got[i]

				assert.Equal(t, w.Projektove.ID, g.Projektove.ID, "issue ID mismatch at index %d", i)
				assert.Equal(t, w.Projektove.Subject, g.Projektove.Subject, "subject mismatch at index %d", i)
				assert.Equal(t, w.Projektove.Description, g.Projektove.Description, "description mismatch at index %d", i)
				assert.Equal(t, w.Projektove.AssignedTo, g.Projektove.AssignedTo, "assignedTo mismatch at index %d", i)

				if w.Github == nil {
					assert.Nil(t, g.Github, "expected nil Github at index %d", i)
				} else {
					assert.NotNil(t, g.Github, "expected non-nil Github at index %d", i)
					assert.Equal(t, w.Github.ID, g.Github.ID, "Github ID mismatch at index %d", i)
					assert.Equal(t, w.Github.Title, g.Github.Title, "Github title mismatch at index %d", i)
					assert.Equal(t, w.Github.State, g.Github.State, "Github state mismatch at index %d", i)
				}
			}
		})
	}
}

func TestSynchronize(t *testing.T) {}
