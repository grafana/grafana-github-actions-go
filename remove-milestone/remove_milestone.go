package main

import (
	"context"
	"errors"
	"fmt"
	"grafana-github-actions-go/utils"
	"log"
	"os"
	"strconv"

	gh "github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

var repoName = "grafana-github-actions-go"

var (
	errorGitHub            = errors.New("gitHub returned an error")
	errorMilestoneNotFound = errors.New("did not find milestone")
)

type milestoneClient interface {
	ListMilestones(ctx context.Context, owner string, repo string, opts *gh.MilestoneListOptions) ([]*gh.Milestone, *gh.Response, error)
	ListByRepo(ctx context.Context, owner string, repo string, opts *gh.IssueListByRepoOptions) ([]*gh.Issue, *gh.Response, error)
	RemoveMilestone(ctx context.Context, owner, repo string, issueNumber int) (*gh.Issue, *gh.Response, error)
	CreateComment(ctx context.Context, owner string, repo string, number int, comment *gh.IssueComment) (*gh.IssueComment, *gh.Response, error)
}

func findMilestone(ctx context.Context, lister milestoneClient, currentVersion string) (*gh.Milestone, error) {
	// List open milestones of repo
	milestones, _, err := lister.ListMilestones(ctx, "grafana", repoName, &gh.MilestoneListOptions{State: "open"})

	if err != nil {
		return nil, fmt.Errorf("%w: %s", errorGitHub, err)
	}

	// Get the milestone with the desired name
	var milestone *gh.Milestone
	for _, ms := range milestones {
		if ms.Title != nil && (*ms.Title == currentVersion) {
			milestone = ms
		}
	}

	if milestone == nil {
		return nil, fmt.Errorf(`%w: milestone %s doesn't exist`, errorMilestoneNotFound, currentVersion)
	}

	return milestone, nil
}

func findIssues(ctx context.Context, lister milestoneClient, milestone *gh.Milestone, currentVersion string) ([]*gh.Issue, error) {
	issues, _, err := lister.ListByRepo(ctx, "grafana", repoName, &gh.IssueListByRepoOptions{Milestone: strconv.Itoa(*milestone.Number)})
	if err != nil {
		return nil, fmt.Errorf("get list of issue by milestone %s number %d failed %s", currentVersion, *milestone.Number, err.Error())
	}
	return issues, nil
}

func deleteMilestone(ctx context.Context, deleter milestoneClient, issues []*gh.Issue, milestone *gh.Milestone, currentVersion string) error {
	for _, issue := range issues {
		_, _, err := deleter.RemoveMilestone(ctx, "grafana", repoName, *issue.Number)
		if err != nil {
			return fmt.Errorf("remove Milestone %s for issue number: %d failed", currentVersion, issue.Number)
		}

		commentContent := fmt.Sprintf("This pull request was removed from the %s milestone because %s is currently being released.", currentVersion, currentVersion)
		deleter.CreateComment(ctx, "grafana", repoName, *issue.Number, &gh.IssueComment{Body: &commentContent})
		if err != nil {
			return fmt.Errorf("the add comment issue %d failed %s", *issue.Number, err.Error())
		}
	}
	return nil
}

func main() {
	token, currentVersion, err := utils.ReadArgs(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	client := gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))

	milestone, err := findMilestone(ctx, client.Issues, currentVersion)
	if err != nil {
		log.Fatal(err)
	}

	issues, err := findIssues(ctx, client.Issues, milestone, currentVersion)
	if err != nil {
		log.Fatal(err)
	}

	deleteMilestone(ctx, client.Issues, issues, milestone, currentVersion)
}
