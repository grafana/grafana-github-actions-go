package main

import (
	"context"
	"fmt"
	"os"

	gh "github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Not enough input parameters")
		os.Exit(1)
	}

	token := os.Args[1]
	currentVersion := os.Args[2]
	ctx := context.Background()
	client := gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))
	// client.Issues.DeleteMilestone()

	// List milestone of repo, so that we can get the milestone number corresponding to the milestone name
	milestones, _, err := client.Issues.ListMilestones(ctx, "grafana", "grafana-github-actions-go", &gh.MilestoneListOptions{State: "open"})

	if err != nil {
		os.Exit(1)
	}

	var milestoneNum *int
	for _, ms := range milestones {
		if ms.Title != nil && (*ms.Title == currentVersion) {
			milestoneNum = ms.Number
		}
	}

	fmt.Println("MILESTONE NUM HERE!!!", milestoneNum)

}
