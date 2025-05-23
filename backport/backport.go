package main

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghutil"
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

	// SourceCommitDate is the datetime when the cherry-picked commit was committed
	SourceCommitDate time.Time

	// Target is the base branch of the backport pull request
	Target ghutil.Branch

	// Labels are labels that will be added to the backport pull request
	Labels []*github.Label

	// IssueNumber will set the "issue" field in the backport pull request
	IssueNumber *int

	Owner      string
	Repository string

	// MergeBase is used to determine how deep in the history to fetch for the cherry-pick to work
	MergeBase *github.Commit
}

type BackportClient interface {
	Create(ctx context.Context, owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
}

type IssueClient interface {
	Edit(ctx context.Context, owner string, repo string, number int, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
}

type CommentClient interface {
	CreateComment(ctx context.Context, owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
}

func Push(ctx context.Context, runner CommandRunner, branch string) error {
	// Retry pushing every 5 seconds for a full minute
	return retry(func() error {
		_, err := runner.Run(ctx, "git", "push", "origin", branch)
		return err
	}, 12, time.Second*5)
}

func CreatePullRequest(ctx context.Context, client BackportClient, issueClient IssueClient, branch string, opts BackportOpts) (*github.PullRequest, error) {
	title := fmt.Sprintf("[%s] %s", opts.Target.Name, opts.SourceTitle)

	body := fmt.Sprintf("Backport %s from #%d\n\n---\n\n%s", opts.SourceSHA, opts.PullRequestNumber, opts.SourceBody)

	pr, _, err := client.Create(ctx, opts.Owner, opts.Repository, &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(branch),
		Base:  github.String(opts.Target.Name),
		Issue: opts.IssueNumber,
		Body:  github.String(body),
		Draft: github.Bool(false),
	})

	if err != nil {
		return nil, err
	}

	labels := []string{}
	for _, v := range opts.Labels {
		if strings.TrimSpace(v.GetName()) == "" {
			continue
		}

		labels = append(labels, v.GetName())
	}

	issue, _, err := issueClient.Edit(ctx, opts.Owner, opts.Repository, pr.GetNumber(), &github.IssueRequest{
		Labels: &labels,
	})

	if err != nil {
		return nil, fmt.Errorf("error updating pull request with new labels: %w", err)
	}

	// Instead of wasting time querying for the PR again to make sure it updated, just
	// use the returned issue, which is basically the same thing
	pr.Labels = issue.Labels
	return pr, nil
}

func BackportBranch(number int, target string) string {
	return fmt.Sprintf("backport-%d-to-%s", number, target)
}

func retry(fn func() error, count int, d time.Duration) error {
	c := time.NewTicker(d)
	var err error
	for i := 0; i < count; i++ {
		<-c.C
		err = fn()
		if err == nil {
			return nil
		}
	}

	return err
}

func backport(ctx context.Context, log *slog.Logger, client BackportClient, issueClient IssueClient, runner CommandRunner, opts BackportOpts) (*github.PullRequest, error) {
	// 1. Run CLI commands to create a branch and cherry-pick
	//   * If the cherry-pick fails, write a comment in the source PR with instructions on manual backporting
	// 2. git push
	// 3. Open the pull request against the appropriate release branch
	branch := BackportBranch(opts.PullRequestNumber, opts.Target.Name)
	if err := CreateCherryPickBranch(ctx, runner, branch, opts); err != nil {
		return nil, fmt.Errorf("error cherry-picking: %w", err)
	}

	if err := Push(ctx, runner, branch); err != nil {
		return nil, fmt.Errorf("error pushing: %w", err)
	}

	var (
		pr *github.PullRequest
	)

	// This will attempt to open the pull request once every second 10 times until it succeeds
	err := retry(func() error {
		log.Info("Attempting to create pull request", "head", branch)
		p, err := CreatePullRequest(ctx, client, issueClient, branch, opts)
		if err != nil {
			return fmt.Errorf("error creating pull request: %w", err)
		}

		pr = p
		return nil
	}, 10, time.Second)

	if err != nil {
		return nil, err
	}
	return pr, nil
}

func Backport(ctx context.Context, log *slog.Logger, backportClient BackportClient, commentClient CommentClient, issueClient IssueClient, execClient CommandRunner, opts BackportOpts) (*github.PullRequest, error) {
	// Remove any `backport` related labels from the original PR, and mark this PR as a "backport"
	labels := []*github.Label{
		{Name: github.String("backport")},
	}

	for _, v := range opts.Labels {
		if strings.Contains(v.GetName(), "backport") {
			continue
		}

		labels = append(labels, v)
	}

	opts.Labels = labels
	pr, err := backport(ctx, log, backportClient, issueClient, execClient, opts)
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
