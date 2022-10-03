package main

import (
	"context"
	"errors"
	"testing"

	gh "github.com/google/go-github/v47/github"
)

func TestReadArg(t *testing.T) {
	// If there are less than 3 args, return an err
	t.Run("If there are less than 3 arguments, return an error", func(t *testing.T) {
		token, currentVersion, err := readArgs([]string{})
		if token != "" {
			t.Error("token is not empty")
		}
		if currentVersion != "" {
			t.Error("current version is not empty")
		}
		if err == nil {
			t.Error("error is empty")
		}
	})
	// If there are correct amount of args, return them
	t.Run("If there are 3 or more arguments, return them", func(t *testing.T) {
		token, currentVersion, err := readArgs([]string{"/bin/go", "1234", "version"})
		if token != "1234" {
			t.Error("token does not equal 1234, token is", token)
		}
		if currentVersion != "version" {
			t.Error("current version does not equal version, current version is", currentVersion)
		}
		if err != nil {
			t.Error("error is not empty", err.Error())
		}
	})
}

// test should clearly say what the expectation is
func TestListMilestone(t *testing.T) {
	t.Run("If the milestone does not exist, return an error", func(t *testing.T) {
		m := &testMilestoneClient{
			milestones: []string{"v1.0.0-alpha", "v2.0", "v3.0", "v4.0"},
		}
		ctx := context.Background()
		ms, err := findMilestone(ctx, m, "v1.0.0")
		// if we return an err, milestone should be nil
		if ms != nil {
			t.Error("milestone should be nil")
		}
		if !errors.Is(err, errorMilestoneNotFound) {
			t.Error("error is the wrong type:", err)
		}
	})

	t.Run("If GitHub returns an error, return an error", func(t *testing.T) {
		m := &testMilestoneClient{
			milestones:  []string{"v1.0.0-alpha", "v2.0", "v3.0", "v4.0"},
			returnError: true,
		}
		ctx := context.Background()
		ms, err := findMilestone(ctx, m, "v1.0.0")
		// if we return an err, milestone should be nil
		if ms != nil {
			t.Error("milestone should be nil")
		}
		if !errors.Is(err, errorGitHub) {
			t.Error("error is the wrong type:", err)
		}
	})
}

func TestEditMilestone(t *testing.T) {
	t.Run("Milestone state successfully set to closed", func(t *testing.T) {
		num := 1
		ms := gh.Milestone{Number: &num}
		m := &testMilestoneClient{
			expectedMilestoneState: "closed",
		}
		err := editMilestone(context.Background(), m, "v1.0.0", &ms)
		if *ms.State != m.expectedMilestoneState {
			t.Error("milestone state is not closed")
		}
		if err != nil {
			t.Error("failed to update milestone")
		}
	})
	t.Run("If GitHub returns an error, return an error", func(t *testing.T) {
		num := 1
		ms := gh.Milestone{Number: &num}
		m := &testMilestoneClient{
			expectedMilestoneState: "closed",
			returnError:            true,
		}
		err := editMilestone(context.Background(), m, "v1.0.0", &ms)
		if !errors.Is(err, errorMilestoneNotUpdated) {
			t.Error("error is the wrong type:", err)
		}
	})
}

type testMilestoneClient struct {
	milestones             []string
	expectedMilestoneState string
	returnError            bool
}

func (m *testMilestoneClient) EditMilestone(ctx context.Context, owner string, repo string, number int, milestone *gh.Milestone) (*gh.Milestone, *gh.Response, error) {
	//check milestone status is definitely closed
	if m.returnError {
		return nil, nil, errors.New("github failed")
	}
	return milestone, nil, nil
}

// converting list of strings we provided into list of GH milestones, purpose is so we can write a test
func (m *testMilestoneClient) ListMilestones(ctx context.Context, owner string, repo string, opts *gh.MilestoneListOptions) ([]*gh.Milestone, *gh.Response, error) {
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
