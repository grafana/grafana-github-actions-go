package utils

import (
	"context"
	"errors"
	"testing"

	gh "github.com/google/go-github/v47/github"
	"github.com/stretchr/testify/require"
)

func TestReadArg(t *testing.T) {
	// If there are less than 3 args, return an err
	t.Run("If there are less than 3 arguments, return an error", func(t *testing.T) {
		token, currentVersion, err := ReadArgs([]string{})
		require.Empty(t, token)
		require.Empty(t, currentVersion)
		require.NotNil(t, err)
	})
	// If there are correct amount of args, return them
	t.Run("If there are 3 or more arguments, return them", func(t *testing.T) {
		token, currentVersion, err := ReadArgs([]string{"/bin/go", "1234", "version"})
		require.Equal(t, "1234", token)
		require.Equal(t, "version", currentVersion)
		require.Nil(t, err)
	})
}

func TestFindMilestone(t *testing.T) {
	t.Run("If the milestone does not exist, return an error", func(t *testing.T) {
		m := &TestMilestoneClient{
			milestones: []string{"v1.0.0-alpha", "v2.0", "v3.0", "v4.0"},
		}
		ms, err := FindMilestone(context.Background(), m, "v1.0.0")
		require.Nil(t, ms)
		require.ErrorContains(t, err, ErrorMilestoneNotFound.Error())
	})

	t.Run("If GitHub returns an error, return an error", func(t *testing.T) {
		m := &TestMilestoneClient{
			milestones:  []string{"v1.0.0-alpha", "v2.0", "v3.0", "v4.0"},
			returnError: true,
		}
		ms, err := FindMilestone(context.Background(), m, "v1.0.0")
		require.Nil(t, ms)
		require.ErrorContains(t, err, ErrorGitHub.Error())
	})
}

type TestMilestoneClient struct {
	milestones             []string
	expectedMilestoneState string
	returnError            bool
}

func (m *TestMilestoneClient) ListMilestones(ctx context.Context, owner string, repo string, opts *gh.MilestoneListOptions) ([]*gh.Milestone, *gh.Response, error) {
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

func (m *TestMilestoneClient) EditMilestone(ctx context.Context, owner string, repo string, number int, milestone *gh.Milestone) (*gh.Milestone, *gh.Response, error) {
	// Check milestone status is definitely closed
	if m.returnError {
		return nil, nil, errors.New("github failed")
	}
	return milestone, nil, nil
}
