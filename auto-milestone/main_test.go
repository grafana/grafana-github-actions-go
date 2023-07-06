package main

import (
	"context"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghgql"
	"github.com/stretchr/testify/require"
)

func TestVersionExtraction(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectErr    bool
		expectOutput string
	}{
		{
			name:         "valid",
			input:        `{"version": "10.1.0-pre"}`,
			expectErr:    false,
			expectOutput: "10.1.x",
		},
		{
			name:         "no-field",
			input:        `{}`,
			expectErr:    true,
			expectOutput: "",
		},
		{
			name:         "invalid-version",
			input:        `{"version": "hello"}`,
			expectErr:    true,
			expectOutput: "",
		},
		{
			name:         "not-json",
			input:        `hello`,
			expectErr:    true,
			expectOutput: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := versionFromPackage(test.input)
			if err != nil && !test.expectErr {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if err == nil && test.expectErr {
				t.Fatalf("missing expected error")
			}
			require.Equal(t, test.expectOutput, v)
		})
	}
}

func TestDetermineAction(t *testing.T) {
	v10xTitle := "10.0.x"
	v10Title := "10.0.0"
	tests := []struct {
		name             string
		currentMilestone *github.Milestone
		targetMilestone  *ghgql.Milestone
		expected         action
	}{
		{
			name:             "no-milestone-set",
			currentMilestone: nil,
			targetMilestone:  &ghgql.Milestone{Title: "10.0.x"},
			expected: action{
				Type:      actionTypeSetToMilestone,
				Milestone: &ghgql.Milestone{Title: "10.0.x"},
			},
		},
		{
			name: "milestone-correct",
			currentMilestone: &github.Milestone{
				Title: &v10xTitle,
			},
			targetMilestone: &ghgql.Milestone{Title: "10.0.x"},
			expected: action{
				Type: actionTypeNoop,
			},
		},
		{
			name: "milestone-incorrect",
			currentMilestone: &github.Milestone{
				Title: &v10xTitle,
			},
			targetMilestone: &ghgql.Milestone{Title: "10.1.x"},
			expected: action{
				Type:      actionTypeSetToMilestone,
				Milestone: &ghgql.Milestone{Title: "10.1.x"},
			},
		},
		{
			name: "release-milestone-set",
			currentMilestone: &github.Milestone{
				Title: &v10Title,
			},
			targetMilestone: &ghgql.Milestone{Title: "10.0.x"},
			expected: action{
				Type: actionTypeNoop,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			output := determineAction(ctx, test.currentMilestone, test.targetMilestone)
			require.Equal(t, test.expected, output)
		})
	}
}
