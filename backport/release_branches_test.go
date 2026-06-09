package main

import (
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackportTargetsFromLabels(t *testing.T) {
	t.Run("with backport labels", func(t *testing.T) {
		labels := []string{
			"backport v12.2.x",
			"backport v12.0.x",
			"backport v11.0.x",
		}

		targets := BackportTargetsFromLabels(labels, "backport ")
		require.Equal(t, []string{
			"v12.2.x",
			"v12.0.x",
			"v11.0.x",
		}, targets)
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

		targets := BackportTargetsFromLabels(labels, "backport ")
		require.Equal(t, []string{
			"v12.2.x",
			"v12.0.x",
			"v11.0.x",
		}, targets)
	})

	t.Run("with no labels", func(t *testing.T) {
		targets := BackportTargetsFromLabels(nil, "backport ")
		require.Empty(t, targets)
	})

	t.Run("with only non-backport labels", func(t *testing.T) {
		labels := []string{"type/bug", "release/latest", "add-to-changelog"}

		targets := BackportTargetsFromLabels(labels, "backport ")
		require.Empty(t, targets)
	})
}

func TestBackportTarget(t *testing.T) {
	assertError := func(t *testing.T, label string, branches []*github.Branch) {
		t.Helper()
		b, err := BackportTarget(label, branches)
		assert.Error(t, err)
		assert.Empty(t, b)
	}

	assertBranch := func(t *testing.T, label string, branches []*github.Branch, branch string) {
		t.Helper()
		b, err := BackportTarget(label, branches)
		assert.NoError(t, err)
		assert.Equal(t, branch, b.Name)
	}

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

	assertError(t, "v3.2.x", branches)
	assertError(t, "v4.0.x", branches)
	assertError(t, "v13.0.x", branches)
	assertError(t, "v10.5.x", branches)
	assertError(t, "v11.8.x", branches)
	assertBranch(t, "v11.0.x", branches, "release-11.0.1")
	assertBranch(t, "v12.1.x", branches, "release-12.1.15")
	assertBranch(t, "v12.0.x", branches, "release-12.0.15")
	assertBranch(t, "v1.2.x", branches, "release-1.2.3")
	assertBranch(t, "v10.2.x", branches, "release-10.2.4")
}
