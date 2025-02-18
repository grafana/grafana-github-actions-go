package versions_test

import (
	"testing"

	"github.com/grafana/grafana-github-actions-go/pkg/versions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type versionBranchCase struct {
	Version       string
	VersionBranch string
}

func TestVersionBranch(t *testing.T) {
	cases := []versionBranchCase{
		{
			Version:       "8.2.1",
			VersionBranch: "v8.2.x",
		},
		{
			Version:       "v9.4.0-preview",
			VersionBranch: "v9.4.x",
		},
		{
			Version:       "v11.0.0",
			VersionBranch: "v11.0.x",
		},
		{
			Version:       "200.200.200",
			VersionBranch: "v200.200.x",
		},
		{
			Version:       "v1.2.3-preview.patch-01",
			VersionBranch: "v1.2.x",
		},
	}

	for _, v := range cases {
		res, err := versions.VersionBranch(v.Version)
		require.NoError(t, err)
		assert.Equal(t, v.VersionBranch, res)
	}
}

type releaseBranchCase struct {
	Version       string
	ReleaseBranch string
}

func TestReleaseBranch(t *testing.T) {
	cases := []releaseBranchCase{
		{
			Version:       "8.2.1",
			ReleaseBranch: "release-8.2.1",
		},
		{
			Version:       "v9.4.0-preview",
			ReleaseBranch: "release-9.4.0-preview",
		},
		{
			Version:       "v11.0.0",
			ReleaseBranch: "release-11.0.0",
		},
		{
			Version:       "200.200.200",
			ReleaseBranch: "release-200.200.200",
		},
		{
			Version:       "v1.2.3-preview.patch-01",
			ReleaseBranch: "release-1.2.3-preview.patch-01",
		},
	}

	for _, v := range cases {
		res, err := versions.ReleaseBranch(v.Version)
		require.NoError(t, err)
		assert.Equal(t, v.ReleaseBranch, res)
	}
}
