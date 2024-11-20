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
	openPRs, err := findOpenPRs(ctx, client, owner, repo, prevBranch)
	if err != nil {
		githubactions.Fatalf("failed to find open PRs: %v", err)
	}
	// if no open PRs, exit/early return?

	// update base branch for each pull request to nextBranch
	// notify user of update
}

func findOpenPRs(ctx context.Context, client *github.Client, owner, repo, branch string) ([]PullRequestInfo, error) {
	// build pull request list options
	opts := &github.PullRequestListOptions{
		State: "open",
		Base:  branch,
	}

	// get all open pull requests
	open_prs, _, err := client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		// handle error with early return
		return nil, err
	}

	// create new empty slice and clean up data
	results := make([]PullRequestInfo, len(open_prs))
	for i, open_pr := range open_prs {
 		results[i] = PullRequestInfo{
			Number: open_pr.GetNumber(),
			AuthorName: open_pr.GetUser().GetLogin(),
		}
 	}

	return results, nil
}
