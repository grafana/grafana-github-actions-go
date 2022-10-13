package main

import (
	"context"
	"fmt"
	"grafana-github-actions-go/utils"
	"log"
	"os"

	gh "github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

func updateMilestone(ctx context.Context, editor utils.CloseMilestoneClient, currentVersion string, milestone *gh.Milestone) error {
	// Update milestone status to "closed"
	milestone.State = gh.String("closed")

	_, _, err := editor.EditMilestone(ctx, "grafana", utils.RepoName, *milestone.Number, milestone)
	if err != nil {
		return fmt.Errorf("did not find milestone: %s", milestone.String())
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

	milestone, err := utils.FindMilestone(ctx, client.Issues, currentVersion)
	if err != nil {
		log.Fatal(err)
	}

	updateMilestone(ctx, client.Issues, currentVersion, milestone)
}
