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
	projektoveURL := flag.String("projektoveURL", "", "Projektove base URL (required)")
	projektoveToken := flag.String("projektoveToken", "", "Projektove API token (required)")
	githubToken := flag.String("githubToken", "", "GitHub API token (required)")
	githubURL := flag.String("githubURL", "https://api.github.com", "GitHub API base URL")
	usersRaw := flag.String("users", "", `list of users in JSON format: [{"projektove": {"id": ..., "name": ...}, "github": {"id": ..., "login": ...}}]`)
	dryRun := flag.Bool("dryRun", false, "dry run exits before the requests are made")
	flag.Parse()

	if *projektoveURL == "" || *projektoveToken == "" || *githubToken == "" {
		flag.Usage()
		os.Exit(2)
	}

	users, err := parseUsers(*usersRaw)
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

	return projektove.Synchronize(context.Background(), projAPI, ghAPI, users, *dryRun)
}

func parseUsers(raw string) (projektove.Users, error) {
	if raw == "" {
		return projektove.Users{}, fmt.Errorf("no users supplied")
	}
	m := make(projektove.Users, 0)
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return projektove.Users{}, fmt.Errorf("when unmarshalling users")
	}
	return m, nil
}
