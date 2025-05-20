package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
)

type Inputs struct {
	Source            string
	SecurityBranchNum string
	Owner             string
	Repo              string
}

func GetInputs() (Inputs, error) {
	source := githubactions.GetInput("release_branch")
	secNum := githubactions.GetInput("security_branch_number")
	ownerRepo := githubactions.GetInput("repository")

	if source == "" || secNum == "" || ownerRepo == "" {
		return Inputs{}, fmt.Errorf("all inputs (release_branch, security_branch_number, repository) are required")
	}

	r := strings.Split(ownerRepo, "/")
	if len(r) != 2 {
		return Inputs{}, fmt.Errorf("invalid repository format: %s, expected owner/repo", ownerRepo)
	}

	return Inputs{
		Source:            source,
		SecurityBranchNum: secNum,
		Owner:             r[0],
		Repo:              r[1],
	}, nil
}

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatalf("GITHUB_TOKEN is required")
	}

	ctx := context.Background()
	client := github.NewTokenClient(ctx, token)

	inputs, err := GetInputs()
	if err != nil {
		log.Fatalf("error getting inputs: %s", err)
	}

	branch, err := CreateSecurityBranch(ctx, client.Git, inputs)
	if err != nil {
		log.Fatalf("error creating security branch: %s", err)
	}

	fmt.Fprint(os.Stdout, branch)
}

type GitClient interface {
	GetRef(ctx context.Context, owner string, repo string, ref string) (*github.Reference, *github.Response, error)
	CreateRef(ctx context.Context, owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error)
}

func CreateSecurityBranch(ctx context.Context, client GitClient, inputs Inputs) (string, error) {
	securityBranch := fmt.Sprintf("%s+security-%s", inputs.Source, inputs.SecurityBranchNum)

	// Check if branch already exists
	if _, _, err := client.GetRef(ctx, inputs.Owner, inputs.Repo, "heads/"+securityBranch); err == nil {
		return "", fmt.Errorf("security branch %s already exists", securityBranch)
	}

	// Get the base branch
	baseRef, _, err := client.GetRef(ctx, inputs.Owner, inputs.Repo, "heads/"+inputs.Source)
	if err != nil {
		return "", fmt.Errorf("error getting base ref: %w", err)
	}

	// Create the security branch
	if _, _, err := client.CreateRef(ctx, inputs.Owner, inputs.Repo, &github.Reference{
		Ref:    github.String("heads/" + securityBranch),
		Object: baseRef.Object,
	}); err != nil {
		return "", fmt.Errorf("error creating security branch: %w", err)
	}

	return securityBranch, nil
}
