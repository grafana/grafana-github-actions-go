package main

import (
	"context"
	"fmt"
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
	var (
		ctx    = context.Background()
		token  = os.Getenv("GITHUB_TOKEN")
		inputs = GetInputs()
		client = github.NewTokenClient(ctx, token)
	)

	if token == "" {
		panic("token can not be empty")
	}

	if err := CreateNewReleaseBranch(ctx, client.Git, inputs.Owner, inputs.Repo, inputs.Source); err != nil {
		panic(fmt.Errorf("error creating new release branch: %s", err))
	}
}

type GitClient interface {
	GetRef(ctx context.Context, owner string, repo string, ref string) (*github.Reference, *github.Response, error)
	CreateRef(ctx context.Context, owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error)
}

func CreateNewReleaseBranch(ctx context.Context, client GitClient, owner, repo, source string) error {
	target, err := versions.BumpReleaseBranch(source)
	if err != nil {
		return err
	}

	ref, _, err := client.GetRef(ctx, owner, repo, source)
	if err != nil {
		return err
	}

	if _, _, err := client.CreateRef(ctx, owner, repo, &github.Reference{
		Ref:    github.String(target),
		Object: ref.Object,
	}); err != nil {
		return err
	}

	return nil
}
