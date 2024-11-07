package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func RequireContainsLabel(t *testing.T, labels []*github.Label, label *github.Label) {
	t.Helper()

	for _, v := range labels {
		if v.GetName() == label.GetName() {
			return
		}
	}

	t.Fatalf("label '%+v' not found in set '%+v'", label, labels)
}
func RequireNotContainsLabel(t *testing.T, labels []*github.Label, label *github.Label) {
	t.Helper()

	for _, v := range labels {
		if v.GetName() == label.GetName() {
			t.Fatalf("label '%+v' found in set '%+v'", label, labels)
		}
	}
}

func TestMostRecentBranch(t *testing.T) {
	assertError := func(t *testing.T, major, minor string, branches []string) {
		t.Helper()
		b, err := MostRecentBranch(major, minor, branches)
		assert.Error(t, err)
		assert.Empty(t, b)
	}

	assertBranch := func(t *testing.T, major, minor string, branches []string, branch string) {
		t.Helper()
		b, err := MostRecentBranch(major, minor, branches)
		assert.NoError(t, err)
		assert.Equal(t, branch, b)
	}
	branches := []string{
		"release-11.0.1",
		"release-1.2.3",
		"release-11.0.1+security-01",
		"release-10.0.0",
		"release-10.2.3",
		"release-10.2.4",
		"release-10.2.4+security-01",
		"release-12.0.3",
		"release-12.1.3",
		"release-12.0.15",
		"release-12.1.15",
		"release-12.2.12",
	}

	assertError(t, "3", "2", branches)
	assertError(t, "4", "0", branches)
	assertError(t, "13", "0", branches)
	assertError(t, "10", "5", branches)
	assertError(t, "11", "8", branches)
	assertBranch(t, "11", "0", branches, "release-11.0.1")
	assertBranch(t, "12", "1", branches, "release-12.1.15")
	assertBranch(t, "12", "0", branches, "release-12.0.15")
	assertBranch(t, "1", "2", branches, "release-1.2.3")
	assertBranch(t, "10", "2", branches, "release-10.2.4")
}

func TestBackportTarget(t *testing.T) {
	assertError := func(t *testing.T, label *github.Label, branches []string) {
		t.Helper()
		b, err := BackportTarget(label, branches)
		assert.Error(t, err)
		assert.Empty(t, b)
	}

	assertBranch := func(t *testing.T, label *github.Label, branches []string, branch string) {
		t.Helper()
		b, err := BackportTarget(label, branches)
		assert.NoError(t, err)
		assert.Equal(t, branch, b)
	}

	branches := []string{
		"release-11.0.1",
		"release-1.2.3",
		"release-11.0.1+security-01",
		"release-10.0.0",
		"release-10.2.3",
		"release-10.2.4",
		"release-10.2.4+security-01",
		"release-12.0.3",
		"release-12.1.3",
		"release-12.0.15",
		"release-12.1.15",
		"release-12.2.12",
	}

	assertError(t, &github.Label{
		Name: github.String("backport v3.2.x"),
	}, branches)
	assertError(t, &github.Label{
		Name: github.String("backport v4.0.x"),
	}, branches)
	assertError(t, &github.Label{
		Name: github.String("backport v13.0.x"),
	}, branches)
	assertError(t, &github.Label{
		Name: github.String("backport v10.5.x"),
	}, branches)
	assertError(t, &github.Label{
		Name: github.String("backport v11.8.x"),
	}, branches)
	assertBranch(t, &github.Label{
		Name: github.String("backport v11.0.x"),
	}, branches, "release-11.0.1")
	assertBranch(t, &github.Label{
		Name: github.String("backport v12.1.x"),
	}, branches, "release-12.1.15")
	assertBranch(t, &github.Label{
		Name: github.String("backport v12.0.x"),
	}, branches, "release-12.0.15")
	assertBranch(t, &github.Label{
		Name: github.String("backport v1.2.x"),
	}, branches, "release-1.2.3")
	assertBranch(t, &github.Label{
		Name: github.String("backport v10.2.x"),
	}, branches, "release-10.2.4")
}

