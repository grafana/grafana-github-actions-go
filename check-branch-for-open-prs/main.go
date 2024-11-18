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

// JEV: abstract this out?
func getOpenPullRequests() {}

// JEV: abstract this out?
func formatPullRequests() {}
