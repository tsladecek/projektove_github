package projektove

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type ProjektoveGithub struct {
	Projektove ProjektoveIssue
	Github     *GithubIssue // nil if does not exist
}

const (
	IssueClosedInProjektove = "\n\nIssue has been closed in Projektove"
)

// returns all projektove issues that should be synchronized with github and
// corresponding github issues (if available)
func MatchIssues(projektove []ProjektoveIssue, github map[string][]GithubIssue) []ProjektoveGithub {
	var matched []ProjektoveGithub
	for _, p := range projektove {
		repo := p.GithubRepository()

		if repo == "" {
			continue
		}

		ghID := p.GithubID()
		var ghIssue *GithubIssue

		// search for ghIssue in github[repo]
		if issues, ok := github[repo]; ok {
			for _, i := range issues {
				if i.ID == ghID {
					ghIssue = &i
					break
				}
			}
		}

		matched = append(matched, ProjektoveGithub{Projektove: p, Github: ghIssue})
	}
	return matched
}

func printPlan(plan map[string]func() error) {
	for p, _ := range plan {
		slog.Info(p)
	}
}

func Synchronize(ctx context.Context, projektove Projektove, github Github, users Users, dryRun, withConfirmation bool) error {
	// fetch all issues from projektove
	projektoveIssues, err := projektove.ListIssues(ctx)
	if err != nil {
		return fmt.Errorf("when listing issues from projektove: %w", err)
	}

	// fetch all issues from github
	repos := make(map[string]bool)
	for _, p := range projektoveIssues {
		repo := p.GithubRepository()
		if repo != "" {
			repos[repo] = true
		}
	}

	githubIssues := make(map[string][]GithubIssue)
	for repo := range repos {
		issues, err := github.ListIssues(ctx, repo)
		if err != nil {
			return fmt.Errorf("when listing issues from github for %s: %w", repo, err)
		}
		githubIssues[repo] = issues
	}

	matched := MatchIssues(projektoveIssues, githubIssues)

	plan := make(map[string]func() error)

	for _, m := range matched {
		repo := m.Projektove.GithubRepository()
		key := fmt.Sprintf("ProjektoveIssue: %d (%s)\tGithub Repository: %s", m.Projektove.ID, m.Projektove.Subject, repo)

		switch {
		case m.Github == nil: // create
			plan["create: "+key] = func() error {
				body, err := GithubIssueCreateFromProjektove(m.Projektove, users)
				if err != nil {
					return fmt.Errorf("when constructing github create object from projektove issue: %w", err)
				}
				gi, err := github.CreateIssue(ctx, repo, body)
				if err != nil {
					return fmt.Errorf("when creating issue on github: %w", err)
				}

				m.Projektove.AnnotateWithGithubIssue(gi)
				updateObj := ProjektoveIssueUpdate{Description: m.Projektove.Description}
				if err := projektove.UpdateIssue(ctx, m.Projektove.ID, updateObj); err != nil {
					return fmt.Errorf("when updating projektove issue description: %w", err)
				}

				return nil
			}

		case m.Github.State == GithubIssueStateClosed && !m.Projektove.Status.IsClosed: // close
			plan["close projektove: "+key] = func() error {
				if err := projektove.UpdateIssue(ctx, m.Projektove.ID, ProjektoveIssueUpdate{
					StatusID: int(ProjektoveStatusClosed),
				}); err != nil {
					return fmt.Errorf("when closing issue #%d in projektove: %w", m.Projektove.ID, err)
				}
				return nil
			}

		case m.Projektove.Status.IsClosed && m.Github.State != GithubIssueStateClosed && !strings.Contains(m.Github.Body, IssueClosedInProjektove):
			plan["close github: "+key] = func() error {
				// we do not want to close the issue, since work might have begun
				if err := github.UpdateIssue(ctx, repo, m.Github.ID, GithubIssueUpdate{Body: m.Github.Body + IssueClosedInProjektove}); err != nil {
					return fmt.Errorf("when closing github issue %d: %w", m.Github.ID, err)
				}

				return nil
			}
		}
	}

	if len(plan) == 0 {
		slog.Info("Nothing to sync. If this is not expected, ensure that the repository key is correctly typed.", "Repository Key", PrefixGithubRepository)
		return nil
	}

	if dryRun {
		printPlan(plan)
		return nil
	}

	if withConfirmation {
		printPlan(plan)

		fmt.Print("Would you like to proceed? [y,N]: ")

		reader := bufio.NewReader(os.Stdin)
		decision, _ := reader.ReadString('\n')

		decision = strings.TrimSpace(decision)

		if strings.ToLower(decision) != "y" {
			slog.Info("Aborting")
			return nil
		}
	}

	// Phase 2: Concurrent Execution
	var wg sync.WaitGroup
	var mu sync.Mutex

	for key, action := range plan {
		wg.Add(1)
		go func(k string, a func() error) {
			defer wg.Done()

			mu.Lock()
			slog.Info("[START]", "key", k)
			mu.Unlock()

			if err := a(); err != nil {
				mu.Lock()
				slog.Error("[FAIL]", "key", k, "err", err)
				mu.Unlock()
			} else {
				mu.Lock()
				slog.Info("[SUCCESS]", "key", k)
				mu.Unlock()
			}
		}(key, action)
	}
	wg.Wait()

	return nil
}
