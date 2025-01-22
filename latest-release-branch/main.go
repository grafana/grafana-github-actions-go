package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghutil"
	"github.com/sethvargo/go-githubactions"
)

type Inputs struct {
	Owner   string
	Repo    string
	Pattern string
}

func GetInputs() Inputs {
	var (
		pattern   = githubactions.GetInput("pattern")
		ownerRepo = githubactions.GetInput("ownerRepo")
	)

	r := strings.Split(ownerRepo, "/")
	owner := r[0]
	repo := r[1]
	return Inputs{
		Pattern: pattern,
		Owner:   owner,
		Repo:    repo,
	}
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	var (
		ctx    = context.Background()
		token  = os.Getenv("GITHUB_TOKEN")
		client = github.NewTokenClient(ctx, token)
		inputs = GetInputs()
	)

	if token == "" {
		panic("token can not be empty")
	}
	major, minor, _ := ghutil.MajorMinorPatch(strings.TrimPrefix(strings.ReplaceAll(inputs.Pattern, "x", "0"), "v"))

	branches, err := ghutil.GetReleaseBranches(ctx, client.Repositories, inputs.Owner, inputs.Repo)
	if err != nil {
		log.Error("error getting release branches", "err", err)
		panic(err)
	}

	branch, err := ghutil.MostRecentBranch(major, minor, branches)
	if err != nil {
		log.Error("error getting release branches", "err", err)
		panic(err)
	}

	log.Info("found branch", "branch", branch)
	fmt.Fprint(os.Stdout, branch)
}
