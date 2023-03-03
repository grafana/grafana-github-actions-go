package changelog

import (
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/require"
)

func TestDeprecationNotice(t *testing.T) {
	tests := []struct {
		name           string
		issue          func(*github.Issue)
		expectedOutput string
	}{
		{
			name: "single-line",
			issue: func(i *github.Issue) {
				i.Number = pointerOf(123)
				i.Body = pointerOf("something else\n## Deprecation notice:\nhello.")
			},
			expectedOutput: "hello. Issue [#123](https://github.com/grafana/grafana/issues/123)",
		},
		{
			name: "blank-line-after-start",
			issue: func(i *github.Issue) {
				i.Number = pointerOf(123)
				i.Body = pointerOf("something else\n## Deprecation notice:\n\n\nhello.")
			},
			expectedOutput: "hello. Issue [#123](https://github.com/grafana/grafana/issues/123)",
		},
		{
			name: "multi-line",
			issue: func(i *github.Issue) {
				i.Number = pointerOf(123)
				i.Body = pointerOf("something else\n## Deprecation notice:\nhello\nworld.")
			},
			expectedOutput: "hello\nworld. Issue [#123](https://github.com/grafana/grafana/issues/123)",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issue := &github.Issue{}
			test.issue(issue)
			output := getDeprecationNotice(issue)
			require.Equal(t, test.expectedOutput, output)
		})
	}
}

func TestIssueLine(t *testing.T) {
	tests := []struct {
		name           string
		issue          func(*github.Issue)
		expectedOutput string
	}{
		{
			name: "enterprise",
			issue: func(i *github.Issue) {
				i.Title = pointerOf("hello")
				addLabel(t, i, "enterprise")
			},
			expectedOutput: "- hello. (Enterprise)\n",
		},
		{
			name: "enterprise-with-category",
			issue: func(i *github.Issue) {
				i.Title = pointerOf("hello: world")
				addLabel(t, i, "enterprise")
			},
			expectedOutput: "- **hello:** world. (Enterprise)\n",
		},
		{
			name: "oss-issue",
			issue: func(i *github.Issue) {
				i.Title = pointerOf("hello")
				i.Number = pointerOf(123)
			},
			expectedOutput: "- hello. [#123](https://github.com/grafana/grafana/issues/123)\n",
		},
		{
			name: "oss-pull-request",
			issue: func(i *github.Issue) {
				i.Title = pointerOf("hello")
				i.Number = pointerOf(123)
				i.User = &github.User{
					Login: pointerOf("author"),
				}
				i.PullRequestLinks = &github.PullRequestLinks{}
			},
			expectedOutput: "- hello. [#123](https://github.com/grafana/grafana/issues/123), [@author](https://github.com/author)\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issue := &github.Issue{}
			test.issue(issue)
			output := issueAsMarkdown(issue)
			require.Equal(t, test.expectedOutput, output)
		})
	}
}

func addLabel(t *testing.T, issue *github.Issue, labelName string) {
	if issue.Labels == nil {
		issue.Labels = make([]*github.Label, 0, 5)
	}
	label := &github.Label{}
	label.Name = pointerOf(labelName)
	issue.Labels = append(issue.Labels, label)
}

func pointerOf[T any](value T) *T {
	return &value
}
