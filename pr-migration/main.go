package main

import (
	"log/slog"
	"os"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
)

type PullRequestInfo struct{}

func main() {
	// setup logging
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// setup github context
	ghctx, err := githubactions.Context()
	if err != nil {
		log.Error("error reading github context", "error", err)
		panic(err)
	}

	// get inputs
	// validate inputs
	// build github client
	// get owner and repo from context
	// get all open pull requests from prevBranch
	// update base branch for each pull request to nextBranch
	// notify user of update
}
