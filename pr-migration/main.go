package main

import (
	"log/slog"
	"os"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
)

type PullRequestInfo struct {
	Number     int
	AuthorName string
}

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

	// get and validate inputs
	prevBranch := githubactions.GetInput("prevBranch")
	if prevBranch == "" {
		panic("prevBranch is undefined")
	}

	nextBranch := githubactions.GetInput("nextBranch")
	if nextBranch == "" {
		panic("nextBranch is undefined")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		panic("GITHUB_TOKEN is undefined")
	}

	// build github client
	
	// get owner and repo from context
	// get all open pull requests from prevBranch
	// update base branch for each pull request to nextBranch
	// notify user of update
}
