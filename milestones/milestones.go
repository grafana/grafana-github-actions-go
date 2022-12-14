package milestones

import (
	"context"
	"errors"
	"fmt"

	gh "github.com/google/go-github/v47/github"
)

var RepoName = "grafana-ci-sandbox"

var (
	ErrorGitHub            = errors.New("gitHub returned an error")
	ErrorMilestoneNotFound = errors.New("did not find milestone")
)

type CloseMilestoneClient interface {
	EditMilestone(ctx context.Context, owner string, repo string, number int, milestone *gh.Milestone) (*gh.Milestone, *gh.Response, error)
	ListMilestones(ctx context.Context, owner string, repo string, opts *gh.MilestoneListOptions) ([]*gh.Milestone, *gh.Response, error)
}

type RemoveMilestoneClient interface {
	ListByRepo(ctx context.Context, owner string, repo string, opts *gh.IssueListByRepoOptions) ([]*gh.Issue, *gh.Response, error)
	RemoveMilestone(ctx context.Context, owner, repo string, issueNumber int) (*gh.Issue, *gh.Response, error)
}

type Commenter interface {
	CreateComment(ctx context.Context, owner string, repo string, number int, comment *gh.IssueComment) (*gh.IssueComment, *gh.Response, error)
}

type MilestoneRemover interface {
	RemoveMilestoneClient
	Commenter
}

type AdjustMilestoneClient interface {
	CloseMilestoneClient
	RemoveMilestoneClient
}

func FindMilestone(ctx context.Context, lister AdjustMilestoneClient, currentVersion string) (*gh.Milestone, error) {
	// List open milestones of repo
	milestones, _, err := lister.ListMilestones(ctx, "grafana", RepoName, &gh.MilestoneListOptions{})

	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrorGitHub, err)
	}

	// Get the milestone with the desired name
	var milestone *gh.Milestone
	for _, ms := range milestones {
		if ms.Title != nil && (*ms.Title == currentVersion) {
			milestone = ms
		}
	}

	if milestone == nil {
		return nil, fmt.Errorf(`%w: milestone %s doesn't exist`, ErrorMilestoneNotFound, currentVersion)
	}

	return milestone, nil
}
