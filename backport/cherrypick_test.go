package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-github/v50/github"
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
				MergeBase: &github.Commit{
					Committer: &github.CommitAuthor{
						Date: &github.Timestamp{
							Time: testCommitDate,
						},
					},
				},
			}
			runner = newErrorRunner(map[string]error{
				"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234": errors.New("cherry-pick error"),
				"git diff -s --exit-code .betterer.results":                                         errors.New("command returned 1"),
			})
		)

		expect := []string{
			"git fetch origin asdf1234",
			"git fetch origin release-1.0.0:refs/remotes/origin/release-1.0.0",
			"git fetch --shallow-since=1577923200",
			"git checkout -b example origin/release-1.0.0",
			"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234",
			"git ls-files --error-unmatch .betterer.results",
			"git diff -s --exit-code .betterer.results",
			"yarn run betterer",
			"git add .betterer.results",
			"git -c core.editor=true cherry-pick --continue",
		}

		require.NoError(t, CreateCherryPickBranch(context.Background(), runner, branch, opts))

		require.Equal(t, expect, runner.History.Commands)
	})

	t.Run("It should configure and restore origin URL when GitToken is set", func(t *testing.T) {
		var (
			testCommitDate, _ = time.Parse(time.RFC3339, "2020-01-02T00:00:00Z")
			branch            = "example"
			opts              = BackportOpts{
				GitToken: "test-token",
				Target: ghutil.Branch{
					Name: "release-1.0.0",
					SHA:  "fdsa4321",
				},
				SourceSHA:        "asdf1234",
				SourceCommitDate: testCommitDate,
				MergeBase: &github.Commit{
					Committer: &github.CommitAuthor{
						Date: &github.Timestamp{
							Time: testCommitDate,
						},
					},
				},
			}
			runner = &mockRunner{
				Commands: []string{},
				Outputs: map[string]string{
					"git remote get-url origin": "https://github.com/test-owner/test-repo.git",
				},
			}
		)

		expect := []string{
			"git remote get-url origin",
			"git remote set-url origin https://x-access-token:test-token@github.com/test-owner/test-repo.git", // trufflehog:ignore
			"git fetch origin asdf1234",
			"git fetch origin release-1.0.0:refs/remotes/origin/release-1.0.0",
			"git fetch --shallow-since=1577923200",
			"git checkout -b example origin/release-1.0.0",
			"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234",
			"git remote set-url origin https://github.com/test-owner/test-repo.git",
		}

		require.NoError(t, CreateCherryPickBranch(context.Background(), runner, branch, opts))
		require.Equal(t, expect, runner.Commands)
	})

	t.Run("It should restore origin URL even if cherry-pick fails", func(t *testing.T) {
		var (
			testCommitDate, _ = time.Parse(time.RFC3339, "2020-01-02T00:00:00Z")
			branch            = "example"
			opts              = BackportOpts{
				GitToken: "test-token",
				Target: ghutil.Branch{
					Name: "release-1.0.0",
					SHA:  "fdsa4321",
				},
				SourceSHA:        "asdf1234",
				SourceCommitDate: testCommitDate,
				MergeBase: &github.Commit{
					Committer: &github.CommitAuthor{
						Date: &github.Timestamp{
							Time: testCommitDate,
						},
					},
				},
			}
			runner = newErrorRunner(map[string]error{
				"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234": errors.New("cherry-pick error"),
			})
		)
		runner.History.Outputs = map[string]string{
			"git remote get-url origin": "https://github.com/test-owner/test-repo.git",
		}

		expect := []string{
			"git remote get-url origin",
			"git remote set-url origin https://x-access-token:test-token@github.com/test-owner/test-repo.git", // trufflehog:ignore
			"git fetch origin asdf1234",
			"git fetch origin release-1.0.0:refs/remotes/origin/release-1.0.0",
			"git fetch --shallow-since=1577923200",
			"git checkout -b example origin/release-1.0.0",
			"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234",
			"git ls-files --error-unmatch .betterer.results",
			"git diff -s --exit-code .betterer.results",
			"git diff --name-only --diff-filter=U",
			"git cherry-pick --abort",
			"git remote set-url origin https://github.com/test-owner/test-repo.git",
		}

		require.Error(t, CreateCherryPickBranch(context.Background(), runner, branch, opts))
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
				MergeBase: &github.Commit{
					Committer: &github.CommitAuthor{
						Date: &github.Timestamp{
							Time: testCommitDate,
						},
					},
				},
			}
			runner = newErrorRunner(map[string]error{
				"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234": errors.New("cherry-pick error"),
			})
		)

		expect := []string{
			"git fetch origin asdf1234",
			"git fetch origin release-1.0.0:refs/remotes/origin/release-1.0.0",
			"git fetch --shallow-since=1577923200",
			"git checkout -b example origin/release-1.0.0",
			"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234",
			"git ls-files --error-unmatch .betterer.results",
			"git diff -s --exit-code .betterer.results",
			"git diff --name-only --diff-filter=U",
			"git cherry-pick --abort",
		}

		require.Error(t, CreateCherryPickBranch(context.Background(), runner, branch, opts))
		require.Equal(t, expect, runner.History.Commands)
	})

	t.Run("It should return an error if .betterer.results does not exist", func(t *testing.T) {
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
				MergeBase: &github.Commit{
					Committer: &github.CommitAuthor{
						Date: &github.Timestamp{
							Time: testCommitDate,
						},
					},
				},
			}
			runner = newErrorRunner(map[string]error{
				"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234": errors.New("cherry-pick error"),
				"git ls-files --error-unmatch .betterer.results":                                    errors.New("command returned 1"),
			})
		)

		expect := []string{
			"git fetch origin asdf1234",
			"git fetch origin release-1.0.0:refs/remotes/origin/release-1.0.0",
			"git fetch --shallow-since=1577923200",
			"git checkout -b example origin/release-1.0.0",
			"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234",
			"git ls-files --error-unmatch .betterer.results",
			"git diff --name-only --diff-filter=U",
			"git cherry-pick --abort",
		}

		require.Error(t, CreateCherryPickBranch(context.Background(), runner, branch, opts))
		require.Equal(t, expect, runner.History.Commands)
	})

	t.Run("It should handle go.sum conflicts", func(t *testing.T) {
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
				MergeBase: &github.Commit{
					Committer: &github.CommitAuthor{
						Date: &github.Timestamp{
							Time: testCommitDate,
						},
					},
				},
			}
			runner = newErrorRunner(map[string]error{
				"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234": errors.New("cherry-pick error"),
				// .betterer.results isn't tracked, so the betterer resolver bails out.
				"git ls-files --error-unmatch .betterer.results": errors.New("command returned 1"),
			})
		)
		// go.sum is the only conflicting file.
		runner.History.Outputs = map[string]string{
			"git diff --name-only --diff-filter=U": "go.sum\n",
		}

		expect := []string{
			"git fetch origin asdf1234",
			"git fetch origin release-1.0.0:refs/remotes/origin/release-1.0.0",
			"git fetch --shallow-since=1577923200",
			"git checkout -b example origin/release-1.0.0",
			"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234",
			"git ls-files --error-unmatch .betterer.results",
			"git diff --name-only --diff-filter=U",
			"go mod tidy",
			"git add go.mod go.sum",
			"git -c core.editor=true cherry-pick --continue",
		}

		require.NoError(t, CreateCherryPickBranch(context.Background(), runner, branch, opts))
		require.Equal(t, expect, runner.History.Commands)
	})

	t.Run("It should not fix go.sum when other files also conflict", func(t *testing.T) {
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
				MergeBase: &github.Commit{
					Committer: &github.CommitAuthor{
						Date: &github.Timestamp{
							Time: testCommitDate,
						},
					},
				},
			}
			runner = newErrorRunner(map[string]error{
				"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234": errors.New("cherry-pick error"),
				"git ls-files --error-unmatch .betterer.results":                                    errors.New("command returned 1"),
			})
		)
		runner.History.Outputs = map[string]string{
			"git diff --name-only --diff-filter=U": "go.sum\nmain.go\n",
		}

		expect := []string{
			"git fetch origin asdf1234",
			"git fetch origin release-1.0.0:refs/remotes/origin/release-1.0.0",
			"git fetch --shallow-since=1577923200",
			"git checkout -b example origin/release-1.0.0",
			"git -c user.name=grafanabot -c user.email=bot@grafana.com cherry-pick -x asdf1234",
			"git ls-files --error-unmatch .betterer.results",
			"git diff --name-only --diff-filter=U",
			"git cherry-pick --abort",
		}

		require.Error(t, CreateCherryPickBranch(context.Background(), runner, branch, opts))
		require.Equal(t, expect, runner.History.Commands)
	})
}
