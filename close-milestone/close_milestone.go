package main

import (
	"context"
	"log"
	"os"

	"grafana-github-actions-go/args"
	"grafana-github-actions-go/milestones"

	gh "github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

func updateMilestone(ctx context.Context, editor milestones.CloseMilestoneClient, currentVersion string, milestone *gh.Milestone) error {
	_, _, err := editor.EditMilestone(ctx, "grafana", milestones.RepoName, *milestone.Number, &gh.Milestone{
		Title:       milestone.Title,
		Description: milestone.Description,
		DueOn:       milestone.DueOn,
		State:       gh.String("closed"),
	})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	token, milestone, err := args.ReadArgs(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	client := gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))

	m, err := milestones.FindMilestone(ctx, client.Issues, milestone)
	if err != nil {
		log.Fatal(err)
	}

	if err := updateMilestone(ctx, client.Issues, milestone, m); err != nil {
		log.Fatal(err)
	}
}
