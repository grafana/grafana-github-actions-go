package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	gh "github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

func readArgs(args []string) (string, string, error) {
	// Check if enough input parameters
	if len(args) < 3 {
		return "", "", fmt.Errorf("not enough input parameters")
	}

	token := args[1]
	currentVersion := args[2]
	return token, currentVersion, nil
}

type milestoneClient interface {
	ListMilestones(ctx context.Context, owner string, repo string, opts *gh.MilestoneListOptions) ([]*gh.Milestone, *gh.Response, error)
	EditMilestone(ctx context.Context, owner string, repo string, number int, milestone *gh.Milestone) (*gh.Milestone, *gh.Response, error)
}

//cant have pointer to interface, bc the thing that satisfies interface itself is a pointer
//interfaces cant have properties

var (
	errorGitHub              = errors.New("gitHub returned an error")
	errorMilestoneNotFound   = errors.New("did not find milestone")
	errorMilestoneNotUpdated = errors.New("did not update milestone")
)

func findMilestone(ctx context.Context, lister milestoneClient, currentVersion string) (*gh.Milestone, error) { //ctx means func could do something async
	// List open milestones of repo
	milestones, _, err := lister.ListMilestones(ctx, "grafana", "grafana-github-actions-go", &gh.MilestoneListOptions{State: "open"})

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

func editMilestone(ctx context.Context, editor milestoneClient, currentVersion string, milestone *gh.Milestone) error {
	// Update milestone status to "closed"
	milestone.State = gh.String("closed")

	_, _, err := editor.EditMilestone(ctx, "grafana", "grafana-github-actions-go", *milestone.Number, milestone)
	if err != nil {
		return errorMilestoneNotUpdated
	}
	return nil
}

func main() {
	token, currentVersion, err := readArgs(os.Args)
	if err != nil {
		log.Fatal(err) //logs err and quits program
	}

	ctx := context.Background()
	client := gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))

	milestone, err := findMilestone(ctx, client.Issues, currentVersion)
	if err != nil {
		log.Fatal(err)
	}

	editMilestone(ctx, client.Issues, currentVersion, milestone)
}
