package changelog

import (
	"context"
	"testing"

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

func addLabel(t *testing.T, issue *ghgql.PullRequest, labelName string) {
	if issue.Labels == nil {
		issue.Labels = make([]string, 0, 5)
	}
	issue.Labels = append(issue.Labels, labelName)
}

func pointerOf[T any](value T) *T {
	return &value
}
