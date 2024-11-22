package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
)

func main() {
	// retrieve and validate inputs
	prevBranch := githubactions.GetInput("prevBranch")
	if prevBranch == "" {
		githubactions.Fatalf("prevBranch input is undefined (value: '%s')", prevBranch)
	}
	nextBranch := githubactions.GetInput("nextBranch")
	if nextBranch == "" {
		githubactions.Fatalf("nextBranch input is undefined")
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		githubactions.Fatalf("GITHUB_TOKEN is undefined")
	}

	// setup Go context and github client and context
	ctx := context.Background()
	client := github.NewTokenClient(ctx, token)
	ghctx, err := githubactions.Context()
	if err != nil {
		githubactions.Fatalf("failed to read github context: %v", err)
	}

	// retrieve owner and repo from github context
	owner, repo := ghctx.Repo()

	openPRs, err := findOpenPRs(ctx, client, owner, repo, prevBranch)
	if err != nil {
		githubactions.Fatalf("failed to find open PRs: %v", err)
	}

	// if no open PRs, exit Action successfully with a notification
	if len(openPRs) == 0 {
		githubactions.Noticef("no open PRs found for %s", prevBranch)
		os.Exit(0)
	}

	ghClient := &GitHubClient{
		Client: client,
		Owner:  owner,
		Repo:   repo,
	}

	// iterate through all open PRs and update the base branch for each PR to `nextBranch`, then notify user of successful update
	for _, openPr := range openPRs {
		if err := UpdateBaseBranch(ctx, ghClient, openPr, nextBranch); err != nil {
			// log error and notify user to manually update their base branch
			githubactions.Errorf("failed to update base branch for PR %d: %v", openPr.Number, err)
			if err := NotifyUser(ctx, ghClient, openPr, prevBranch, nextBranch, false); err != nil {
				githubactions.Errorf("failed to notify user of update for PR %d: %v", openPr.Number, err)
			}
			continue
		}

		// notify user of successful update
		if err := NotifyUser(ctx, ghClient, openPr, prevBranch, nextBranch, true); err != nil {
			// log error and continue
			githubactions.Errorf("failed to notify user of update for PR %d: %v", openPr.Number, err)
			continue
		}

		// log success
		githubactions.Noticef("successfully updated PR %d to target %s", openPr.Number, nextBranch)
	}

}

func findOpenPRs(ctx context.Context, client *github.Client, owner, repo, branch string) ([]PullRequestInfo, error) {
	opts := &github.PullRequestListOptions{
		State: "open",
		Base:  branch,
	}

	openPRs, _, err := client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		// handle error with early return
		return nil, err
	}

	results := make([]PullRequestInfo, len(openPRs))
	for i, openPR := range openPRs {
		results[i] = PullRequestInfo{
			Number:     openPR.GetNumber(),
			AuthorName: openPR.GetUser().GetLogin(),
		}
	}

	return results, nil
}
