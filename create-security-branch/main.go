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
	Version           string
	SecurityBranchNum string
	Owner             string
	Repo              string
}

func GetInputs() (Inputs, error) {
	var (
		version   = githubactions.GetInput("version")
		secNum    = githubactions.GetInput("security_branch_number")
		ownerRepo = githubactions.GetInput("repository")
	)

	r := strings.Split(ownerRepo, "/")
	if len(r) != 2 {
		return Inputs{}, fmt.Errorf("invalid repository format: %s, expected owner/repo", ownerRepo)
	}

	return Inputs{
		Version:           version,
		SecurityBranchNum: secNum,
		Owner:             r[0],
		Repo:              r[1],
	}, nil
}

func main() {
	log.Println("Getting token...")
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok || token == "" {
		log.Fatalf("token cannot be empty")
	}

	var (
		ctx    = context.Background()
		client = github.NewTokenClient(ctx, token)
	)

	inputs, err := GetInputs()
	if err != nil {
		log.Fatalf("error getting inputs: %s", err)
	}

	log.Println("Creating new security branch...")
	branch, err := CreateSecurityBranch(ctx, client.Git, inputs)
	if err != nil {
		log.Fatalf("error creating security branch: %s", err)
	}

	log.Println("created new security branch:", branch)

	// Write the new branch name to stdout so that it can be reused
	fmt.Fprint(os.Stdout, branch)
}

type GitClient interface {
	GetRef(ctx context.Context, owner string, repo string, ref string) (*github.Reference, *github.Response, error)
	CreateRef(ctx context.Context, owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error)
}

func CreateSecurityBranch(ctx context.Context, client GitClient, inputs Inputs) (string, error) {
	// Validate version format
	if !strings.Contains(inputs.Version, ".") {
		return "", fmt.Errorf("invalid version format: %s, expected x.y.z", inputs.Version)
	}

	// Validate security branch number format
	if len(inputs.SecurityBranchNum) != 2 {
		return "", fmt.Errorf("invalid security branch number format: %s, expected two digits (e.g., 01)", inputs.SecurityBranchNum)
	}

	securityBranch := fmt.Sprintf("%s+security-%s", inputs.Version, inputs.SecurityBranchNum)

	// Check if branch already exists
	if _, _, err := client.GetRef(ctx, inputs.Owner, inputs.Repo, "heads/"+securityBranch); err == nil {
		return "", fmt.Errorf("security branch %s already exists", securityBranch)
	}

	// Get the base branch (release branch)
	baseRef, _, err := client.GetRef(ctx, inputs.Owner, inputs.Repo, "heads/release-"+inputs.Version)
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
