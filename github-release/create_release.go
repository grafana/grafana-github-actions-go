package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v50/github"
)

type ReleaseCreator interface {
	ListTags(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error)
	CreateRelease(ctx context.Context, owner, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error)
	EditRelease(ctx context.Context, owner, repo string, id int64, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error)
}

func CreateRelease(ctx context.Context, client ReleaseCreator, owner, repo, tag string, release *github.RepositoryRelease) (string, error) {
	tagExists, err := verifyTagExists(ctx, client, owner, repo, tag)
	if err != nil {
		return "", fmt.Errorf("failed to verify tag exists: %w", err)
	}

	if !tagExists {
		return "", fmt.Errorf("tag `%s` does not exist", tag)
	}

	rel, resp, err := client.GetReleaseByTag(ctx, owner, repo, tag)
	if err != nil {
		if resp.StatusCode != http.StatusNotFound {
			return "", fmt.Errorf("failed to check for existing release: %w", err)
		}
	}

	if rel != nil {
		rel.Name = release.Name
		rel.Body = release.Body
		rel.MakeLatest = release.MakeLatest

		r, _, err := client.EditRelease(ctx, owner, repo, rel.GetID(), rel)
		if err != nil {
			return "", fmt.Errorf("failed to update existing release: %w", err)
		}

		return r.GetHTMLURL(), nil
	}

	rel, _, err = client.CreateRelease(ctx, owner, repo, release)
	if err != nil {
		return "", err
	}

	return "", nil
}

func verifyTagExists(ctx context.Context, client ReleaseCreator, owner string, repo string, tag string) (bool, error) {
	opts := github.ListOptions{}
	opts.Page = 1
	for {
		tags, resp, err := client.ListTags(ctx, owner, repo, &opts)
		if err != nil {
			return false, err
		}
		for _, t := range tags {
			if t.GetName() == tag {
				return true, nil
			}
		}
		if resp.NextPage <= opts.Page {
			break
		}
		opts.Page = resp.NextPage
	}
	return false, nil
}
