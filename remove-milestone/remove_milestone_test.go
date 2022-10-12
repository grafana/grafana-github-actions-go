package main

import (
	"context"
	"errors"
	"testing"

	gh "github.com/google/go-github/v47/github"
	"github.com/stretchr/testify/require"
)

func TestFindMilestone(t *testing.T) {
	t.Run("If the milestone does not exist, return an error", func(t *testing.T) {
		m := &testMilestoneClient{
			milestones: []string{"v1.0.0-alpha", "v2.0", "v3.0", "v4.0"},
		}
		ms, err := findMilestone(context.Background(), m, "v1.0.0")
		require.Nil(t, ms)
		require.ErrorContains(t, err, errorMilestoneNotFound.Error())
	})

	t.Run("If GitHub returns an error, return an error", func(t *testing.T) {
		m := &testMilestoneClient{
			milestones:  []string{"v1.0.0-alpha", "v2.0", "v3.0", "v4.0"},
			returnError: true,
		}
		ms, err := findMilestone(context.Background(), m, "v1.0.0")
		require.Nil(t, ms)
		require.ErrorContains(t, err, errorGitHub.Error())
	})
}

//we mock things bc we dont want to make api calls to GH
//actual GA should talk to GH, when unit testing we just test the logic (that's why we have expected and actual output)
//if we want to test actual functionality, we write integration tests

func TestRemoveMilestone(t *testing.T) {
	t.Run("If milestone exists, remove it from the issue", func(t *testing.T) {
		m := &testMilestoneClient{
			milestones: []string{"v1.0.0-alpha", "v2.0", "v3.0", "v4.0"},
		}
		err := removeMilestone(context.Background(), m, nil, nil, "v2.0")
		require.NoError(t, err)
	})

	t.Run("If milestone does not exist, throw error", func(t *testing.T) {
		m := &testMilestoneClient{
			milestones:  []string{"v1.0.0-alpha", "v2.0", "v3.0", "v4.0"},
			returnError: true,
		}
		id := int(1)
		issues := []*gh.Issue{&gh.Issue{Number: &id}}
		err := removeMilestone(context.Background(), m, issues, nil, "v2.0")
		require.Error(t, err, errorGitHub.Error())
	})
}

type testMilestoneClient struct {
	milestones  []string
	returnError bool
}

func (m *testMilestoneClient) ListMilestones(ctx context.Context, owner string, repo string, opts *gh.MilestoneListOptions) ([]*gh.Milestone, *gh.Response, error) {
	// Convert list of strings into list of GH milestones for testing
	milestones := make([]*gh.Milestone, len(m.milestones))
	for i := range m.milestones {
		milestones[i] = &gh.Milestone{
			Title: &m.milestones[i],
		}
	}
	if m.returnError {
		return nil, nil, errors.New("github failed")
	}
	return milestones, nil, nil
}

func (m *testMilestoneClient) CreateComment(ctx context.Context, owner string, repo string, number int, comment *gh.IssueComment) (*gh.IssueComment, *gh.Response, error) {
	if m.returnError {
		return nil, nil, errors.New("github failed")
	}
	return comment, nil, nil
}

func (m *testMilestoneClient) ListByRepo(ctx context.Context, owner string, repo string, opts *gh.IssueListByRepoOptions) (issue []*gh.Issue, res *gh.Response, err error) {
	if m.returnError {
		return nil, nil, errors.New("github failed")
	}
	return issue, nil, nil
}

func (m *testMilestoneClient) RemoveMilestone(ctx context.Context, owner, repo string, issueNumber int) (issue *gh.Issue, res *gh.Response, err error) {
	if m.returnError {
		return nil, nil, errors.New("github failed")
	}
	return issue, nil, nil
}
