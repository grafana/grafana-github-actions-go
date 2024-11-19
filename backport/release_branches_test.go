package main

import (
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/require"
)

func TestBackportTargets(t *testing.T) {
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

	t.Run("with backport labels", func(t *testing.T) {
		labels := []*github.Label{
			&github.Label{
				Name: github.String("backport v12.2.x"),
			},
			&github.Label{
				Name: github.String("backport v12.0.x"),
			},
			&github.Label{
				Name: github.String("backport v11.0.x"),
			},
		}

		targets, err := BackportTargets(branches, labels)
		require.NoError(t, err)
		require.Equal(t, []string{
			"release-12.2.12",
			"release-12.0.15",
			"release-11.0.1",
		}, targets)
	})

	t.Run("with non-backport labels", func(t *testing.T) {
		labels := []*github.Label{
			&github.Label{
				Name: github.String("type/bug"),
			},
			&github.Label{
				Name: github.String("backport v12.2.x"),
			},
			&github.Label{
				Name: github.String("release/latest"),
			},
			&github.Label{
				Name: github.String("backport v12.0.x"),
			},
			&github.Label{
				Name: github.String("type/ci"),
			},
			&github.Label{
				Name: github.String("backport v11.0.x"),
			},
			&github.Label{
				Name: github.String("add-to-changelog"),
			},
		}

		targets, err := BackportTargets(branches, labels)
		require.NoError(t, err)
		require.Equal(t, []string{
			"release-12.2.12",
			"release-12.0.15",
			"release-11.0.1",
		}, targets)
	})

}
