package main

import (
	"context"
	"fmt"
	"log/slog"

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

func GetBackportPrInfo(ctx context.Context, log *slog.Logger, client *github.Client, ghctx *githubactions.GitHubContext, repoOwner string, repoName string, prNumber int, prLabel string) (PrInfo, error) {
	log.Debug("getting PR info", "event_path", ghctx.EventPath, "env_pr_label", prLabel, "env_pr_number", prNumber, "env_repo_owner", repoOwner, "env_repo_name", repoName)

	// Prefer using the env vars and API if they are set
	if prNumber != 0 && repoOwner != "" && repoName != "" {
		log.Debug("getting PR info from API")
		return getFromApi(ctx, client, prLabel, repoOwner, repoName, prNumber)
	}

	// Fall back to event data if present
	if ghctx.EventPath != "" {
		log.Debug("getting PR info from event")
		return getFromEvent(ghctx)
	}

	return PrInfo{}, fmt.Errorf("no PR info found")
}

func getFromEvent(ghctx *githubactions.GitHubContext) (PrInfo, error) {
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
