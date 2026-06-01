package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	projektove "github.com/tsladecek/projektove_github"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	projektoveURL := flag.String("projektove-url", "https://geneton.projektove.cz", "Projektove base URL (required)")
	projektoveToken := flag.String("projektove-token", "", "Projektove API token (required)")
	githubToken := flag.String("github-token", "", "GitHub API token (required)")
	githubURL := flag.String("github-url", "https://api.github.com", "GitHub API base URL")
	usernameMapRaw := flag.String("username-map", "", "Comma-separated github_user=projektove_user_id (e.g. tomas=736,john=123)")
	dryRun := flag.Bool("dry-run", false, "dry run exits before the requests are made")
	flag.Parse()

	if *projektoveURL == "" || *projektoveToken == "" || *githubToken == "" {
		flag.Usage()
		os.Exit(2)
	}

	usernameMap, err := parseUsernameMap(*usernameMapRaw)
	if err != nil {
		return fmt.Errorf("parse username map: %w", err)
	}

	client := projektove.NewClient()

	projAPI, err := projektove.NewProjektoveAPI(*projektoveURL, *projektoveToken, client)
	if err != nil {
		return fmt.Errorf("init projektove: %w", err)
	}

	ghAPI, err := projektove.NewGithubAPI(*githubURL, *githubToken, client)
	if err != nil {
		return fmt.Errorf("init github: %w", err)
	}

	return projektove.Synchronize(context.Background(), projAPI, ghAPI, usernameMap, *dryRun)
}

func parseUsernameMap(raw string) (projektove.Users, error) {
	if raw == "" {
		return nil, nil
	}
	m := make(projektove.Users, 0)
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return projektove.Users{}, fmt.Errorf("when unmarshalling users")
	}
	return m, nil
}
