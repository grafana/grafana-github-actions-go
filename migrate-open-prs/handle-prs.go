package handleprs

import (
	"context"
	"fmt"

	"github.com/google/go-github/v50/github"
)

type PullRequestInfo struct {
	Number     int
	AuthorName string
}

type Client interface {
	EditPR(ctx context.Context, number int, branch string) error
	CreateComment(ctx context.Context, number int, body string) error
}

type GitHubClient struct {
	Client *github.Client
	Owner  string
	Repo   string
}

func (g *GitHubClient) EditPR(ctx context.Context, number int, branch string) error {
	_, _, err := g.Client.PullRequests.Edit(ctx, g.Owner, g.Repo, number, &github.PullRequest{
		Base: &github.PullRequestBranch{
			Ref: github.String(branch),
		},
	})

	return err
}
