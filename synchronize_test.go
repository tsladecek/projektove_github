package projektove

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchIssues(t *testing.T) {
	tests := []struct {
		name       string
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
					Description: fmt.Sprintf("%s: owner/my-repo\n\nNeed to add login page", PrefixGithubRepository),
				},
			},
			github: nil,
			want: []ProjektoveGithub{
				{
					Projektove: ProjektoveIssue{
						ID:          42,
						Subject:     "Implement login",
						Description: fmt.Sprintf("%s: owner/my-repo\n\nNeed to add login page", PrefixGithubRepository),
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
					Description: fmt.Sprintf("%s: owner/repo\n%s: 99\nhttps://github.com/owner/repo/issues/99\n\nFix the crashing bug", PrefixGithubRepository, PrefixGithubIssueID),
					AssignedTo:  &ProjektoveUser{ID: 1},
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
						Description: fmt.Sprintf("%s: owner/repo\n%s: 99\nhttps://github.com/owner/repo/issues/99\n\nFix the crashing bug", PrefixGithubRepository, PrefixGithubIssueID),
						AssignedTo:  &ProjektoveUser{ID: 1},
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
					Description: fmt.Sprintf("%s: org/other\nhttps://github.com/org/other/issues/999", PrefixGithubRepository),
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
						Description: fmt.Sprintf("%s: org/other\nhttps://github.com/org/other/issues/999", PrefixGithubRepository),
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
					Description: fmt.Sprintf("%s: a/repo\n%s: 10", PrefixGithubRepository, PrefixGithubIssueID),
				},
				{
					ID:          2,
					Subject:     "B",
					Description: fmt.Sprintf("%s: b/repo\nhttps://github.com/b/repo/issues/20", PrefixGithubRepository),
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
					Projektove: ProjektoveIssue{ID: 1, Subject: "A", Description: fmt.Sprintf("%s: a/repo\n%s: 10", PrefixGithubRepository, PrefixGithubIssueID)},
					Github:     &GithubIssue{ID: 10, Title: "A", State: GithubIssueStateOpen},
				},
				{
					Projektove: ProjektoveIssue{ID: 2, Subject: "B", Description: fmt.Sprintf("%s: b/repo\nhttps://github.com/b/repo/issues/20", PrefixGithubRepository)},
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
					require.NotNil(t, g.Github, "expected non-nil Github at index %d", i)
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

type updateCallProjektove struct {
	issueID int
	update  ProjektoveIssueUpdate
}

type updateCallGithub struct {
	issueID int
	update  GithubIssueUpdate
}

func TestSynchronize(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                 string
		projektoveIssues     []ProjektoveIssue
		githubIssues         map[string][]GithubIssue
		users                Users
		listIssuesErr        bool
		wantCreates          []createCall
		wantUpdateProjektove []updateCallProjektove
		wantUpdateGithub     []updateCallGithub
		wantErr              bool
	}{
		{
			name:                 "no issues",
			projektoveIssues:     nil,
			wantCreates:          nil,
			wantUpdateProjektove: nil,
		},
		{
			name: "no repo marker",
			projektoveIssues: []ProjektoveIssue{
				{ID: 1, Subject: "normal task", Description: "nothing to sync"},
			},
			wantCreates:          nil,
			wantUpdateProjektove: nil,
		},
		{
			name: "create new github issue",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          10,
					Subject:     "Implement login",
					Description: fmt.Sprintf("%s: owner/repo\n\nNeed login page", PrefixGithubRepository),
				},
			},
			githubIssues: nil,
			wantCreates: []createCall{
				{repo: "owner/repo", issue: GithubIssueCreate{
					Title: "Implement login",
					Body:  fmt.Sprintf("%s: owner/repo\n\nNeed login page\n\nhttps://app.projektove.cz//tasks/10", PrefixGithubRepository),
				}},
			},
			wantUpdateProjektove: []updateCallProjektove{
				{issueID: 10, update: ProjektoveIssueUpdate{Description: fmt.Sprintf("%s: owner/repo\n\nNeed login page\n\n%s: \n\n%s: 999", PrefixGithubRepository, PrefixGithubIssueURL, PrefixGithubIssueID)}},
			},
		},
		{
			name: "close projektove when github issue is closed",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          20,
					Subject:     "Done task",
					Description: fmt.Sprintf("%s: owner/repo\n%s: 5", PrefixGithubRepository, PrefixGithubIssueID),
					Status:      ProjektoveIssueStatus{IsClosed: false},
				},
			},
			githubIssues: map[string][]GithubIssue{
				"owner/repo": {
					{ID: 5, Title: "Done task", State: GithubIssueStateClosed},
				},
			},
			wantCreates: nil,
			wantUpdateProjektove: []updateCallProjektove{
				{issueID: 20, update: ProjektoveIssueUpdate{StatusID: int(ProjektoveStatusClosed)}},
			},
		},
		{
			name: "both open does nothing",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          30,
					Subject:     "WIP",
					Description: fmt.Sprintf("%s: owner/repo\n%s: 7", PrefixGithubRepository, PrefixGithubIssueID),
					Status:      ProjektoveIssueStatus{IsClosed: false},
				},
			},
			githubIssues: map[string][]GithubIssue{
				"owner/repo": {
					{ID: 7, Title: "WIP", State: GithubIssueStateOpen},
				},
			},
			wantCreates:          nil,
			wantUpdateProjektove: nil,
		},
		{
			name: "both closed does nothing",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          40,
					Subject:     "Already done",
					Description: fmt.Sprintf("%s: owner/repo\n%s: 9", PrefixGithubRepository, PrefixGithubIssueID),
					Status:      ProjektoveIssueStatus{IsClosed: true},
				},
			},
			githubIssues: map[string][]GithubIssue{
				"owner/repo": {
					{ID: 9, Title: "Already done", State: GithubIssueStateClosed},
				},
			},
			wantCreates:          nil,
			wantUpdateProjektove: nil,
		},
		{
			name: "mixed create and close",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          100,
					Subject:     "New feature",
					Description: fmt.Sprintf("%s: team/proj\n\nBrand new", PrefixGithubRepository),
				},
				{
					ID:          101,
					Subject:     "Old feature",
					Description: fmt.Sprintf("%s: team/proj\n%s: 2", PrefixGithubRepository, PrefixGithubIssueID),
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
					Body:  fmt.Sprintf("%s: team/proj\n\nBrand new\n\nhttps://app.projektove.cz//tasks/100", PrefixGithubRepository),
				}},
			},
			wantUpdateProjektove: []updateCallProjektove{
				{issueID: 101, update: ProjektoveIssueUpdate{StatusID: int(ProjektoveStatusClosed)}},
				{issueID: 100, update: ProjektoveIssueUpdate{Description: fmt.Sprintf("%s: team/proj\n\nBrand new\n\n%s: \n\n%s: 999", PrefixGithubRepository, PrefixGithubIssueURL, PrefixGithubIssueID)}},
			},
		},
		{
			name: "create with assignee from username map",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          200,
					Subject:     "Assigned task",
					Description: fmt.Sprintf("%s: team/repo", PrefixGithubRepository),
					AssignedTo:  &ProjektoveUser{ID: 1},
				},
			},
			users: Users{{Github: GithubUser{Login: "login"}, Projektove: ProjektoveUser{ID: 1}}},
			wantCreates: []createCall{
				{repo: "team/repo", issue: GithubIssueCreate{
					Title:     "Assigned task",
					Body:      fmt.Sprintf("%s: team/repo\n\nhttps://app.projektove.cz//tasks/200", PrefixGithubRepository),
					Assignees: []string{"login"},
				}},
			},
			wantUpdateProjektove: []updateCallProjektove{
				{issueID: 200, update: ProjektoveIssueUpdate{Description: fmt.Sprintf("%s: team/repo\n\n%s: \n\n%s: 999", PrefixGithubRepository, PrefixGithubIssueURL, PrefixGithubIssueID)}},
			},
		},
		{
			name:          "list issues error",
			listIssuesErr: true,
			wantErr:       true,
		},
		{
			name: "issue closed in projektove but not in github",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          200,
					Subject:     "Assigned task",
					Description: fmt.Sprintf("%s: team/proj\n\n%s: 2", PrefixGithubRepository, PrefixGithubIssueID),
					AssignedTo:  &ProjektoveUser{ID: 1},
					Status:      ProjektoveIssueStatus{ID: int(ProjektoveStatusClosed), IsClosed: true},
				},
			},
			githubIssues: map[string][]GithubIssue{
				"team/proj": {
					{ID: 2, Title: "Old feature", State: GithubIssueStateOpen},
				},
			},
			users: Users{{Github: GithubUser{Login: "login"}, Projektove: ProjektoveUser{ID: 1}}},
			wantUpdateGithub: []updateCallGithub{
				{issueID: 2, update: GithubIssueUpdate{Body: IssueClosedInProjektove}},
			},
		},
		{
			name: "issue closed in projektove but not in github, but github issue contains the string in description",
			projektoveIssues: []ProjektoveIssue{
				{
					ID:          200,
					Subject:     "Assigned task",
					Description: fmt.Sprintf("%s: team/proj\n\n%s: 2", PrefixGithubRepository, PrefixGithubIssueID),
					AssignedTo:  &ProjektoveUser{ID: 1},
					Status:      ProjektoveIssueStatus{ID: int(ProjektoveStatusClosed), IsClosed: true},
				},
			},
			githubIssues: map[string][]GithubIssue{
				"team/proj": {
					{ID: 2, Title: "Old feature", State: GithubIssueStateOpen, Body: "body" + IssueClosedInProjektove},
				},
			},
			users: Users{{Github: GithubUser{Login: "login"}, Projektove: ProjektoveUser{ID: 1}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			var gotCreates []createCall
			var gotUpdatesProjektove []updateCallProjektove
			var gotUpdatesGithub []updateCallGithub

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
			mockGH.UpdateIssueFunc = func(ctx context.Context, repository string, id int, issue GithubIssueUpdate) error {
				mu.Lock()
				gotUpdatesGithub = append(gotUpdatesGithub, updateCallGithub{issueID: id, update: issue})
				mu.Unlock()
				return nil
			}

			mockProj.UpdateIssueFunc = func(ctx context.Context, issueID int, obj ProjektoveIssueUpdate) error {
				mu.Lock()
				gotUpdatesProjektove = append(gotUpdatesProjektove, updateCallProjektove{issueID: issueID, update: obj})
				mu.Unlock()
				return nil
			}

			err := Synchronize(ctx, mockProj, mockGH, tt.users, false)

			sort.Slice(gotCreates, func(i, j int) bool {
				return gotCreates[i].repo < gotCreates[j].repo
			})
			sort.Slice(gotUpdatesProjektove, func(i, j int) bool {
				return gotUpdatesProjektove[i].issueID < gotUpdatesProjektove[j].issueID
			})
			sort.Slice(gotUpdatesGithub, func(i, j int) bool {
				return gotUpdatesGithub[i].issueID < gotUpdatesGithub[j].issueID
			})
			sort.Slice(tt.wantCreates, func(i, j int) bool {
				return tt.wantCreates[i].repo < tt.wantCreates[j].repo
			})
			sort.Slice(tt.wantUpdateProjektove, func(i, j int) bool {
				return tt.wantUpdateProjektove[i].issueID < tt.wantUpdateProjektove[j].issueID
			})
			sort.Slice(tt.wantUpdateGithub, func(i, j int) bool {
				return tt.wantUpdateGithub[i].issueID < tt.wantUpdateGithub[j].issueID
			})

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

			if tt.wantUpdateProjektove == nil {
				assert.Empty(t, gotUpdatesProjektove, "expected no close calls")
			} else {
				require.Len(t, gotUpdatesProjektove, len(tt.wantUpdateProjektove))
				for i, w := range tt.wantUpdateProjektove {
					assert.Equal(t, w.issueID, gotUpdatesProjektove[i].issueID, "close issue ID mismatch at %d", i)
					assert.Equal(t, w.update, gotUpdatesProjektove[i].update, "close update mismatch at %d", i)
				}
			}

			if tt.wantUpdateGithub == nil {
				assert.Empty(t, gotUpdatesGithub, "expected no update calls")
			} else {
				require.Len(t, gotUpdatesGithub, len(tt.wantUpdateGithub))
				for i, w := range tt.wantUpdateGithub {
					assert.Equal(t, w.issueID, gotUpdatesGithub[i].issueID, "update issue ID mismatch at %d", i)
					assert.Equal(t, w.update, gotUpdatesGithub[i].update, "update mismatch at %d", i)
				}
			}
		})
	}
}
