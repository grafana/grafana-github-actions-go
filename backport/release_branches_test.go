package main

import (
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghutil"
	"github.com/stretchr/testify/require"
)

func TestBackportTargets(t *testing.T) {
	branches := []*github.Branch{
		{Name: github.String("release-11.0.1")},
		{Name: github.String("release-1.2.3")},
		{Name: github.String("release-11.0.1+security-01")},
		{Name: github.String("release-10.0.0")},
		{Name: github.String("release-10.2.3")},
		{Name: github.String("release-10.2.4")},
		{Name: github.String("release-10.2.4+security-01")},
		{Name: github.String("release-12.0.3")},
		{Name: github.String("release-12.1.3")},
		{Name: github.String("release-12.0.15")},
		{Name: github.String("release-12.1.15")},
		{Name: github.String("release-12.2.12")},
	}

	t.Run("with backport labels", func(t *testing.T) {
		labels := []string{
			"backport v12.2.x",
			"backport v12.0.x",
			"backport v11.0.x",
		}

		targets, err := BackportTargets(branches, labels)
		require.NoError(t, err)
		require.Equal(t, []string{
			"release-12.2.12",
			"release-12.0.15",
			"release-11.0.1",
		}, toStringList(targets))
	})

	t.Run("with non-backport labels", func(t *testing.T) {
		labels := []string{
			"type/bug",
			"backport v12.2.x",
			"release/latest",
			"backport v12.0.x",
			"type/ci",
			"backport v11.0.x",
			"add-to-changelog",
		}

		targets, err := BackportTargets(branches, labels)
		require.NoError(t, err)
		require.Equal(t, []string{
			"release-12.2.12",
			"release-12.0.15",
			"release-11.0.1",
		}, toStringList(targets))
	})

}

func toStringList(branches []ghutil.Branch) []string {
	r := make([]string, len(branches))

	for i, v := range branches {
		r[i] = v.Name
	}

	return r
}
