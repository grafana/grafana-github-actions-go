package main

import (
	"context"
	"errors"
	"testing"

	gh "github.com/google/go-github/v47/github"
	"github.com/stretchr/testify/require"
)

func TestGetOrCreateRelease(t *testing.T) {
	t.Run("If a release is found, the function should return it", func(t *testing.T) {
		client := &testReleaseClient{
			shouldGetRelease: true,
		}
		r, created, err := getOrCreateRelease(context.Background(), client, "grafana", "repo name", "v1.0.0", &gh.RepositoryRelease{})

		require.NotNil(t, r)
		require.False(t, created)
		require.NoError(t, err)
	})
	t.Run("If a release is not found, create one", func(t *testing.T) {
		client := &testReleaseClient{
			shouldCreateRelease: true,
		}
		r, created, err := getOrCreateRelease(context.Background(), client, "grafana", "repo name", "v1.0.0", &gh.RepositoryRelease{})

		require.NotNil(t, r)
		require.True(t, created)
		require.NoError(t, err)
	})
	t.Run("If a release is not found, and one could not be created, return an error", func(t *testing.T) {
		client := &testReleaseClient{}
		r, created, err := getOrCreateRelease(context.Background(), client, "grafana", "repo name", "v1.0.0", &gh.RepositoryRelease{})

		require.Nil(t, r)
		require.False(t, created)
		require.Error(t, err)
	})
}

func TestUpsertRelease(t *testing.T) {
	t.Run("If a release is updated, the function should return it", func(t *testing.T) {
		client := &testReleaseClient{
			shouldEditRelease: true,
		}
		r, err := upsertRelease(context.Background(), client, "grafana", "repo name", "v1.0.0", &gh.RepositoryRelease{})

		require.NotNil(t, r)
		require.NoError(t, err)
	})
	t.Run("If a release is not updated, create a new one", func(t *testing.T) {
		client := &testReleaseClient{
			shouldCreateRelease: true,
		}
		r, err := upsertRelease(context.Background(), client, "grafana", "repo name", "v1.0.0", &gh.RepositoryRelease{})

		require.NotNil(t, r)
		require.NoError(t, err)
	})
	t.Run("If a release is not updated, and one could not be created, return an error", func(t *testing.T) {
		client := &testReleaseClient{}
		r, err := upsertRelease(context.Background(), client, "grafana", "repo name", "v1.0.0", &gh.RepositoryRelease{})

		require.Nil(t, r)
		require.Error(t, err)
	})
}

type testReleaseClient struct {
	shouldCreateRelease bool
	shouldGetRelease    bool
	shouldEditRelease   bool
}

func (c *testReleaseClient) CreateRelease(ctx context.Context, owner string, repo string, release *gh.RepositoryRelease) (*gh.RepositoryRelease, *gh.Response, error) {
	if c.shouldCreateRelease {
		return &gh.RepositoryRelease{}, nil, nil
	}
	return nil, nil, errors.New("failed to create release")
}

func (c *testReleaseClient) GetReleaseByTag(ctx context.Context, owner string, repo string, tag string) (*gh.RepositoryRelease, *gh.Response, error) {
	if c.shouldGetRelease {
		return &gh.RepositoryRelease{}, nil, nil
	}
	return nil, nil, errors.New("not found")
}

func (c *testReleaseClient) EditRelease(ctx context.Context, owner string, repo string, id int64, release *gh.RepositoryRelease) (*gh.RepositoryRelease, *gh.Response, error) {
	if c.shouldEditRelease || c.shouldCreateRelease {
		return &gh.RepositoryRelease{}, nil, nil
	}
	return nil, nil, errors.New("not updated")
}
