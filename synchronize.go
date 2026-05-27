package projektove

import (
	"context"
	"fmt"
)

type ProjektoveGithub struct {
	Projektove ProjektoveIssue
	Github     *GithubIssue // nil if does not exist
}

// returns all projektove issues that should be synchronized with github and
// corresponding github issues (if available)
func MatchIssues(projektove []ProjektoveIssue, github map[string][]GithubIssue) []ProjektoveGithub {
	return nil
}

func Synchronize(ctx context.Context, projektove Projektove, github Github) error {
	// fetch all issues from projektove
	projectoveIssues, err := projektove.ListIssues(ctx)
	if err != nil {
		return fmt.Errorf("when listing issues from projektove: %w", err)
	}

	// fetch all issues from github
	githubIssues := make(map[string][]GithubIssue, 0)

	matched := MatchIssues(projectoveIssues, githubIssues)

	cbs := make([]func(context.Context) error, 0)

	for _, m := range matched {
		if m.Github == nil { // create
			cbs = append(cbs, func(c context.Context) error {
				body := GithubIssueCreate{Title: m.Projektove.Subject, Body: m.Projektove.Description, Assignees: m.Projektove.AssignedTo}
				if err := github.CreateIssue(c); err != nil {
					return fmt.Errorf("when creating issue on github: %w")
				}
			})
		}
	}

	return nil
}
