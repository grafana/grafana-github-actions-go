package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/google/go-github/v50/github"
)

var semverRegex = regexp.MustCompile(`^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>x|0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

type BackportOpts struct {
	// PullRequestNumber is the integer ID of the pull request being backported
	PullRequestNumber int

	// SourceSHA is the commit hash that will be cherry-picked into a pull request targeting Target
	SourceSHA string

	// SourceTitle is the title of the source PR which will be reused in the backport PRs
	SourceTitle string

	// SourceBody is the body of the source PR which will be reused in the backport PRs
	SourceBody string

	// Target is the base branch of the backport pull request
	Target string

	// Labels are labels that will be added to the backport pull request
	Labels []*github.Label

	// IssueNumber will set the "issue" field in the backport pull request
	IssueNumber *int

	Owner      string
	Repository string
}

func Run(ctx context.Context, command string, args ...string) (string, error) {
	var (
		stderr = bytes.NewBuffer(nil)
	)
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stderr = stderr

	err := cmd.Run()
	return stderr.String(), err
}

func Push(ctx context.Context, branch string) error {
	_, err := Run(ctx, "git", "push", "origin", branch)
	return err
}

func CreatePullRequest(ctx context.Context, client *github.Client, branch string, opts BackportOpts) error {
	title := fmt.Sprintf("[%s] %s", opts.Target, opts.SourceTitle)

	pr, _, err := client.PullRequests.Create(ctx, opts.Owner, opts.Repository, &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(branch),
		Base:  github.String(opts.Target),
		Issue: opts.IssueNumber,
	})

	if err != nil {
		return err
	}

	pr.Labels = opts.Labels
	if _, _, err := client.PullRequests.Edit(ctx, opts.Owner, opts.Repository, *pr.Number, pr); err != nil {
		return fmt.Errorf("error updating pull request with new labels: %w", err)
	}

	return nil
}

func BackportBranch(number int, target string) string {
	return fmt.Sprintf("backport-%d-to-%s", number, target)
}

func Backport(ctx context.Context, client *github.Client, opts BackportOpts) (string, error) {
	// 1. Run CLI commands to create a branch and cherry-pick
	//   * If the cherry-pick fails, write a comment in the source PR with instructions on manual backporting
	// 2. git push
	// 3. Open the pull request against the appropriate release branch
	branch := BackportBranch(opts.PullRequestNumber, opts.Target)
	if err := CreateCherryPickBranch(ctx, branch, opts); err != nil {
		return "", fmt.Errorf("error cherry-picking: %w", err)
	}

	if err := Push(ctx, branch); err != nil {
		return "", fmt.Errorf("error pushing: %w", err)
	}

	if err := CreatePullRequest(ctx, client, branch, opts); err != nil {
		return "", fmt.Errorf("error creating pull request: %w", err)
	}

	return "", nil
}
