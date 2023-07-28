package changelog

import (
	"context"
	"testing"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghgql"
	"github.com/stretchr/testify/require"
)

func TestDeprecationNotice(t *testing.T) {
	tests := []struct {
		name           string
		issue          func(*ghgql.PullRequest)
		expectedOutput string
	}{
		{
			name: "single-line",
			issue: func(i *ghgql.PullRequest) {
				i.Number = pointerOf(123)
				i.Body = pointerOf("something else\n## Deprecation notice:\nhello.")
			},
			expectedOutput: "hello. Issue [#123](https://github.com/grafana/grafana/issues/123)",
		},
		{
			name: "blank-line-after-start",
			issue: func(i *ghgql.PullRequest) {
				i.Number = pointerOf(123)
				i.Body = pointerOf("something else\n## Deprecation notice:\n\n\nhello.")
			},
			expectedOutput: "hello. Issue [#123](https://github.com/grafana/grafana/issues/123)",
		},
		{
			name: "multi-line",
			issue: func(i *ghgql.PullRequest) {
				i.Number = pointerOf(123)
				i.Body = pointerOf("something else\n## Deprecation notice:\nhello\nworld.")
			},
			expectedOutput: "hello\nworld. Issue [#123](https://github.com/grafana/grafana/issues/123)",
		},
		{
			name: "strip-extra-empty-tail",
			issue: func(i *ghgql.PullRequest) {
				i.Number = pointerOf(123)
				i.Body = pointerOf("something else\n## Deprecation notice:\nhello\nworld.\n\n\n\n\n")
			},
			expectedOutput: "hello\nworld. Issue [#123](https://github.com/grafana/grafana/issues/123)",
		},
		// If the notice ends with a codeblock, then we have to introduce an extra newline:
		{
			name: "codeblock-end",
			issue: func(i *ghgql.PullRequest) {
				i.Number = pointerOf(123)
				i.Body = pointerOf("something else\n## Deprecation notice:\n```\nhello\n```")
			},
			expectedOutput: "```\nhello\n```\nIssue [#123](https://github.com/grafana/grafana/issues/123)",
		},
		// If the notice ends with a codeblock and also with a newline (or
		// more) then only a single newline should remain:
		{
			name: "codeblock-end-plus-newline",
			issue: func(i *ghgql.PullRequest) {
				i.Number = pointerOf(123)
				i.Body = pointerOf("something else\n## Deprecation notice:\n```\nhello\n```\n\n\n")
			},
			expectedOutput: "```\nhello\n```\nIssue [#123](https://github.com/grafana/grafana/issues/123)",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issue := &ghgql.PullRequest{}
			test.issue(issue)
			output := getDeprecationNotice(*issue)
			require.Equal(t, test.expectedOutput, output)
		})
	}
}

func TestIssueLine(t *testing.T) {
	tests := []struct {
		name           string
		issue          func(*ghgql.PullRequest)
		expectedOutput string
	}{
		{
			name: "enterprise",
			issue: func(i *ghgql.PullRequest) {
				i.Title = pointerOf("hello")
				addLabel(t, i, "enterprise")
			},
			expectedOutput: "- hello. (Enterprise)\n",
		},
		{
			name: "version-prefix",
			issue: func(i *ghgql.PullRequest) {
				i.Title = pointerOf("[v9.5.x] Chore: hello")
				i.Number = pointerOf(123)
			},
			expectedOutput: "- **Chore:** hello. [#123](https://github.com/grafana/grafana/issues/123)\n",
		},
		{
			name: "version-prefix-doubledigit",
			issue: func(i *ghgql.PullRequest) {
				i.Title = pointerOf("[v10.0.x] Chore: hello")
				i.Number = pointerOf(123)
			},
			expectedOutput: "- **Chore:** hello. [#123](https://github.com/grafana/grafana/issues/123)\n",
		},
		{
			name: "version-prefix-plus-colon",
			issue: func(i *ghgql.PullRequest) {
				i.Title = pointerOf("[v10.0.x]: Chore: hello")
				i.Number = pointerOf(123)
			},
			expectedOutput: "- **Chore:** hello. [#123](https://github.com/grafana/grafana/issues/123)\n",
		},
		{
			name: "enterprise-with-category",
			issue: func(i *ghgql.PullRequest) {
				i.Title = pointerOf("hello: world")
				addLabel(t, i, "enterprise")
			},
			expectedOutput: "- **hello:** world. (Enterprise)\n",
		},
		{
			name: "oss-issue",
			issue: func(i *ghgql.PullRequest) {
				i.Title = pointerOf("hello")
				i.Number = pointerOf(123)
			},
			expectedOutput: "- hello. [#123](https://github.com/grafana/grafana/issues/123)\n",
		},
		{
			name: "oss-pull-request",
			issue: func(i *ghgql.PullRequest) {
				i.Title = pointerOf("hello")
				i.Number = pointerOf(123)
				i.AuthorLogin = pointerOf("author")
			},
			expectedOutput: "- hello. [#123](https://github.com/grafana/grafana/issues/123), [@author](https://github.com/author)\n",
		},
	}

	for _, test := range tests {
		r := defaultRenderer{}
		t.Run(test.name, func(t *testing.T) {
			issue := &ghgql.PullRequest{}
			test.issue(issue)
			output := r.issueAsMarkdown(*issue)
			require.Equal(t, test.expectedOutput, output)
		})
	}
}

