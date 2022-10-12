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
		// if ms != nil {
		// 	t.Error("milestone should be nil")
		// }
		// if !errors.Is(err, errorMilestoneNotFound) {
		// 	t.Error("error is the wrong type:", err)
		// }
	})

	t.Run("If GitHub returns an error, return an error", func(t *testing.T) {
		m := &testMilestoneClient{
			milestones:  []string{"v1.0.0-alpha", "v2.0", "v3.0", "v4.0"},
			returnError: true,
		}
		ms, err := findMilestone(context.Background(), m, "v1.0.0")
		require.Nil(t, ms)
		require.ErrorContains(t, err, errorGitHub.Error())
		// if ms != nil {
		// 	t.Error("milestone should be nil")
		// }
		// if !errors.Is(err, errorGitHub) {
		// 	t.Error("error is the wrong type:", err)
		// }
	})
}

func TestUpdateMilestone(t *testing.T) {
	t.Run("Milestone state successfully set to closed", func(t *testing.T) {
		num := 1
		ms := gh.Milestone{Number: &num}
		m := &testMilestoneClient{
			expectedMilestoneState: "closed",
		}
		err := updateMilestone(context.Background(), m, "v1.0.0", &ms)
		require.NoError(t, err)
		require.Equal(t, *(ms.State), m.expectedMilestoneState)
	})
	t.Run("If GitHub returns an error, return an error", func(t *testing.T) {
		num := 1
		ms := gh.Milestone{Number: &num}
		m := &testMilestoneClient{
			expectedMilestoneState: "closed",
			returnError:            true,
		}
		err := updateMilestone(context.Background(), m, "v1.0.0", &ms)
		require.Error(t, err, errors.New("did not find milestone: v1.0.0"))
	})
}

type testMilestoneClient struct {
	milestones             []string
	expectedMilestoneState string
	returnError            bool
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

func (m *testMilestoneClient) EditMilestone(ctx context.Context, owner string, repo string, number int, milestone *gh.Milestone) (*gh.Milestone, *gh.Response, error) {
	// Check milestone status is definitely closed
	if m.returnError {
		return nil, nil, errors.New("github failed")
	}
	return milestone, nil, nil
}
