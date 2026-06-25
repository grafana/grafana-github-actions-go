package ghutil

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/go-github/v88/github"
	"github.com/stretchr/testify/require"
)

func TestGetReleaseBranches(t *testing.T) {
	branch := func(name string, protected bool) *github.Branch {
		return &github.Branch{Name: github.String(name), Protected: github.Bool(protected)}
	}

	client := &fakeBranchClient{
		pages: [][]*github.Branch{
			{branch("release-12.1.4", true), branch("some-feature", false)},
			{branch("release-12.2.0", true), branch("release-12.1.5", false)},
		},
	}

	branches, err := GetReleaseBranches(t.Context(), slog.New(slog.NewTextHandler(io.Discard, nil)), client, "grafana", "grafana")
	require.NoError(t, err)

	// Only keep protected branches
	names := make([]string, len(branches))
	for i, b := range branches {
		names[i] = b.GetName()
	}
	require.Equal(t, []string{"release-12.1.4", "release-12.2.0"}, names)
}

type fakeBranchClient struct{ pages [][]*github.Branch }

func (f *fakeBranchClient) ListBranches(_ context.Context, _, _ string, opts *github.BranchListOptions) ([]*github.Branch, *github.Response, error) {
	page := opts.Page
	if page == 0 {
		page = 1
	}

	resp := &github.Response{}
	if page < len(f.pages) {
		resp.NextPage = page + 1
	}

	return f.pages[page-1], resp, nil
}

func (f *fakeBranchClient) GetBranch(_ context.Context, _, _, _ string, _ bool) (*github.Branch, *github.Response, error) {
	return nil, nil, nil
}
