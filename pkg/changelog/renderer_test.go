package changelog

import (
	"testing"

	"github.com/grafana/grafana-github-actions-go/pkg/ghgql"
	"github.com/stretchr/testify/require"
)

func TestIsBot(t *testing.T) {
	t.Run("apps are bots", func(t *testing.T) {
		issue := ghgql.PullRequest{}
		issue.AuthorResourcePath = pointerOf("/apps/hello")
		require.True(t, isBotUser(issue))
	})

	t.Run("grafanabot is bot", func(t *testing.T) {
		issue := ghgql.PullRequest{}
		issue.AuthorLogin = pointerOf("grafanabot")
		require.True(t, isBotUser(issue))
	})

	t.Run("user is not bot", func(t *testing.T) {
		issue := ghgql.PullRequest{}
		issue.AuthorLogin = pointerOf("zerok")
		require.False(t, isBotUser(issue))
	})
}
