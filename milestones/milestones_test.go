package milestones

import (
	"context"
	"errors"
	"testing"

	gh "github.com/google/go-github/v47/github"
	"github.com/stretchr/testify/require"
)

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

func (m *TestMilestoneClient) ListByRepo(ctx context.Context, owner string, repo string, opts *gh.IssueListByRepoOptions) (issue []*gh.Issue, res *gh.Response, err error) {
	if m.returnError {
		return nil, nil, errors.New("github failed")
	}
	return issue, nil, nil
}

func (m *TestMilestoneClient) RemoveMilestone(ctx context.Context, owner, repo string, issueNumber int) (issue *gh.Issue, res *gh.Response, err error) {
	if m.returnError {
		return nil, nil, errors.New("github failed")
	}
	return issue, nil, nil
}

func (m *TestMilestoneClient) CreateComment(ctx context.Context, owner string, repo string, number int, comment *gh.IssueComment) (*gh.IssueComment, *gh.Response, error) {
	if m.returnError {
		return nil, nil, errors.New("github failed")
	}
	return comment, nil, nil
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
