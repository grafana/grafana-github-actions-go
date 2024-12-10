package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/google/go-github/v50/github"
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
		ctx     = context.Background()
		token   = os.Getenv("GITHUB_TOKEN")
		client  = github.NewTokenClient(ctx, token)
		inputs  = GetInputs()
		payload = &github.PullRequestTargetEvent{}
	)

	if token == "" {
		panic("token can not be empty")
	}

	if err := UnmarshalEventData(ghctx, &payload); err != nil {
		log.Error("error reading github payload", "error", err)
		panic(err)
	}

	var (
		owner = payload.GetRepo().GetOwner().GetLogin()
		repo  = payload.GetRepo().GetName()
	)

	log = log.With("pull_request", payload.GetNumber())
	branches, err := GetReleaseBranches(ctx, client.Repositories, owner, repo)
	if err != nil {
		log.Error("error getting branches", "error", err)
		panic(err)
	}

	targets, err := BackportTargetsFromPayload(branches, payload)
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
		opts := BackportOpts{
			PullRequestNumber: payload.GetPullRequest().GetNumber(),
			SourceSHA:         payload.GetPullRequest().GetMergeCommitSHA(),
			SourceTitle:       payload.GetPullRequest().GetTitle(),
			SourceBody:        payload.GetPullRequest().GetBody(),
			Target:            target,
			Labels:            append(inputs.Labels, payload.GetPullRequest().Labels...),
			Owner:             owner,
			Repository:        repo,
		}
		pr, err := Backport(ctx, client.PullRequests, client.Issues, client.Issues, NewShellCommandRunner(log), opts)
		if err != nil {
			log.Error("backport failed", "error", err)
			continue
		}
		log.Info("backport successful", "url", pr.GetURL())
	}
}
