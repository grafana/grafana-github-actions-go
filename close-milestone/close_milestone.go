package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"grafana-github-actions-go/args"
	"grafana-github-actions-go/milestones"

	gh "github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

func updateMilestone(ctx context.Context, editor milestones.CloseMilestoneClient, currentVersion string, milestone *gh.Milestone) error {
	// Update milestone status to "closed"
	milestone.State = gh.String("closed")

	_, _, err := editor.EditMilestone(ctx, "grafana", milestones.RepoName, *milestone.Number, milestone)
	if err != nil {
		return fmt.Errorf("did not find milestone: %s", milestone.String())
	}
	return nil
}

func main() {
	token, milestone, err := args.ReadArgs(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("closing milestone %s...", milestone)

	ctx := context.Background()
	client := gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))

	log.Printf("finding milestone %s...", milestone)
	m, err := milestones.FindMilestone(ctx, client.Issues, milestone)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("found milestone %s", milestone)

	log.Printf("updating milestone %s...", milestone)
	if err := updateMilestone(ctx, client.Issues, milestone, m); err != nil {
		log.Fatal(err)
	}
	log.Printf("done updating milestone %s", milestone)
	log.Printf("done closing milestone %s", milestone)
}
