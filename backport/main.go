package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
)

type Inputs struct {
	Title                  string
	Labels                 []*github.Label
	RemoveDefaultReviewers bool
}

func GetInputs() Inputs {
	var (
		labelsStr                 = githubactions.GetInput("labelsToAdd")
		removeDefaultReviewersStr = githubactions.GetInput("removeDefaultReviewers")
		title                     = githubactions.GetInput("title")
	)

	labelStrings := strings.Split(labelsStr, ",")
	labels := make([]*github.Label, len(labelStrings))
	for i, v := range labelStrings {
		labels[i] = &github.Label{
			Name: github.String(v),
		}
	}

	removeDefaultReviewers, _ := strconv.ParseBool(removeDefaultReviewersStr)

	return Inputs{
		Title:                  title,
		Labels:                 labels,
		RemoveDefaultReviewers: removeDefaultReviewers,
	}
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ghctx, err := githubactions.Context()
	if err != nil {
		log.Error("error reading github context", "error", err)
		panic(err)
	}

	var (
		ctx     = context.Background()
		token   = githubactions.GetInput("token")
		client  = github.NewTokenClient(ctx, token)
		inputs  = GetInputs()
		payload = &github.PullRequestEvent{}
	)

	if err := UnmarshalEventData(ghctx, &payload); err != nil {
		log.Error("error reading github payload", "error", err)
		panic(err)
	}

	var (
		owner = payload.GetRepo().GetOwner().GetLogin()
		repo  = payload.GetRepo().GetName()
	)

	log = log.With("pull_request", payload.GetNumber())

	branches, err := GetReleaseBranches(ctx, client, owner, repo)
	if err != nil {
		log.Error("error getting branches", "error", err)
		panic(err)
	}

	targets, err := BackportTargets(branches, payload.GetPullRequest().Labels)
	if err != nil {
		log.Error("error getting backport target", "error", err)
		panic(err)
	}

	failures := []FailureOpts{}

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
		u, err := Backport(ctx, client, opts)
		if err != nil {
			log.Error("backport failed")
			failures = append(failures, FailureOpts{
				BackportOpts: opts,
				Error:        err,
			})
		}
		log.Info("backport successful", "url", u)
	}

	for _, v := range failures {
		if err := CommentFailure(ctx, client, v); err != nil {
			log.Error("error commenting backport instructions", "error", err)
		}
	}
}
