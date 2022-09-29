package main

import (
	"context"
	"fmt"
	"os"

	gh "github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

func main() {
	// Check if enough input parameters
	if len(os.Args) < 3 {
		fmt.Println("Not enough input parameters")
		os.Exit(1)
	}

	token := os.Args[1]
	currentVersion := os.Args[2]
	ctx := context.Background()
	client := gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))

	// List open milestones of repo
	milestones, _, err := client.Issues.ListMilestones(ctx, "grafana", "grafana-github-actions-go", &gh.MilestoneListOptions{State: "open"})

	if err != nil {
		os.Exit(1)
	}

	// Get the milestone with the desired name
	var milestone *gh.Milestone
	for _, ms := range milestones {
		if ms.Title != nil && (*ms.Title == currentVersion) {
			milestone = ms
		}
	}

	if milestone == nil {
		fmt.Printf(`Milestone %s doesn't exist %s`, currentVersion, err.Error())
		os.Exit(1)
	}

	// Update milestone status to "closed"
	milestone.State = gh.String("closed")

	_, _, err = client.Issues.EditMilestone(ctx, "grafana", "grafana-github-actions-go", *milestone.Number, milestone)
	if err != nil {
		fmt.Println("Close Milestone ", currentVersion, " for issue number: ", milestone.Number, " failed.", err)
		os.Exit(1)
	}
}
