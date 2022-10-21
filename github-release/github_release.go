package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	gh "github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

type releaseCreator interface {
	CreateRelease(ctx context.Context, owner string, repo string, release *gh.RepositoryRelease) (*gh.RepositoryRelease, *gh.Response, error)
}

type releaseGetter interface {
	GetReleaseByTag(ctx context.Context, owner string, repo string, tag string) (*gh.RepositoryRelease, *gh.Response, error)
}

type releaseEditor interface {
	EditRelease(ctx context.Context, owner string, repo string, id int64, release *gh.RepositoryRelease) (*gh.RepositoryRelease, *gh.Response, error)
}

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

func createRelease(ctx context.Context, client releaseCreator, owner string, repo string, release *gh.RepositoryRelease) (*gh.RepositoryRelease, error) {
	r, _, err := client.CreateRelease(ctx, owner, repo, release)
	return r, err
}

type releaseClient interface {
	releaseCreator
	releaseGetter
	releaseEditor
}

// getOrCreateRelease will create a GitHub release if one was not found. The release argument is ignored if a release was found.
// If a release was not found, then one is created using the release argument.
// Returns true if a release was created, otherwise returns false.
func getOrCreateRelease(ctx context.Context, client releaseClient, owner string, repo string, version string, release *gh.RepositoryRelease) (*gh.RepositoryRelease, bool, error) {
	r, _, err := client.GetReleaseByTag(ctx, owner, repoName, version)
	if err != nil {
		r, err := createRelease(ctx, client, owner, repoName, release)
		if err != nil {
			return nil, false, err
		}
		return r, true, nil
	}
	return r, false, nil
}

// upsertRelease will update a GitHub release or create a release if the update failed.
func upsertRelease(ctx context.Context, client releaseClient, owner string, repo string, version string, release *gh.RepositoryRelease) (*gh.RepositoryRelease, error) {
	if release == nil {
		return nil, errors.New("release cannot be nil")
	}
	if release.ID == nil {
		return nil, errors.New("release ID cannot be nil")
	}
	r, _, err := client.EditRelease(ctx, owner, repoName, *release.ID, release)
	if err != nil {
		r, err = createRelease(ctx, client, owner, repoName, release)
		if err != nil {
			return nil, err
		}
	}
	return r, nil
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

	newRelease := &gh.RepositoryRelease{
		Name:       gh.String(getReleaseTitle(currentVersion)),
		Body:       gh.String(""),
		TagName:    gh.String("v" + currentVersion),
		Prerelease: isPreRelease(currentVersion),
	}

	existingRelease, created, err := getOrCreateRelease(ctx, client.Repositories, owner, repoName, currentVersion, newRelease)
	if err != nil {
		log.Fatal(err)
	}

	if created {
		return
	}

	existingRelease.Name = newRelease.Name
	existingRelease.Body = newRelease.Body
	existingRelease.TagName = newRelease.TagName

	if _, err := upsertRelease(ctx, client.Repositories, owner, repoName, currentVersion, existingRelease); err != nil {
		log.Fatal(err)
	}
}
