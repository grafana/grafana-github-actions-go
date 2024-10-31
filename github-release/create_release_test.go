package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-github/v50/github"
)

type TestReleaseCreator struct {
	ListTagsFunc        func(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error)
	CreateReleaseFunc   func(ctx context.Context, owner, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error)
	GetReleaseByTagFunc func(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error)
	EditReleaseFunc     func(ctx context.Context, owner, repo string, id int64, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error)
}

func (c *TestReleaseCreator) ListTags(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
	return c.ListTagsFunc(ctx, owner, repo, opts)
}

func (c *TestReleaseCreator) CreateRelease(ctx context.Context, owner, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error) {
	return c.CreateReleaseFunc(ctx, owner, repo, release)
}

func (c *TestReleaseCreator) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error) {
	return c.GetReleaseByTagFunc(ctx, owner, repo, tag)
}

func (c *TestReleaseCreator) EditRelease(ctx context.Context, owner, repo string, id int64, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error) {
	return c.EditReleaseFunc(ctx, owner, repo, id, release)
}

func TestCreateRelease(t *testing.T) {
	ctx := context.Background()

	listTagsFunc := func(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
		return []*github.RepositoryTag{
				{
					Name: github.String("v1.2.3"),
				},
				{
					Name: github.String("v1.2.3+security-01"),
				},
				{
					Name: github.String("v1.2.4"),
				},
			}, &github.Response{
				NextPage: 0,
			}, nil
	}

	createReleaseFunc := func(ctx context.Context, owner, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error) {
		return release, nil, nil
	}

	getReleaseByTagFunc := func(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error) {
		return &github.RepositoryRelease{
			TagName: github.String(tag),
		}, nil, nil
	}

	editReleaseFunc := func(ctx context.Context, owner, repo string, id int64, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error) {
		return &github.RepositoryRelease{
			ID:      github.Int64(id),
			TagName: release.TagName,
			Name:    release.Name,
			Body:    release.Body,
		}, nil, nil
	}

	t.Run("It should return an error if the tag doesn't exist", func(t *testing.T) {
		client := &TestReleaseCreator{
			ListTagsFunc:        listTagsFunc,
			CreateReleaseFunc:   createReleaseFunc,
			GetReleaseByTagFunc: getReleaseByTagFunc,
			EditReleaseFunc:     editReleaseFunc,
		}

		_, err := CreateRelease(ctx, client, "grafana", "grafana", "v3.2.1", &github.RepositoryRelease{})
		if err == nil {
			t.Fatal("CreateRelease should return an error if the tag does not exist in the repo")
		}
	})

	t.Run("It should create a github release if one doesn't exist", func(t *testing.T) {
		created := false
		edited := false
		client := &TestReleaseCreator{
			ListTagsFunc: listTagsFunc,
			CreateReleaseFunc: func(ctx context.Context, owner, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error) {
				created = true
				return release, nil, nil
			},
			GetReleaseByTagFunc: func(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error) {
				return nil, &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
					},
				}, nil
			},
			EditReleaseFunc: func(ctx context.Context, owner, repo string, id int64, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error) {
				edited = true
				return nil, nil, nil
			},
		}

		_, err := CreateRelease(ctx, client, "grafana", "grafana", "v1.2.3", &github.RepositoryRelease{
			MakeLatest: LatestString(true),
		})
		if err != nil {
			t.Fatal("CreateRelease should not return an error; error: ", err)
		}

		if !created {
			t.Fatal("CreateRelease not called")
		}
		if edited {
			t.Fatal("EditRelease should not be called")
		}
	})

	t.Run("It should edit a github release if one does exist", func(t *testing.T) {
		created := false
		edited := false
		client := &TestReleaseCreator{
			ListTagsFunc: listTagsFunc,
			CreateReleaseFunc: func(ctx context.Context, owner, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error) {
				created = true
				return release, nil, nil
			},
			GetReleaseByTagFunc: func(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error) {
				return &github.RepositoryRelease{
						Name:    github.String(tag),
						TagName: github.String(tag),
					}, &github.Response{
						Response: &http.Response{
							StatusCode: http.StatusOK,
						},
					}, nil
			},
			EditReleaseFunc: func(ctx context.Context, owner, repo string, id int64, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error) {
				edited = true
				return release, nil, nil
			},
		}

		_, err := CreateRelease(ctx, client, "grafana", "grafana", "v1.2.3", &github.RepositoryRelease{
			MakeLatest: LatestString(true),
		})
		if err != nil {
			t.Fatal("CreateRelease should not return an error; error: ", err)
		}

		if created {
			t.Fatal("CreateRelease should not be called")
		}

		if !edited {
			t.Fatal("EditRelease was not called")
		}
	})
}
