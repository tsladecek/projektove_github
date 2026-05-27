package projektove

import (
	"time"
)

type ProjektoveProject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProjektoveIssueStatus struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsClosed bool   `json:"isClosed"`
	Role     string `json:"role"`
}

type ProjectoveUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProjektoveIssue struct {
	ID          int                   `json:"id"`
	Project     ProjektoveProject     `json:"project"`
	Status      ProjektoveIssueStatus `json:"status"`
	Author      ProjectoveUser        `json:"author"`
	AssignedTo  ProjectoveUser        `json:"assignedTo"`
	Subject     string                `json:"subject"`
	Description string                `json:"description"`
	DueDate     time.Time             `json:"dueDate"`
	CreatedOn   time.Time             `json:"createdOn"`
}

// determines whether this issue should be synced with github repository
// empty string means no
func (p ProjektoveIssue) GithubRepository() string {
	return ""
}

// returns id of the github issue
// this is filled when the issue is synchronized with github
// 0 means that the issue has not been yet created
func (p ProjektoveIssue) GithubID() int {
	return 0
}

type ProjektoveIssueUpdate struct {
	Subject      string    `json:"subject"`
	Description  string    `json:"description"`
	ProjectID    int       `json:"project_id"`
	StartDate    time.Time `json:"start_date"`
	DueDate      time.Time `json:"due_date"`
	AuthorID     int       `json:"author_id"`
	AssignedToID int       `json:"assigned_to_id"`
}

type GithubIssueState string

const (
	GithubIssueStateOpen   GithubIssueState = "open"
	GithubIssueStateClosed GithubIssueState = "closed"
)

type GithubPullRequestState string

const (
	GithubPullRequestStateOpen   GithubPullRequestState = "open"
	GithubPullRequestStateClosed GithubPullRequestState = "closed"
)

type GithubIssue struct {
	ID        int              `json:"id"`
	Title     string           `json:"title"`
	Body      string           `json:"body"`
	Assignees []string         `json:"assignees"`
	State     GithubIssueState `json:"state"` // either open or closed
}

type GithubIssueCreate struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Assignees []string `json:"assignees"`
}

type GithubIssueUpdate struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Assignees []string `json:"assignees"`
}

type GithubPullRequest struct {
	ID        int                    `json:"id"`
	State     GithubPullRequestState `json:"state"`
	URL       string                 `json:"url"`
	CreatedAt time.Time              `json:"created_at"`
	ClosedAt  time.Time              `json:"closed_at"`
	CreatedBy string                 `json:"created_by"`          // github username
	Assignees []string               `json:"assignees"`           // github usernames
	Reviewers []string               `json:"requested_reviewers"` // github usernames
}
