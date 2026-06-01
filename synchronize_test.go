package projektove

import (
	"context"
	"fmt"
	"sync"
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

type createCall struct {
	repo  string
	issue GithubIssueCreate
}

type closeCall struct {
	issueID int
	update  ProjektoveIssueUpdate
}

func TestSynchronize(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		projektoveIssues []ProjektoveIssue
		githubIssues     map[string][]GithubIssue
		usernameMap      map[string]int
		listIssuesErr    bool
		wantCreates      []createCall
		wantCloses       []closeCall
		wantErr          bool
	}{
		{
			name:             "no issues",
			projektoveIssues: nil,
			wantCreates:      nil,
			wantCloses:       nil,
		},
		{
			name: "no repo marker",
			projektoveIssues: []ProjektoveIssue{
				{ID: 1, Subject: "normal task", Description: "nothing to sync"},
			},
			wantCreates: nil,
			wantCloses:  nil,
		},
		{
			name: "create new github issue",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          10,
					Subject:     "Implement login",
					Description: "GitHub Repository: owner/repo\n\nNeed login page",
				},
			},
			githubIssues: nil,
			wantCreates: []createCall{
				{repo: "owner/repo", issue: GithubIssueCreate{
					Title: "Implement login",
					Body:  "GitHub Repository: owner/repo\n\nNeed login page",
				}},
			},
			wantCloses: nil,
		},
		{
			name: "close projektove when github issue is closed",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          20,
					Subject:     "Done task",
					Description: "GitHub Repo: owner/repo\nhttps://github.com/owner/repo/issues/5",
					Status:      ProjektoveIssueStatus{IsClosed: false},
				},
			},
			githubIssues: map[string][]GithubIssue{
				"owner/repo": {
					{ID: 5, Title: "Done task", State: GithubIssueStateClosed},
				},
			},
			wantCreates: nil,
			wantCloses: []closeCall{
				{issueID: 20, update: ProjektoveIssueUpdate{StatusID: int(ProjektoveStatusClosed)}},
			},
		},
		{
			name: "both open does nothing",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          30,
					Subject:     "WIP",
					Description: "GitHub Repo: owner/repo\nhttps://github.com/owner/repo/issues/7",
					Status:      ProjektoveIssueStatus{IsClosed: false},
				},
			},
			githubIssues: map[string][]GithubIssue{
				"owner/repo": {
					{ID: 7, Title: "WIP", State: GithubIssueStateOpen},
				},
			},
			wantCreates: nil,
			wantCloses:  nil,
		},
		{
			name: "both closed does nothing",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          40,
					Subject:     "Already done",
					Description: "GitHub Repo: owner/repo\nhttps://github.com/owner/repo/issues/9",
					Status:      ProjektoveIssueStatus{IsClosed: true},
				},
			},
			githubIssues: map[string][]GithubIssue{
				"owner/repo": {
					{ID: 9, Title: "Already done", State: GithubIssueStateClosed},
				},
			},
			wantCreates: nil,
			wantCloses:  nil,
		},
		{
			name: "mixed create and close",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          100,
					Subject:     "New feature",
					Description: "GitHub Repository: team/proj\n\nBrand new",
				},
				{
					ID:          101,
					Subject:     "Old feature",
					Description: "GitHub Repo: team/proj\nhttps://github.com/team/proj/issues/2",
					Status:      ProjektoveIssueStatus{IsClosed: false},
				},
			},
			githubIssues: map[string][]GithubIssue{
				"team/proj": {
					{ID: 2, Title: "Old feature", State: GithubIssueStateClosed},
				},
			},
			wantCreates: []createCall{
				{repo: "team/proj", issue: GithubIssueCreate{
					Title: "New feature",
					Body:  "GitHub Repository: team/proj\n\nBrand new",
				}},
			},
			wantCloses: []closeCall{
				{issueID: 101, update: ProjektoveIssueUpdate{StatusID: int(ProjektoveStatusClosed)}},
			},
		},
		{
			name: "create with assignee from username map",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          200,
					Subject:     "Assigned task",
					Description: "GitHub Repo: team/repo",
					AssignedTo:  ProjectoveUser{ID: 736, Name: "Tomas"},
				},
			},
			usernameMap: map[string]int{"tomas": 736},
			wantCreates: []createCall{
				{repo: "team/repo", issue: GithubIssueCreate{
					Title:     "Assigned task",
					Body:      "GitHub Repo: team/repo",
					Assignees: []GithubUser{{Login: "tomas"}},
				}},
			},
			wantCloses: nil,
		},
		{
			name:          "list issues error",
			listIssuesErr: true,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			var gotCreates []createCall
			var gotCloses []closeCall

			mockProj := &MockProjektove{}
			mockGH := &MockGithub{}

			if tt.listIssuesErr {
				mockProj.ListIssuesFunc = func(ctx context.Context) ([]ProjektoveIssue, error) {
					return nil, fmt.Errorf("api error")
				}
			} else {
				mockProj.ListIssuesFunc = func(ctx context.Context) ([]ProjektoveIssue, error) {
					return tt.projektoveIssues, nil
				}
			}

			mockGH.ListIssuesFunc = func(ctx context.Context, repo string) ([]GithubIssue, error) {
				if tt.githubIssues != nil {
					return tt.githubIssues[repo], nil
				}
				return nil, nil
			}

			mockGH.CreateIssueFunc = func(ctx context.Context, repo string, issue GithubIssueCreate) (GithubIssue, error) {
				mu.Lock()
				gotCreates = append(gotCreates, createCall{repo: repo, issue: issue})
				mu.Unlock()
				return GithubIssue{ID: 999, Title: issue.Title, State: GithubIssueStateOpen}, nil
			}

			mockProj.UpdateIssueFunc = func(ctx context.Context, issueID int, obj ProjektoveIssueUpdate) error {
				mu.Lock()
				gotCloses = append(gotCloses, closeCall{issueID: issueID, update: obj})
				mu.Unlock()
				return nil
			}

			err := Synchronize(ctx, mockProj, mockGH, tt.usernameMap)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if tt.wantCreates == nil {
				assert.Empty(t, gotCreates, "expected no create calls")
			} else {
				assert.Len(t, gotCreates, len(tt.wantCreates))
				for i, w := range tt.wantCreates {
					assert.Equal(t, w.repo, gotCreates[i].repo, "create repo mismatch at %d", i)
					assert.Equal(t, w.issue.Title, gotCreates[i].issue.Title, "create title mismatch at %d", i)
					assert.Equal(t, w.issue.Body, gotCreates[i].issue.Body, "create body mismatch at %d", i)
					if w.issue.Assignees == nil {
						assert.Nil(t, gotCreates[i].issue.Assignees, "expected nil assignees at %d", i)
					} else {
						assert.Equal(t, w.issue.Assignees, gotCreates[i].issue.Assignees, "assignees mismatch at %d", i)
					}
				}
			}

			if tt.wantCloses == nil {
				assert.Empty(t, gotCloses, "expected no close calls")
			} else {
				assert.Len(t, gotCloses, len(tt.wantCloses))
				for i, w := range tt.wantCloses {
					assert.Equal(t, w.issueID, gotCloses[i].issueID, "close issue ID mismatch at %d", i)
					assert.Equal(t, w.update, gotCloses[i].update, "close update mismatch at %d", i)
				}
			}
		})
	}
}
