package main

import (
	"context"
	"fmt"
	"log"
	"os"

	gh "github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

func readArgs(args []string) (string, string, error) {
	// Check if enough input parameters
	if len(args) < 3 {
		return "", "", fmt.Errorf("not enough input parameters")
	}

	token := args[1]
	currentVersion := args[2]
	return token, currentVersion, nil
}

func updateRelease(ctx context.Context, client *gh.Client, owner string, repo string, id int64, release *gh.RepositoryRelease) error {
	client.Repositories.EditRelease()

	return nil
}

var repoName = "grafana-github-actions-go"
var owner = "grafana"

func main() {
	// Get token and version
	token, currentVersion, err := readArgs(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	// Conntect to GH client
	ctx := context.Background()
	client := gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))

	existingRelease, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repoName, currentVersion)

	updateRelease(ctx, client, owner, repoName, *existingRelease.ID, existingRelease)
}
