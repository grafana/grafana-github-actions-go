package main

import (
	"context"
	"fmt"
	"grafana-github-actions-go/args"
	"grafana-github-actions-go/milestones"
	"log"
	"os"
	"strconv"

	gh "github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

func findIssues(ctx context.Context, lister milestones.RemoveMilestoneClient, milestone *gh.Milestone, currentVersion string) ([]*gh.Issue, error) {
	issues, _, err := lister.ListByRepo(ctx, "grafana", milestones.RepoName, &gh.IssueListByRepoOptions{Milestone: strconv.Itoa(*milestone.Number)})
	if err != nil {
		return nil, fmt.Errorf("get list of issue by milestone %s number %d failed %s", currentVersion, *milestone.Number, err.Error())
	}
	return issues, nil
}

func removeMilestone(ctx context.Context, deleter milestones.RemoveMilestoneClient, issues []*gh.Issue, milestone *gh.Milestone, currentVersion string) error {
	for _, issue := range issues {
		_, _, err := deleter.RemoveMilestone(ctx, "grafana", milestones.RepoName, *issue.Number)
		if err != nil {
			return fmt.Errorf("remove Milestone %s for issue number: %d failed", currentVersion, issue.Number)
		}

		commentContent := fmt.Sprintf("This pull request was removed from the %s milestone because %s is currently being released.", currentVersion, currentVersion)
		_, _, err = deleter.CreateComment(ctx, "grafana", milestones.RepoName, *issue.Number, &gh.IssueComment{Body: &commentContent})
		if err != nil {
			return fmt.Errorf("the add comment issue %d failed %s", *issue.Number, err.Error())
		}
	}
	return nil
}

func main() {
	token, currentVersion, err := args.ReadArgs(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	client := gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))

	milestone, err := milestones.FindMilestone(ctx, client.Issues, currentVersion)
	if err != nil {
		log.Fatal(err)
	}

	issues, err := findIssues(ctx, client.Issues, milestone, currentVersion)
	if err != nil {
		log.Fatal(err)
	}

	removeMilestone(ctx, client.Issues, issues, milestone, currentVersion)
}
