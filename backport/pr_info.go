package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
)

type PrInfo struct {
	Pr *github.PullRequest

	// Contains the relevant labels for the backport.
	// When triggered by being labeled, it should be only the applied label.
	// When triggered by being closed, it should be all labels on the PR.
	Labels []string

	RepoOwner string
	RepoName  string
}

func GetBackportPrInfo(ctx context.Context, client *github.Client, ghctx *githubactions.GitHubContext) (PrInfo, error) {
	prLabel := os.Getenv("PR_LABEL")
	prNumber, _ := strconv.Atoi(os.Getenv("PR_NUMBER"))
	repoOwner := os.Getenv("REPO_OWNER")
	repoName := os.Getenv("REPO_NAME")

	// First, try to use the API to get the PR info
	if prNumber != 0 && repoOwner != "" && repoName != "" {
		return getFromApi(ctx, client, prLabel, repoOwner, repoName, prNumber)
	}

	// Try to use event data if present
	eventPath := ghctx.EventPath
	if eventPath != "" {
		return getFromEvent(ctx, client, ghctx)
	}

	return PrInfo{}, fmt.Errorf("no PR info found")
}

func getFromEvent(ctx context.Context, client *github.Client, ghctx *githubactions.GitHubContext) (PrInfo, error) {
	payload := &github.PullRequestTargetEvent{}

	if err := UnmarshalEventData(ghctx, &payload); err != nil {
		return PrInfo{}, err
	}

	if payload.PullRequest == nil {
		return PrInfo{}, fmt.Errorf("pull request is nil")
	}

	pr := payload.GetPullRequest()
	prInfo := PrInfo{
		Pr:        pr,
		RepoOwner: payload.GetRepo().GetOwner().GetLogin(),
		RepoName:  payload.GetRepo().GetName(),
	}

	switch action := payload.GetAction(); action {
	case "labeled":
		prInfo.Labels = []string{payload.GetLabel().GetName()}
	case "closed":
		prInfo.Labels = labelsToStrings(pr.Labels)
	default:
		return PrInfo{}, fmt.Errorf("unsupported action: %s", action)
	}

	return prInfo, nil
}

func getFromApi(ctx context.Context, client *github.Client, prLabel, repoOwner, repoName string, prNumber int) (PrInfo, error) {
	pr, _, err := client.PullRequests.Get(ctx, repoOwner, repoName, prNumber)
	if err != nil {
		return PrInfo{}, err
	}

	prInfo := PrInfo{
		Pr:        pr,
		RepoOwner: repoOwner,
		RepoName:  repoName,
	}

	if prLabel != "" {
		prInfo.Labels = []string{prLabel}
	} else {
		prInfo.Labels = labelsToStrings(pr.Labels)
	}

	return prInfo, nil
}

func labelsToStrings(labels []*github.Label) []string {
	strings := make([]string, len(labels))
	for i, label := range labels {
		strings[i] = label.GetName()
	}
	return strings
}
