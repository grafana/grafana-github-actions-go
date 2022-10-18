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

	if err == nil {
		_, _, err := client.Repositories.EditRelease(ctx, owner, repoName, *existingRelease.ID, existingRelease)
	} else {
		client.Repositories.CreateRelease(ctx, owner, repoName, existingRelease)
		var newRelease gh.RepositoryRelease

		newRelease := gh.RepositoryRelease{
			Name: "Release notes for Grafana" + currentVersion
			Body: ""
			TagName: "v" + currentVersion
			Prerelease: ,
		}
		// todo
		// this.title = `Release notes for Grafana ${this.version}`
		// calculate the title afterwards


		// name: title,
		// body: content,
		// tag_name: tag,
		// prerelease: isPreRelease(tag),
	}


}
