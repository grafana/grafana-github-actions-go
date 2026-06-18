package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghgql"
	"github.com/sethvargo/go-githubactions"
)

type Inputs struct {
	TitleTemplate string
	Labels        []*github.Label
}

func GetInputs() Inputs {
	var (
		labelsStr     = githubactions.GetInput("labels_to_add")
		titleTemplate = os.Getenv("PR_TITLE_TEMPLATE")
	)

	if titleTemplate == "" {
		titleTemplate = githubactions.GetInput("pr_title_template")
	}

	labelStrings := strings.Split(labelsStr, ",")
	labels := make([]*github.Label, len(labelStrings))
	for i, v := range labelStrings {
		labels[i] = &github.Label{
			Name: github.String(v),
		}
	}

	return Inputs{
		TitleTemplate: titleTemplate,
		Labels:        labels,
	}
}

// resolveTokens decides which token authenticates the push operations (cherry-pick
// fetch and the signed-commit publish) and which authenticates the PR operations
// (opening the PR, editing labels, commenting on failure).
//
// The granular tokens let a caller use two separate GitHub Apps with different
// permissions. They must be set together or not at all; when set they take
// precedence over token. When they are absent, token is used for both.
func resolveTokens(token, gitPushToken, prOpenToken string) (pushToken, prToken string, err error) {
	if (gitPushToken == "") != (prOpenToken == "") {
		return "", "", errors.New("GITHUB_GIT_PUSH_TOKEN and GITHUB_PR_OPEN_TOKEN must be set together")
	}

	if gitPushToken != "" {
		return gitPushToken, prOpenToken, nil
	}

	if token == "" {
		return "", "", errors.New("token can not be empty")
	}

	return token, token, nil
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ghctx, err := githubactions.Context()
	if err != nil {
		log.Error("error reading github context", "error", err)
		panic(err)
	}

	var (
		ctx    = context.Background()
		token  = os.Getenv("GITHUB_TOKEN")
		inputs = GetInputs()

		// If specified, takes precedence over event data
		repoOwner   = os.Getenv("REPO_OWNER")
		repoName    = os.Getenv("REPO_NAME")
		prNumber, _ = strconv.Atoi(os.Getenv("PR_NUMBER"))
		prLabel     = os.Getenv("PR_LABEL")
		runID       = os.Getenv("GITHUB_RUN_ID")

		// Granular tokens if you have more than one app that does backporting: one
		// pushes the commits (contents), the other opens the PR (pull requests).
		gitPushToken = os.Getenv("GITHUB_GIT_PUSH_TOKEN")
		prOpenToken  = os.Getenv("GITHUB_PR_OPEN_TOKEN")
	)

	pushToken, prToken, err := resolveTokens(token, gitPushToken, prOpenToken)
	if err != nil {
		panic(err)
	}
	if gitPushToken != "" {
		log.Info("Using GITHUB_GIT_PUSH_TOKEN and GITHUB_PR_OPEN_TOKEN instead of GITHUB_TOKEN")
	}

	// pushClient authenticates repository reads and commit pushes; prClient
	// authenticates reading the source PR, opening the backport PR and commenting.
	// When no granular tokens are set, both wrap the same token.
	pushClient := github.NewTokenClient(ctx, pushToken)
	prClient := github.NewTokenClient(ctx, prToken)

	prInfo, err := GetBackportPrInfo(ctx, log, prClient, ghctx, repoOwner, repoName, prNumber, prLabel)
	if err != nil {
		log.Error("error getting PR info", "error", err)
		panic(err)
	}

	if !prInfo.Pr.GetMerged() {
		panic("PR hasn't been merged yet")
	}

	log = log.With("repo", fmt.Sprintf("%s/%s", prInfo.RepoOwner, prInfo.RepoName), "pull_request", prInfo.Pr.GetNumber())

	targetNames := BackportTargetsFromLabels(prInfo.Labels, "backport ")
	if len(targetNames) == 0 {
		log.Info("no backport labels found, nothing to do")
		return
	}
	targets, err := BackportTargets(ctx, log, pushClient.Repositories, prInfo.RepoOwner, prInfo.RepoName, targetNames)
	if err != nil {
		panic(err)
	}

	failed := false
	for _, target := range targets {
		log := log.With("target", target)
		mergeBase, err := MergeBase(ctx, pushClient.Repositories, prInfo.RepoOwner, prInfo.RepoName, target.Name, prInfo.Pr.GetBase().GetRef())
		if err != nil {
			log.Error("error finding merge-base", "error", err)
			failed = true
			continue
		}

		opts := BackportOpts{
			PullRequestNumber: prInfo.Pr.GetNumber(),
			SourceSHA:         prInfo.Pr.GetMergeCommitSHA(),
			SourceCommitDate:  prInfo.Pr.GetMergedAt().Time,
			SourceTitle:       prInfo.Pr.GetTitle(),
			TitleTemplate:     inputs.TitleTemplate,
			SourceBody:        prInfo.Pr.GetBody(),
			Target:            target,
			Labels:            append(inputs.Labels, prInfo.Pr.Labels...),
			Owner:             prInfo.RepoOwner,
			Repository:        prInfo.RepoName,
			MergeBase:         mergeBase,
			RunID:             runID,
			GitToken:          pushToken,
		}

		commandRunner := NewShellCommandRunner(log)
		gqlClient := ghgql.NewClient(pushToken)
		prOut, err := Backport(ctx, log, prClient.PullRequests, prClient.Issues, prClient.Issues, pushClient.Git, gqlClient, commandRunner, opts)
		if err != nil {
			log.Error("backport failed", "error", err)
			failed = true
			continue
		}

		log.Info("backport successful", "url", prOut.GetURL())
	}

	if failed {
		os.Exit(1)
	}
}
