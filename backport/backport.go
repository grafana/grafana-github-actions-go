package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

type BackportClient interface {
	Create(ctx context.Context, owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
	Edit(ctx context.Context, owner string, repo string, number int, pull *github.PullRequest) (*github.PullRequest, *github.Response, error)
}

type CommentClient interface {
	CreateComment(ctx context.Context, owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
}

func Push(ctx context.Context, runner CommandRunner, branch string) error {
	_, err := runner.Run(ctx, "git", "push", "origin", branch)
	return err
}

func CreatePullRequest(ctx context.Context, client BackportClient, branch string, opts BackportOpts) (*github.PullRequest, error) {
	title := fmt.Sprintf("[%s] %s", opts.Target, opts.SourceTitle)

	pr, _, err := client.Create(ctx, opts.Owner, opts.Repository, &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(branch),
		Base:  github.String(opts.Target),
		Issue: opts.IssueNumber,
		Draft: github.Bool(false),
	})

	if err != nil {
		return nil, err
	}

	pr.Labels = opts.Labels
	if _, _, err := client.Edit(ctx, opts.Owner, opts.Repository, *pr.Number, pr); err != nil {
		return nil, fmt.Errorf("error updating pull request with new labels: %w", err)
	}

	return pr, nil
}

func BackportBranch(number int, target string) string {
	return fmt.Sprintf("backport-%d-to-%s", number, target)
}

func backport(ctx context.Context, client BackportClient, runner CommandRunner, opts BackportOpts) (*github.PullRequest, error) {
	// 1. Run CLI commands to create a branch and cherry-pick
	//   * If the cherry-pick fails, write a comment in the source PR with instructions on manual backporting
	// 2. git push
	// 3. Open the pull request against the appropriate release branch
	branch := BackportBranch(opts.PullRequestNumber, opts.Target)
	if err := CreateCherryPickBranch(ctx, runner, branch, opts); err != nil {
		return nil, fmt.Errorf("error cherry-picking: %w", err)
	}

	if err := Push(ctx, runner, branch); err != nil {
		return nil, fmt.Errorf("error pushing: %w", err)
	}

	pr, err := CreatePullRequest(ctx, client, branch, opts)
	if err != nil {
		return nil, fmt.Errorf("error creating pull request: %w", err)
	}

	return pr, nil
}

func Backport(ctx context.Context, backportClient BackportClient, commentClient CommentClient, execClient CommandRunner, opts BackportOpts) (*github.PullRequest, error) {
	// Remove any `backport` related labels from the original PR, and mark this PR as a "backport"
	labels := []*github.Label{
		&github.Label{
			Name: github.String("backport"),
		},
	}
	for _, v := range opts.Labels {
		if strings.Contains(v.GetName(), "backport") {
			continue
		}

		labels = append(labels, v)
	}

	opts.Labels = labels
	pr, err := backport(ctx, backportClient, execClient, opts)
	if err != nil {
		if err := CommentFailure(ctx, commentClient, FailureOpts{
			BackportOpts: opts,
			Error:        err,
		}); err != nil {
			return nil, fmt.Errorf("error creating backport comment: %w", err)
		}
		return nil, err
	}

	return pr, nil
}
