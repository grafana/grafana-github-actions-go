package changelog

import (
	"testing"

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
		t.Run(test.name, func(t *testing.T) {
			issue := &ghgql.PullRequest{}
			test.issue(issue)
			output := issueAsMarkdown(*issue, nil)
			require.Equal(t, test.expectedOutput, output)
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

func TestGetOwnerAndRepo(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		testURL := "https://api.github.com/repos/grafana/grafana"
		issue := &github.Issue{
			RepositoryURL: &testURL,
		}
		owner, repo := getOwnerAndRepo(issue)
		require.Equal(t, "grafana", owner)
		require.Equal(t, "grafana", repo)
	})
}
