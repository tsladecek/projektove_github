package projektove

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ProjektoveStatus int

const (
	ProjektoveStatusNew ProjektoveStatus = iota
	ProjektoveStatusInProgress
	ProjektoveStatusResolved
	ProjektoveStatusFeedback
	ProjektoveStatusClosed
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

type ProjektoveUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProjektoveIssueTracker struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProjektoveIssue struct {
	ID          int                    `json:"id"`
	Project     ProjektoveProject      `json:"project"`
	Status      ProjektoveIssueStatus  `json:"status"`
	Author      ProjektoveUser         `json:"author"`
	AssignedTo  *ProjektoveUser        `json:"assignedTo"`
	Subject     string                 `json:"subject"`
	Description string                 `json:"description"`
	DueDate     time.Time              `json:"dueDate"`
	CreatedOn   time.Time              `json:"createdOn"`
	Tracker     ProjektoveIssueTracker `json:"tracker"`
}

func (pi ProjektoveIssue) Link() string {
	return fmt.Sprintf("https://app.projektove.cz/%s/tasks/%d", strings.ToLower(pi.Tracker.Name), pi.ID)
}

// regex to extract GitHub issue ID from the Projektove description
var ghIssueIDRegex = regexp.MustCompile(`GitHub Issue ID:\s*([0-9]+)`)

// determines whether this issue should be synced with github repository
// empty string means no
var ghRepoRegex = regexp.MustCompile(`GitHub Repository:\s*([\w-]+/[\w-]+)`)

func (p ProjektoveIssue) GithubRepository() string {
	matches := ghRepoRegex.FindStringSubmatch(p.Description)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// returns id of the github issue
// this is filled when the issue is synchronized with github
// 0 means that the issue has not been yet created
func (p ProjektoveIssue) GithubID() int {
	matches := ghIssueIDRegex.FindStringSubmatch(p.Description)
	if len(matches) > 1 {
		id, err := strconv.Atoi(matches[1])
		if err == nil {
			return id
		}
	}
	return 0
}

func (p *ProjektoveIssue) AnnotateWithGithubIssue(g GithubIssue) {
	p.Description = fmt.Sprintf("%s\n\nGitHub Issue URL: %s\n\nGitHub Issue ID: %d", p.Description, g.URL, g.ID)
}

type ProjektoveIssueUpdate struct {
	Subject      string    `json:"subject,omitempty"`
	Description  string    `json:"description,omitempty"`
	ProjectID    int       `json:"project_id,omitempty"`
	StartDate    time.Time `json:"start_date,omitzero"`
	DueDate      time.Time `json:"due_date,omitzero"`
	AuthorID     int       `json:"author_id,omitempty"`
	AssignedToID int       `json:"assigned_to_id,omitempty"`
	StatusID     int       `json:"status_id,omitempty"`
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

type GithubUser struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

type GithubIssue struct {
	ID        int              `json:"number"`
	Title     string           `json:"title"`
	Body      string           `json:"body"`
	URL       string           `json:"html_url"`
	Assignees []GithubUser     `json:"assignees"`
	State     GithubIssueState `json:"state"`
}

func GithubIssueCreateFromProjektove(pi ProjektoveIssue, users Users) (GithubIssueCreate, error) {
	var assignees []string
	if pi.AssignedTo != nil {
		assignee, found := users.GetGithubUser(*pi.AssignedTo)
		if !found {
			return GithubIssueCreate{}, fmt.Errorf("github user for projektove assignee %+v not found", pi.AssignedTo)
		}
		assignees = []string{assignee.Login}
	}

	return GithubIssueCreate{
		Title:     pi.Subject,
		Body:      fmt.Sprintf("%s\n\n%s", pi.Description, pi.Link()),
		Assignees: assignees,
	}, nil
}

type GithubIssueCreate struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Assignees []string `json:"assignees"`
}

type GithubIssueUpdate struct {
	Title     string       `json:"title"`
	Body      string       `json:"body"`
	Assignees []GithubUser `json:"assignees"`
}

type GithubPullRequest struct {
	ID        int                    `json:"number"`
	State     GithubPullRequestState `json:"state"`
	URL       string                 `json:"html_url"`
	CreatedAt time.Time              `json:"created_at"`
	ClosedAt  *time.Time             `json:"closed_at"`
	CreatedBy GithubUser             `json:"user"`
	ClosedBy  *GithubUser            `json:"closed_by"`
	Assignees []GithubUser           `json:"assignees"`
	Reviewers []GithubUser           `json:"requested_reviewers"`
}

type User struct {
	Github     GithubUser     `json:"github"`
	Projektove ProjektoveUser `json:"projektove"`
}

type Users []User

func (u Users) GetProjektoveUser(g GithubUser) (ProjektoveUser, bool) {
	for _, user := range u {
		if user.Github.ID == g.ID {
			return user.Projektove, true
		}
	}

	return ProjektoveUser{}, false
}

func (u Users) GetGithubUser(p ProjektoveUser) (GithubUser, bool) {
	for _, user := range u {
		if user.Projektove.ID == p.ID {
			return user.Github, true
		}
	}

	return GithubUser{}, false
}
