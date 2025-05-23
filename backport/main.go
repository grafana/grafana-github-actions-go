package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghutil"
	"github.com/sethvargo/go-githubactions"
)

type Inputs struct {
	Title  string
	Labels []*github.Label
}

func GetInputs() Inputs {
	var (
		labelsStr = githubactions.GetInput("labels_to_add")
	)

	labelStrings := strings.Split(labelsStr, ",")
	labels := make([]*github.Label, len(labelStrings))
	for i, v := range labelStrings {
		labels[i] = &github.Label{
			Name: github.String(v),
		}
	}

	return Inputs{
		Labels: labels,
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
	)

	if token == "" {
		panic("token can not be empty")
	}

	prInfo, err := GetBackportPrInfo(ctx, log, client, ghctx)
	if err != nil {
		log.Error("error getting PR info", "error", err)
		panic(err)
	}

	log = log.With("repo", fmt.Sprintf("%s/%s", prInfo.RepoOwner, prInfo.RepoName), "pull_request", prInfo.Pr.GetNumber())

	branches, err := ghutil.GetReleaseBranches(ctx, log, client.Repositories, prInfo.RepoOwner, prInfo.RepoName)
	if err != nil {
		log.Error("error getting branches", "error", err)
		panic(err)
	}

	targets, err := BackportTargetsFromPayload(branches, prInfo)
	if err != nil {
		if errors.Is(err, ErrorNotMerged) {
			log.Warn("pull request is not merged; nothing to do")
			return
		}

		log.Error("error getting backport targets", "error", err)
		panic(err)
	}

	for _, target := range targets {
		log := log.With("target", target)
		mergeBase, err := MergeBase(ctx, client.Repositories, prInfo.RepoOwner, prInfo.RepoName, target.Name, prInfo.Pr.GetBase().GetRef())
		if err != nil {
			log.Error("error finding merge-base", "error", err)
		}

		opts := BackportOpts{
			PullRequestNumber: prInfo.Pr.GetNumber(),
			SourceSHA:         prInfo.Pr.GetMergeCommitSHA(),
			SourceCommitDate:  prInfo.Pr.GetMergedAt().Time,
			SourceTitle:       prInfo.Pr.GetTitle(),
			SourceBody:        prInfo.Pr.GetBody(),
			Target:            target,
			Labels:            append(inputs.Labels, prInfo.Pr.Labels...),
			Owner:             prInfo.RepoOwner,
			Repository:        prInfo.RepoName,
			MergeBase:         mergeBase,
		}

		commandRunner := NewShellCommandRunner(log)
		prOut, err := Backport(ctx, log, client.PullRequests, client.Issues, client.Issues, commandRunner, opts)
		if err != nil {
			log.Error("backport failed", "error", err)
			continue
		}

		log.Info("backport successful", "url", prOut.GetURL())
	}
}
