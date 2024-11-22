package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
)

type PullRequestInfo struct {
	Number     int
	AuthorName string
}

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

	owner, repo := ghctx.Repo()
	// JEV: do we need to check of the owner and repo are empty?

	openPRs, err := findOpenPRs(ctx, client, owner, repo, prevBranch)
	if err != nil {
		githubactions.Fatalf("failed to find open PRs: %v", err)
	}

	// if no open PRs, exit Action successfully with a notification
	if len(openPRs) == 0 {
		githubactions.Noticef("no open PRs found for %s", prevBranch)
		os.Exit(0)
	}

	// iterate through all open PRs and update the base branch for each PR to `nextBranch`, then notify user of successful update
	for _, openPr := range openPRs {
		if err := updateBaseBranch(ctx, client, owner, repo, openPr.Number, nextBranch); err != nil {
			// log error and notify user to manually update their base branch
			githubactions.Errorf("failed to update base branch for PR %d: %v", openPr.Number, err)
			if err := notifyUserOfUpdate(ctx, client, owner, repo, openPr, prevBranch, nextBranch, false); err != nil {
				githubactions.Errorf("failed to notify user of update for PR %d: %v", openPr.Number, err)
			}
			continue
		}

		// notify user of successful update
		if err := notifyUserOfUpdate(ctx, client, owner, repo, openPr, prevBranch, nextBranch, true); err != nil {
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

func updateBaseBranch(ctx context.Context, client *github.Client, owner, repo string, number int, branch string) error {
	_, _, err := client.PullRequests.Edit(ctx, owner, repo, number, &github.PullRequest{
		Base: &github.PullRequestBranch{
			Ref: github.String(branch),
		},
	})

	return err
}

func notifyUserOfUpdate(ctx context.Context, client *github.Client, owner, repo string, pr PullRequestInfo, prevBranch, nextBranch string, succeeded bool) error {
	successComment := fmt.Sprintf(
		"Hello @%s, we've noticed that the original base branch `%s` for this PR is no longer a release candidate. "+
			"We've attempted to automatically updated your PR's base branch to the current release target: `%s`. "+
			"Please review and resolve any potential merge conflicts. "+
			"If this PR is not merged it will NOT be included in the next release. Thanks!",
		pr.AuthorName, prevBranch, nextBranch)

	failComment := fmt.Sprintf(
		"Hello @%s, we've noticed that the original base branch `%s` for this PR is no longer a release candidate. "+
			"We attempted to automatically update your PR's base branch to the current release target `%s`, but encountered an error. "+
			"Please manually update your PR's base branch to `%s` and resolve any merge conflicts. "+
			"If this PR is not rebased and merged it will NOT be included in the next release.",
		pr.AuthorName, prevBranch, nextBranch, nextBranch)

	comment := successComment
	if !succeeded {
		comment = failComment
	}

	_, _, err := client.Issues.CreateComment(ctx, owner, repo, pr.Number, &github.IssueComment{
		Body: &comment,
	})

	return err
}
