package main

import (
	"context"
	"os"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
)

type Inputs struct {
	From  string
	To    string
	Owner string
	Repo  string
}

func GetInputs() Inputs {
	var (
		from      = githubactions.GetInput("to")
		to        = githubactions.GetInput("from")
		ownerRepo = githubactions.GetInput("ownerRepo")
	)

	r := strings.Split(ownerRepo, "/")
	owner := r[0]
	repo := r[1]

	return Inputs{
		From:  from,
		To:    to,
		Owner: owner,
		Repo:  repo,
	}
}

func main() {
	// retrieve and validate inputs
	from := githubactions.GetInput("from")
	if from == "" {
		githubactions.Fatalf("to input is undefined (value: '%s')", from)
	}
	to := githubactions.GetInput("to")
	if to == "" {
		githubactions.Fatalf("to input is undefined")
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		githubactions.Fatalf("GITHUB_TOKEN is undefined")
	}

	// setup Go context and github client and context
	var (
		ctx    = context.Background()
		client = github.NewTokenClient(ctx, token)
		inputs = GetInputs()
		owner  = inputs.Owner
		repo   = inputs.Repo
	)

	openPRs, err := findOpenPRs(ctx, client, owner, repo, from)
	if err != nil {
		githubactions.Fatalf("failed to find open PRs: %v", err)
	}

	// if no open PRs, exit Action successfully with a notification
	if len(openPRs) == 0 {
		githubactions.Noticef("no open PRs found for %s", from)
		os.Exit(0)
	}

	ghClient := &GitHubClient{
		Client: client,
		Owner:  owner,
		Repo:   repo,
	}

	// iterate through all open PRs and update the base branch for each PR to `to`, then notify user of successful update
	for _, openPr := range openPRs {
		if err := UpdateBaseBranch(ctx, ghClient, openPr, to); err != nil {
			// log error and notify user to manually update their base branch
			githubactions.Errorf("failed to update base branch for PR %d: %v", openPr.Number, err)
			if err := NotifyUser(ctx, ghClient, openPr, from, to, false); err != nil {
				githubactions.Errorf("failed to notify user of update for PR %d: %v", openPr.Number, err)
			}
			continue
		}

		// notify user of successful update
		if err := NotifyUser(ctx, ghClient, openPr, from, to, true); err != nil {
			githubactions.Errorf("failed to notify user of update for PR %d: %v", openPr.Number, err)
			continue
		}

		githubactions.Noticef("successfully updated PR %d to target %s", openPr.Number, to)
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
