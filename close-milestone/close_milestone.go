package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/v47/github"
)

func main() {
	fmt.Println("Make it workkkkk")

	client := github.NewClient(nil)
	opts := &github.MilestoneListOptions{}

	//func (s *IssuesService) ListMilestones(ctx context.Context, owner string, repo string, opts *MilestoneListOptions) ([]*Milestone, *Response, error)
	milestones, r, e := client.Issues.ListMilestones(context.Background(), "yangkb09", "grafana-github-actions-go", opts)

	fmt.Println("milestones here!!!", milestones)
	fmt.Println("response and err!!!", r, e)
}
