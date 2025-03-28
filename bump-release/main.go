package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/versions"
	"github.com/sethvargo/go-githubactions"
)

type Inputs struct {
	Source string
	Owner  string
	Repo   string
}

func GetInputs() Inputs {
	var (
		source    = githubactions.GetInput("source")
		ownerRepo = githubactions.GetInput("ownerRepo")
	)

	r := strings.Split(ownerRepo, "/")
	owner := r[0]
	repo := r[1]
	return Inputs{
		Source: source,
		Owner:  owner,
		Repo:   repo,
	}
}

func main() {
	log.Println("Getting token...")
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok || token == "" {
		log.Fatalf("token can not be empty")
	}

	var (
		ctx    = context.Background()
		inputs = GetInputs()
		client = github.NewTokenClient(ctx, token)
	)

	log.Println("Creating new release branch...")
	branch, err := CreateNewReleaseBranch(ctx, client.Git, inputs.Owner, inputs.Repo, inputs.Source)
	if err != nil {
		log.Fatalf("error creating new release branch: %s", err)
	}

	log.Println("created new branch:", branch)

	// Write the new branch name to stdout so that it can be reused
	fmt.Fprint(os.Stdout, branch)
}

type GitClient interface {
	GetRef(ctx context.Context, owner string, repo string, ref string) (*github.Reference, *github.Response, error)
	CreateRef(ctx context.Context, owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error)
}

func CreateNewReleaseBranch(ctx context.Context, client GitClient, owner, repo, source string) (string, error) {
	target, err := versions.BumpReleaseBranch(source)
	if err != nil {
		return "", fmt.Errorf("error bumping release branch: %w", err)
	}

	ref, _, err := client.GetRef(ctx, owner, repo, "heads/"+source)
	if err != nil {
		return "", fmt.Errorf("error getting ref: %w", err)
	}

	if _, _, err := client.GetRef(ctx, owner, repo, "heads/"+target); err == nil {
		return "", errors.New("requested branch already exists")
	}

	if _, _, err := client.CreateRef(ctx, owner, repo, &github.Reference{
		Ref:    github.String("heads/" + target),
		Object: ref.Object,
	}); err != nil {
		return "", fmt.Errorf("error creating ref: %w", err)
	}

	return target, nil
}