func TestDeduplicateEntries(t *testing.T) {
	tests := []struct {
		name                string
		currentPullRequests []ghgql.PullRequest
		previousChangelogs  map[string]string
		expectError         bool
		expectResult        []ghgql.PullRequest
	}{
		{
			name:                "empty",
			expectError:         false,
			expectResult:        []ghgql.PullRequest{},
			currentPullRequests: []ghgql.PullRequest{},
			previousChangelogs: map[string]string{
				"10.0.0": "### Bug fixes",
			},
		},
		{
			name:        "no-entries-to-remove",
			expectError: false,
			expectResult: []ghgql.PullRequest{
				{
					Number: pointerOf(1),
					Title:  pointerOf("Category: Title 1"),
				},
			},
			currentPullRequests: []ghgql.PullRequest{
				{
					Number: pointerOf(1),
					Title:  pointerOf("Category: Title 1"),
				},
			},
			previousChangelogs: map[string]string{
				"10.0.0": "### Bug fixes",
			},
		},
		{
			name:        "one-matching-entry-to-be-removed",
			expectError: false,
			expectResult: []ghgql.PullRequest{
				{
					Number: pointerOf(1),
					Title:  pointerOf("Category: Title 1"),
				},
			},
			currentPullRequests: []ghgql.PullRequest{
				{
					Number: pointerOf(1),
					Title:  pointerOf("Category: Title 1"),
				},
				{
					Number: pointerOf(2),
					Title:  pointerOf("Category: Title 2"),
				},
			},
			previousChangelogs: map[string]string{
				"10.0.0": "### Bug fixes\n\n- **Category:** Title 2.\n",
			},
		},
		{
			name:        "one-matching-enterprise-entry-to-be-removed",
			expectError: false,
			expectResult: []ghgql.PullRequest{
				{
					Number: pointerOf(1),
					Title:  pointerOf("Category: Title 1"),
				},
			},
			currentPullRequests: []ghgql.PullRequest{
				{
					Number: pointerOf(1),
					Title:  pointerOf("Category: Title 1"),
				},
				{
					Number: pointerOf(2),
					Title:  pointerOf("Category: Title 2"),
					Labels: []string{"enterprise"},
				},
			},
			previousChangelogs: map[string]string{
				"10.0.0": "### Bug fixes\n\n- **Category:** Title 2. (Enterprise)\n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			output, err := deduplicateEntries(ctx, test.currentPullRequests, test.previousChangelogs)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectResult, output)
			}
		})
	}
}

func TestFilterMilestonesForDeduplication(t *testing.T) {
	allMilestones := []*github.Milestone{
		{
			Title: pointerOf("10.1.0"),
			DueOn: newTS(t, "2023-01-01T12:00:00+00:00"),
		},
		{
			Title: pointerOf("10.0.3"),
			DueOn: newTS(t, "2022-12-31T12:00:00+00:00"),
		},
		{
			Title: pointerOf("10.0.2"),
			DueOn: newTS(t, "2022-12-03T12:00:00+00:00"),
		},
		{
			Title: pointerOf("10.0.1"),
			DueOn: newTS(t, "2022-12-02T12:00:00+00:00"),
		},
		{
			Title: pointerOf("10.0.0"),
			DueOn: newTS(t, "2022-12-01T12:00:00+00:00"),
		},
		{
			Title: pointerOf("9.5.3"),
			DueOn: newTS(t, "2022-11-04T12:00:00+00:00"),
		},
		{
			Title: pointerOf("9.5.2"),
			DueOn: newTS(t, "2022-11-03T12:00:00+00:00"),
		},
		{
			Title: pointerOf("9.5.1"),
			DueOn: newTS(t, "2022-11-02T12:00:00+00:00"),
		},
		{
			Title: pointerOf("9.5.0"),
			DueOn: newTS(t, "2022-11-01T12:00:00+00:00"),
		},
		{
			Title: pointerOf("10.1.x"),
		},
		{
			Title: pointerOf("10.0.x"),
		},
		{
			Title: pointerOf("9.5.x"),
		},
	}
	tests := []struct {
		name             string
		currentMilestone string
		expectOutput     []string
		expectError      bool
	}{
		{
			name:             "no-previous-milestone",
			currentMilestone: "9.5.3",
			expectOutput:     []string{},
			expectError:      false,
		},
		{
			name:             "all-previous-older-than-one-day",
			currentMilestone: "10.1.0",
			// The 10.0.x releases except for 10.0.3 happened more than a day
			// before 10.1.0 and so they should be included:
			expectOutput: []string{"10.0.2", "10.0.1", "10.0.0"},
			expectError:  false,
		},
		{
			name:             "all-previous",
			currentMilestone: "10.0.3",
			expectOutput:     []string{"9.5.3", "9.5.2", "9.5.1", "9.5.0"},
			expectError:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			version, err := semver.NewVersion(test.currentMilestone)
			require.NoError(t, err)
			var currentMilestone *github.Milestone
			for _, ms := range allMilestones {
				if ms.GetTitle() == test.currentMilestone {
					currentMilestone = ms
					break
				}
			}
			output, err := filterMilestonesForDeduplication(ctx, allMilestones, currentMilestone, *version, time.Hour*24)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if err != nil {
				}
				outputTitles := make([]string, 0, len(output))
				for _, o := range output {
					outputTitles = append(outputTitles, o.GetTitle())
				}
				require.Equal(t, test.expectOutput, outputTitles)
			}
		})
	}
}

func addLabel(t *testing.T, issue *ghgql.PullRequest, labelName string) {
	if issue.Labels == nil {
		issue.Labels = make([]string, 0, 5)
	}
	issue.Labels = append(issue.Labels, labelName)
}

func pointerOf[T any](value T) *T {
	return &value
}

func newTS(t *testing.T, s string) *github.Timestamp {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return &github.Timestamp{Time: ts}

}
