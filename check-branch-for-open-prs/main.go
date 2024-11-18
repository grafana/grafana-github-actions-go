package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
)

type PullRequestInfo struct {}

func main() {
	// set up logger
	// get context
	// get inputs
	// build client
	// get open pull requests
	// format prs?
	// convert to JSON?
	// set output
}

// JEV: for extraction?
func setUpLogger() {}

func buildGitHubClient() {}

func getOpenPullRequests() {}

func formatPullRequests() {}
