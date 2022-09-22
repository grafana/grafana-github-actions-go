package main

import (
	"fmt"

	gh "github.com/google/go-github/v47/github"
)

func main() {
	fmt.Println("Make it workkkkk")

	milestones, r, e := gh.ListMilestones(_, "yangkb09", "grafana-github-actions-go", _)

	fmt.Println("milestones here!!!", milestones)

	//func (s *IssuesService) ListMilestones(ctx context.Context, owner string, repo string, opts *MilestoneListOptions) ([]*Milestone, *Response, error)
}
