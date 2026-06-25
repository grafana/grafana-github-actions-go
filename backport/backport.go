package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/go-github/v88/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghutil"
)

type BackportOpts struct {
	// PullRequestNumber is the integer ID of the pull request being backported
	PullRequestNumber int

	// SourceSHA is the commit hash that will be cherry-picked into a pull request targeting Target
	SourceSHA string

	// SourceTitle is the title of the source PR which will be reused in the backport PRs
	SourceTitle string

	// TitleTemplate controls backport PR title formatting via {{branch}} and {{title}} placeholders.
	TitleTemplate string

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

	// RunID is the ID of the run of the GitHub action that is running this.
	RunID string

	// GitToken is used to authenticate git fetch operations over HTTPS.
	// When set, CreateCherryPickBranch temporarily rewrites the origin remote
	// URL to embed the token, then restores it afterwards.
	GitToken string
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

func CreatePullRequest(ctx context.Context, client BackportClient, issueClient IssueClient, branch string, opts BackportOpts) (*github.PullRequest, error) {
	title := FormatBackportTitle(opts.TitleTemplate, opts.Target.Name, opts.SourceTitle)

	body := fmt.Sprintf("Backport %s from #%d <sup>[job run](https://github.com/%s/%s/actions/runs/%s)</sup>\n\n---\n\n%s", opts.SourceSHA, opts.PullRequestNumber, opts.Owner, opts.Repository, opts.RunID, opts.SourceBody)

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

func backport(ctx context.Context, log *slog.Logger, client BackportClient, issueClient IssueClient, refClient RefClient, gqlClient SignedCommitClient, runner CommandRunner, opts BackportOpts) (*github.PullRequest, error) {
	// 1. Run CLI commands to create a branch and cherry-pick
	//   * If the cherry-pick fails, write a comment in the source PR with instructions on manual backporting
	// 2. Publish the cherry-picked commit via createCommitOnBranch so it is signed by GitHub
	// 3. Open the pull request against the appropriate release branch
	branch := BackportBranch(opts.PullRequestNumber, opts.Target.Name)
	if err := CreateCherryPickBranch(ctx, runner, branch, opts); err != nil {
		return nil, fmt.Errorf("error cherry-picking: %w", err)
	}

	// Publish the cherry-picked commit via the GitHub API so it is signed by GitHub's web-flow
	// key (shown as Verified). Unlike the old `git push` path, this is intentionally NOT wrapped
	// in a retry: it consists of two non-idempotent API calls (createRef + createCommitOnBranch),
	// and blind retries after a partial success would either fail with 422 "ref already exists"
	// (handled inline in PublishViaAPI) or with "expected head oid" mismatches. The underlying
	// HTTP clients retry transient errors on their own.
	if err := PublishViaAPI(ctx, runner, refClient, gqlClient, branch, opts); err != nil {
		return nil, fmt.Errorf("error publishing backport branch: %w", err)
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

func Backport(ctx context.Context, log *slog.Logger, backportClient BackportClient, commentClient CommentClient, issueClient IssueClient, refClient RefClient, gqlClient SignedCommitClient, execClient CommandRunner, opts BackportOpts) (*github.PullRequest, error) {
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
	pr, err := backport(ctx, log, backportClient, issueClient, refClient, gqlClient, execClient, opts)
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
