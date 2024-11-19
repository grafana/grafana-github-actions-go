package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
)

type PullRequestInfo struct {
	Number     int
	AuthorName string
}

func main() {
	// setup logging
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// setup github context
	ghctx, err := githubactions.Context()
	if err != nil {
		githubactions.Fatalf("failed to read github context: %v", err)
	}

	// get and validate inputs
	prevBranch := githubactions.GetInput("prevBranch")
	if prevBranch == "" {
		githubactions.Fatalf("prevBranch input is undefined")
	}

	nextBranch := githubactions.GetInput("nextBranch")
	if nextBranch == "" {
		githubactions.Fatalf("nextBranch input is undefined")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		githubactions.Fatalf("GITHUB_TOKEN is undefined")
	}

	// build github client
	ctx := context.Background()
	client := github.NewTokenClient(ctx, token)

	// get owner and repo from context
	owner, repo := ghctx.Repo()

	// get all open pull requests from prevBranch
	openPRs, err := findOpenPRs()
	if err != nil {
		githubactions.Fatalf("failed to find open PRs: %v", err)
	}
	// if no open PRs, exit

	// update base branch for each pull request to nextBranch
	// notify user of update
}

func findOpenPRs(ctx context.Context, client *github.Client, owner, repo, branch string) {
	opts := &github.PullRequestListOptions{
		State: "open",
		Base:  branch,
	}
}
