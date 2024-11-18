package main

import ()

// JEV: reuse from check-branch-for-open-prs/main.go?
type PullRequestInfo struct {}

func main() {
	// set up logger - shared util?
	// get context - shared util?
	// get inputs - shared util?
	// verify GHA inputs - shared util?
	// parse pr JSON
	// build client
	// notify authors for each pr
}

// JEV: for extraction?
func setUpLogger() {}

func buildGitHubClient() {}

func parseJSON() {}

func notifyAuthors() {}