type TestBackportClient struct {
	CreateFunc        func(ctx context.Context, owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
	CreateCommentFunc func(ctx context.Context, owner, repo string, number int, comment *github.PullRequestComment) (*github.PullRequestComment, *github.Response, error)
	EditFunc          func(ctx context.Context, owner string, repo string, number int, pull *github.PullRequest) (*github.PullRequest, *github.Response, error)
}

func (c *TestBackportClient) Create(ctx context.Context, owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	return c.CreateFunc(ctx, owner, repo, pull)
}
func (c *TestBackportClient) CreateComment(ctx context.Context, owner, repo string, number int, comment *github.PullRequestComment) (*github.PullRequestComment, *github.Response, error) {
	return c.CreateCommentFunc(ctx, owner, repo, number, comment)
}
func (c *TestBackportClient) Edit(ctx context.Context, owner string, repo string, number int, pull *github.PullRequest) (*github.PullRequest, *github.Response, error) {
	return c.EditFunc(ctx, owner, repo, number, pull)
}

func TestBackport(t *testing.T) {
	t.Run("Successful backport", func(t *testing.T) {
		createFn := func(ctx context.Context, owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
			return &github.PullRequest{
				Number: github.Int(101),
				Title:  pull.Title,
			}, nil, nil
		}
		createCommentFn := func(ctx context.Context, owner, repo string, number int, comment *github.PullRequestComment) (*github.PullRequestComment, *github.Response, error) {
			return comment, nil, nil
		}
		editFn := func(ctx context.Context, owner string, repo string, number int, pull *github.PullRequest) (*github.PullRequest, *github.Response, error) {
			return pull, nil, nil
		}

		runner := NewNoOpRunner()

		client := &TestBackportClient{
			CreateFunc:        createFn,
			CreateCommentFunc: createCommentFn,
			EditFunc:          editFn,
		}

		pr, err := Backport(context.Background(), client, runner, BackportOpts{
			PullRequestNumber: 100,
			SourceSHA:         "asdf1234",
			SourceTitle:       "Example Bug Fix",
			SourceBody:        "body",
			Target:            "release-12.0.0",
			Labels: []*github.Label{
				{
					Name: github.String("type/bug"),
				},
				{
					Name: github.String("backport v12.0.x"),
				},
			},
			Owner:      "grafana",
			Repository: "grafana",
		})

		require.NoError(t, err)
		require.Equal(t, *pr.Title, "[release-12.0.0] Example Bug Fix")
		require.Equal(t, *pr.Title, "[release-12.0.0] Example Bug Fix")

		// Ensure that all "backport" PRs have the "backport" label
		RequireContainsLabel(t, pr.Labels, &github.Label{
			Name: github.String("backport"),
		})

		// Ensure that backport labels which cause backport PRs are removed
		RequireNotContainsLabel(t, pr.Labels, &github.Label{
			Name: github.String("backport v12.0.x"),
		})
	})

	t.Run("Backport comments", func(t *testing.T) {
		// Simulate an error being returned from the 'git cherry-pick command'
		runner := NewErrorRunner(map[string]error{
			"git cherry-pick -x asdf1234": errors.New("The process '/usr/bin/git' failed with exit code 1"),
		})

		var comment *github.PullRequestComment
		createFn := func(ctx context.Context, owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
			return &github.PullRequest{
				Number: github.Int(101),
				Title:  pull.Title,
			}, nil, nil
		}
		createCommentFn := func(ctx context.Context, owner, repo string, number int, c *github.PullRequestComment) (*github.PullRequestComment, *github.Response, error) {
			comment = c
			return c, nil, nil
		}
		editFn := func(ctx context.Context, owner string, repo string, number int, pull *github.PullRequest) (*github.PullRequest, *github.Response, error) {
			return pull, nil, nil
		}

		client := &TestBackportClient{
			CreateFunc:        createFn,
			CreateCommentFunc: createCommentFn,
			EditFunc:          editFn,
		}

		_, err := Backport(context.Background(), client, runner, BackportOpts{
			PullRequestNumber: 100,
			SourceSHA:         "asdf1234",
			SourceTitle:       "Example Bug Fix",
			SourceBody:        "body",
			Target:            "release-12.0.0",
			Labels: []*github.Label{
				{
					Name: github.String("type/bug"),
				},
			},
			Owner:      "grafana",
			Repository: "grafana",
		})
		t.Log("Got an (expected) error from Backport:", err)

		body, _ := os.ReadFile(filepath.Join("testdata", "comment.txt"))

		require.Error(t, err)
		require.NotNil(t, comment)
		require.Equal(t, string(body), comment.GetBody())
	})
}