package main

import (
	"testing"

	"github.com/google/go-github/v88/github"
	"github.com/stretchr/testify/require"
)

// Regression coverage for BackportTarget across label + branch naming conventions.
// Context: grafana/tempo#7055 — three "backport release-vX.Y" labels all
// collapsed to the last release branch because MajorMinorPatch could not parse
// 2-component versions or labels carrying a "release-" prefix.
func TestBackportTarget_Formats(t *testing.T) {
	tempoBranches := []string{
		"release-v2.8", "release-v2.9", "release-v2.10",
	}
	grafanaBranches := []string{
		"release-v11.0.0", "release-v11.2.0", "release-v11.2.1",
		"release-8.2.0", "release-8.2.1", "release-8.2.2",
	}

	cases := []struct {
		name     string
		branches []string
		label    string
		want     string
	}{
		// Tempo: "release-"-prefixed label against 2-component branches.
		{"tempo release-v2.8", tempoBranches, "release-v2.8", "release-v2.8"},
		{"tempo release-v2.9", tempoBranches, "release-v2.9", "release-v2.9"},
		{"tempo release-v2.10", tempoBranches, "release-v2.10", "release-v2.10"},

		// Grafana: short-form label against 3-component branches picks newest patch.
		{"grafana v11.2.x", grafanaBranches, "v11.2.x", "release-v11.2.1"},

		// Grafana: 3-component branches without the "v" prefix also resolve.
		{"grafana v8.2.x (no-v branch)", grafanaBranches, "v8.2.x", "release-8.2.2"},
		{"grafana 8.2.x (no-v branch)", grafanaBranches, "8.2.x", "release-8.2.2"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			br := make([]*github.Branch, len(tc.branches))
			for i, n := range tc.branches {
				br[i] = &github.Branch{Name: github.String(n)}
			}
			target, err := BackportTarget(tc.label, br)
			require.NoError(t, err)
			require.Equal(t, tc.want, target.Name)
		})
	}
}
