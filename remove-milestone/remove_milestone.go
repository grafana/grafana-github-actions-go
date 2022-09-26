package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	gh "github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

func main() {
	// we need to get all open issue with milestone and remove the milestone from them
	// we need to get all PR opened with milestone and remove the milestone from them
	if len(os.Args) < 3 {
		fmt.Println("Not enough input parameters")
		os.Exit(1)
	}

	// Just using something simple to dmeonstrate using the github package here
	token := os.Args[1]
	currentVersion := os.Args[2]
	ctx := context.Background()
	client := gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))

	// List milestone of repo, so that we can get the milestone number corresponding to the milestone name
	milestones, _, err := client.Issues.ListMilestones(ctx, "grafana", "grafana-github-actions-go", &gh.MilestoneListOptions{State: "open"})

	var milestoneNum *int
	for _, ms := range milestones {
		if ms.Title != nil && (*ms.Title == currentVersion) {
			milestoneNum = ms.Number
		}
	}

	if milestoneNum == nil {
		fmt.Printf(`Milestone %s doesn't exist %s`, currentVersion, err.Error())
		os.Exit(1)
	}

	issues, _, err := client.Issues.ListByRepo(ctx, "grafana", "grafana-github-actions-go", &gh.IssueListByRepoOptions{Milestone: strconv.Itoa(*milestoneNum)})
	if err != nil {
		fmt.Printf("Get list of issue by milestone %s number %d failed %s", currentVersion, *milestoneNum, err.Error())
		os.Exit(1)
	}

	for _, ele := range issues {
		_, _, err := client.Issues.RemoveMilestone(ctx, "grafana", "grafana-github-actions-go", *ele.Number)
		if err != nil {
			fmt.Println("Remove Milestone ", currentVersion, " for issue number: ", ele.Number, " failed.", err)
			os.Exit(1)
		}

		commentContent := fmt.Sprintf("This pull request was removed from the %s milestone because %s is currently being released.", currentVersion, currentVersion)
		client.Issues.CreateComment(ctx, "grafana", "grafana-github-actions-go", *ele.Number, &gh.IssueComment{Body: &commentContent})
		if err != nil {
			fmt.Printf("The add comment issue %d failed %s", *ele.Number, err.Error())
			os.Exit(1)
		}
	}
}
