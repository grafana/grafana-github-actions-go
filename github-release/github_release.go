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

func isPreRelease(_ string) *bool {
	b := true
	return &b
}

func getReleaseTitle(v string) string {
	return "Release notes for Grafana" + v
}

func createRelease(ctx context.Context, client *gh.Client, owner string, repo string, release *gh.RepositoryRelease) error {

	_, _, err := client.Repositories.CreateRelease(ctx, owner, repo, release)
	return err
}

var repoName = "grafana-github-actions-go"
var owner = "grafana"

// is there anything here that needs to be tested? if yes, break it up

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

	releaseTitle := getReleaseTitle(currentVersion)
	releaseBody := ""
	tagName := "v" + currentVersion //look for different way to combine string

	if err != nil { //create release if not found
		newRelease := &gh.RepositoryRelease{
			Name:       &releaseTitle,
			Body:       &releaseBody,
			TagName:    &tagName,
			Prerelease: isPreRelease(currentVersion),
		}

		if err := createRelease(ctx, client, owner, repoName, newRelease); err != nil {
			log.Fatal(err)
		}
		return
	}

	existingRelease.Name = &releaseTitle
	existingRelease.Body = &releaseBody
	existingRelease.TagName = &tagName

	// should i test if i get release by tag and it fails, and one gets created  - behavior i want to preserve
	//good chunk to test
	//combine below
	if _, _, err := client.Repositories.EditRelease(ctx, owner, repoName, *existingRelease.ID, existingRelease); err != nil {
		// log.Fatal(err)
	}

	if err != nil { //create release if not found
		newRelease := &gh.RepositoryRelease{
			Name:       &releaseTitle,
			Body:       &releaseBody,
			TagName:    &tagName,
			Prerelease: isPreRelease(currentVersion),
		}

		if err := createRelease(ctx, client, owner, repoName, newRelease); err != nil {
			log.Fatal(err)
		}
	}
}
