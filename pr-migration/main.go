package main

import (
	"context"
	"fmt"
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
	// log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	// 	Level: slog.LevelDebug,
	// }))

	// setup github context
	ghctx, err := githubactions.Context()
	if err != nil {
		githubactions.Fatalf("failed to read github context: %v", err)
	}
	// log.Debug("github context loaded")

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
	// log.Debug("inputs retrieved", "prevBranch:", prevBranch, "nextBranch:", nextBranch)

	// build github client
	ctx := context.Background()
	client := github.NewTokenClient(ctx, token)

	// get owner and repo from context
	owner, repo := ghctx.Repo()
	// JEV: do we need to check of the owner and repo are empty?

	// get all open pull requests from prevBranch
	openPRs, err := findOpenPRs(ctx, client, owner, repo, prevBranch)
	if err != nil {
		githubactions.Fatalf("failed to find open PRs: %v", err)
	}

	// if no open PRs, exit successfully with a notification
	if len(openPRs) == 0 {
		githubactions.Noticef("no open PRs found for %s", prevBranch)
		os.Exit(0)
	}

	// update base branch for each pull request to nextBranch, and notify user of update
	for _, openPr := range openPRs {
		if err := updateBaseBranch(ctx, client, owner, repo, openPr.Number, nextBranch); err != nil {
			// log error and continue
			githubactions.Errorf("failed to update base branch for PR %d: %v", openPr.Number, err)
			continue
		}

		if err := notifyUserOfUpdate(ctx, client, owner, repo, openPr, prevBranch, nextBranch); err != nil {

		}
	}

}

func findOpenPRs(ctx context.Context, client *github.Client, owner, repo, branch string) ([]PullRequestInfo, error) {
	// build pull request list options
	opts := &github.PullRequestListOptions{
		State: "open",
		Base:  branch,
	}

	// get all open pull requests
	openPRs, _, err := client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		// handle error with early return
		return nil, err
	}

	// create new empty slice and clean up data
	results := make([]PullRequestInfo, len(openPRs))
	for i, openPR := range openPRs {
		results[i] = PullRequestInfo{
			Number:     openPR.GetNumber(),
			AuthorName: openPR.GetUser().GetLogin(),
		}
	}

	return results, nil
}

func updateBaseBranch(ctx context.Context, client *github.Client, owner, repo string, number int, branch string) error {
	_, _, err := client.PullRequests.Edit(ctx, owner, repo, number, &github.PullRequest{
		Base: &github.PullRequestBranch{
			Ref: github.String(branch),
		},
	})

	if err != nil {
		// JEV: should this error be handled in the function or returned?
		return err
	}

	return nil
}

func notifyUserOfUpdate(ctx context.Context, client *github.Client, owner, repo string, pr PullRequestInfo, prevBranch, nextBranch string) error {
	comment := fmt.Sprintf(
		"Hello @%s, the base branch for this PR has been updated from `%s` to `%s`. "+
			"I've updated this PR to target the new branch. "+
			"Please check for any merge conflicts that may need to be resolved.",
		pr.AuthorName, prevBranch, nextBranch)
}
