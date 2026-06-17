package main

import (
	"context"
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
		client = github.NewTokenClient(ctx, token)
		inputs = GetInputs()

		// If specified, takes precedence over event data
		repoOwner   = os.Getenv("REPO_OWNER")
		repoName    = os.Getenv("REPO_NAME")
		prNumber, _ = strconv.Atoi(os.Getenv("PR_NUMBER"))
		prLabel     = os.Getenv("PR_LABEL")
		runID       = os.Getenv("GITHUB_RUN_ID")
	)

	if token == "" {
		panic("token can not be empty")
	}

	prInfo, err := GetBackportPrInfo(ctx, log, client, ghctx, repoOwner, repoName, prNumber, prLabel)
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
	targets, err := BackportTargets(ctx, log, client.Repositories, prInfo.RepoOwner, prInfo.RepoName, targetNames)
	if err != nil {
		panic(err)
	}

	failed := false
	for _, target := range targets {
		log := log.With("target", target)
		mergeBase, err := MergeBase(ctx, client.Repositories, prInfo.RepoOwner, prInfo.RepoName, target.Name, prInfo.Pr.GetBase().GetRef())
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
			GitToken:          token,
		}

		commandRunner := NewShellCommandRunner(log)
		gqlClient := ghgql.NewClient(token)
		prOut, err := Backport(ctx, log, client.PullRequests, client.Issues, client.Issues, client.Git, gqlClient, commandRunner, opts)
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
