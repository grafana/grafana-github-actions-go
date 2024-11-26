// Package main provides PR migration functionality for updating pull request base branches and notifying users of the update.
// It centers around the Client interface which defines the contract for PR operations:
// - `GitHubClient` implements Client for production GitHub API calls
// - `MockClient` implements Client for testing
package main

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

func (g *GitHubClient) CreateComment(ctx context.Context, number int, body string) error {
	_, _, err := g.Client.Issues.CreateComment(ctx, g.Owner, g.Repo, number, &github.IssueComment{
		Body: github.String(body),
	})

	return err
}

func UpdateBaseBranch(ctx context.Context, client Client, pr PullRequestInfo, nextBranch string) error {
	return client.EditPR(ctx, pr.Number, nextBranch)
}

func NotifyUser(ctx context.Context, client Client, pr PullRequestInfo, prevBranch, nextBranch string, succeeded bool) error {
	comment := buildComment(pr.AuthorName, prevBranch, nextBranch, succeeded)
	return client.CreateComment(ctx, pr.Number, comment)
}

// Comment templates for PR notifications
const (
	successCommentTemplate = "Hello @%s, we've noticed that the original base branch `%s` for this PR is no longer a release candidate. " +
		"We've automatically updated your PR's base branch to the current release target: `%s`. " +
		"Please review and resolve any potential merge conflicts. " +
		"If this PR is not merged it will NOT be included in the next release. Thanks!"

	failureCommentTemplate = "Hello @%s, we've noticed that the original base branch `%s` for this PR is no longer a release candidate. " +
		"We attempted to automatically update your PR's base branch to the current release target `%s`, but encountered an error. " +
		"Please manually update your PR's base branch to resolve any merge conflicts. " +
		"If this PR is not rebased and merged it will NOT be included in the next release. Thanks!"
)

func buildComment(authorName, prevBranch, nextBranch string, succeeded bool) string {
	if succeeded {
		return fmt.Sprintf(successCommentTemplate, authorName, prevBranch, nextBranch)
	}
	return fmt.Sprintf(failureCommentTemplate, authorName, prevBranch, nextBranch)
}
