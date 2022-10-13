package main

import (
	"context"
	"errors"
	"testing"

	"grafana-github-actions-go/utils"

	gh "github.com/google/go-github/v47/github"
	"github.com/stretchr/testify/require"
)

func TestUpdateMilestone(t *testing.T) {
	t.Run("Milestone state successfully set to closed", func(t *testing.T) {
		num := 1
		ms := gh.Milestone{Number: &num}
		m := &utils.TestMilestoneClient{
			expectedMilestoneState: "closed",
		}
		err := updateMilestone(context.Background(), m, "v1.0.0", &ms)
		require.NoError(t, err)
		require.Equal(t, *(ms.State), m.expectedMilestoneState)
	})
	t.Run("If GitHub returns an error, return an error", func(t *testing.T) {
		num := 1
		ms := gh.Milestone{Number: &num}
		m := &utils.TestMilestoneClient{
			expectedMilestoneState: "closed",
			returnError:            true,
		}
		err := updateMilestone(context.Background(), m, "v1.0.0", &ms)
		require.Error(t, err, errors.New("did not find milestone: v1.0.0"))
	})
}

// type testMilestoneClient struct {
// 	milestones             []string
// 	expectedMilestoneState string
// 	returnError            bool
// }

// func (m *testMilestoneClient) ListMilestones(ctx context.Context, owner string, repo string, opts *gh.MilestoneListOptions) ([]*gh.Milestone, *gh.Response, error) {
// 	// Convert list of strings into list of GH milestones for testing
// 	milestones := make([]*gh.Milestone, len(m.milestones))
// 	for i := range m.milestones {
// 		milestones[i] = &gh.Milestone{
// 			Title: &m.milestones[i],
// 		}
// 	}
// 	if m.returnError {
// 		return nil, nil, errors.New("github failed")
// 	}
// 	return milestones, nil, nil
// }

// func (m *testMilestoneClient) EditMilestone(ctx context.Context, owner string, repo string, number int, milestone *gh.Milestone) (*gh.Milestone, *gh.Response, error) {
// 	// Check milestone status is definitely closed
// 	if m.returnError {
// 		return nil, nil, errors.New("github failed")
// 	}
// 	return milestone, nil, nil
// }

// func (m *testMilestoneClient) CreateComment(ctx context.Context, owner string, repo string, number int, comment *gh.IssueComment) (*gh.IssueComment, *gh.Response, error) {
// 	if m.returnError {
// 		return nil, nil, errors.New("github failed")
// 	}
// 	return comment, nil, nil
// }

// func (m *testMilestoneClient) ListByRepo(ctx context.Context, owner string, repo string, opts *gh.IssueListByRepoOptions) (issue []*gh.Issue, res *gh.Response, err error) {
// 	if m.returnError {
// 		return nil, nil, errors.New("github failed")
// 	}
// 	return issue, nil, nil
// }

// func (m *testMilestoneClient) RemoveMilestone(ctx context.Context, owner, repo string, issueNumber int) (issue *gh.Issue, res *gh.Response, err error) {
// 	if m.returnError {
// 		return nil, nil, errors.New("github failed")
// 	}
// 	return issue, nil, nil
// }
