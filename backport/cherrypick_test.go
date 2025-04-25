package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grafana/grafana-github-actions-go/pkg/ghutil"
	"github.com/stretchr/testify/require"
)

func TestCreateCherryPickBranch(t *testing.T) {
	t.Run("It should handle betterer conflicts", func(t *testing.T) {
		var (
			testCommitDate, _ = time.Parse(time.RFC3339, "2020-01-02T00:00:00Z")
			branch            = "example"
			opts              = BackportOpts{
				Target: ghutil.Branch{
					Name: "release-1.0.0",
					SHA:  "fdsa4321",
				},
				SourceSHA:        "asdf1234",
				SourceCommitDate: testCommitDate,
			}
			runner = NewErrorRunner(map[string]error{
				"git cherry-pick -x asdf1234":               errors.New("cherry-pick error"),
				"git diff -s --exit-code .betterer.results": errors.New("command returned 1"),
			})
		)

		expect := []string{
			"git fetch --shallow-since=\"2020-01-01\"",
			"git fetch origin release-1.0.0:refs/remotes/origin/release-1.0.0",
			"git checkout -b example --track origin/release-1.0.0",
			"git cherry-pick -x asdf1234",
			"git diff -s --exit-code .betterer.results",
			"yarn run betterer",
			"git add .betterer.results",
			"git -c core.editor=true cherry-pick --continue",
		}

		require.NoError(t, CreateCherryPickBranch(context.Background(), runner, branch, opts))

		require.Equal(t, expect, runner.History.Commands)
	})

	t.Run("It should return an error if there was a non-betterer conflict", func(t *testing.T) {
		var (
			testCommitDate, _ = time.Parse(time.RFC3339, "2020-01-02T00:00:00Z")
			branch            = "example"
			opts              = BackportOpts{
				Target: ghutil.Branch{
					Name: "release-1.0.0",
					SHA:  "fdsa4321",
				},
				SourceSHA:        "asdf1234",
				SourceCommitDate: testCommitDate,
			}
			runner = NewErrorRunner(map[string]error{
				"git cherry-pick -x asdf1234": errors.New("cherry-pick error"),
			})
		)

		expect := []string{
			"git fetch --shallow-since=\"2020-01-01\"",
			"git fetch origin release-1.0.0:refs/remotes/origin/release-1.0.0",
			"git checkout -b example --track origin/release-1.0.0",
			"git cherry-pick -x asdf1234",
			"git diff -s --exit-code .betterer.results",
			"git cherry-pick --abort",
		}

		require.Error(t, CreateCherryPickBranch(context.Background(), runner, branch, opts))
		require.Equal(t, expect, runner.History.Commands)
	})
}
