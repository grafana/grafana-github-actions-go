package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func ResolveBettererConflict(ctx context.Context, runner CommandRunner) error {
	// git diff -s --exit-code returns 1 if the file has changed
	if _, err := runner.Run(ctx, "git", "diff", "-s", "--exit-code", ".betterer.results"); err == nil {
		return errors.New(".better.results has not changed")
	}

	if _, err := runner.Run(ctx, "yarn", "run", "betterer"); err != nil {
		return err
	}

	if _, err := runner.Run(ctx, "git", "add", ".betterer.results"); err != nil {
		return err
	}

	if _, err := runner.Run(ctx, "git", "-c", "core.editor=true", "cherry-pick", "--continue"); err != nil {
		return err
	}

	return nil
}

func CreateCherryPickBranch(ctx context.Context, runner CommandRunner, branch string, opts BackportOpts) error {
	if opts.GitToken != "" {
		origURL, err := runner.Run(ctx, "git", "remote", "get-url", "origin")
		if err != nil {
			return fmt.Errorf("error getting origin URL: %w", err)
		}
		authURL := strings.Replace(origURL, "https://github.com/", "https://x-access-token:"+opts.GitToken+"@github.com/", 1)
		if _, err := runner.Run(ctx, "git", "remote", "set-url", "origin", authURL); err != nil {
			return fmt.Errorf("error setting origin URL: %w", err)
		}
		defer runner.Run(ctx, "git", "remote", "set-url", "origin", origURL) //nolint:errcheck
	}

	// 1. Ensure that we have the commit in the local history to cherry-pick
	if _, err := runner.Run(ctx, "git", "fetch", "origin", opts.SourceSHA); err != nil {
		return fmt.Errorf("error fetching source commit: %w", err)
	}

	// 2. Ensure that the backport branch is in the local history.
	if _, err := runner.Run(ctx, "git", "fetch", "origin", fmt.Sprintf("%[1]s:refs/remotes/origin/%[1]s", opts.Target.Name)); err != nil {
		return fmt.Errorf("error fetching target branch: %w", err)
	}

	// 3 Ensure that we have enough context in the local history to cherry-pick
	if _, err := runner.Run(ctx, "git", "fetch", fmt.Sprintf("--shallow-since=%d", opts.MergeBase.Committer.Date.Unix())); err != nil {
		return fmt.Errorf("error fetching source commit: %w", err)
	}

	if _, err := runner.Run(ctx, "git", "checkout", "-b", branch, "origin/"+opts.Target.Name); err != nil {
		return fmt.Errorf("error creating branch: %w", err)
	}

	// `git cherry-pick` refuses to create the local commit without a committer identity, so
	// pass one inline via `-c`. The local commit's committer is overwritten by GitHub's
	// web-flow key when the commit is published via createCommitOnBranch, so the value is
	// purely a placeholder — but it must be non-empty. Using `-c` keeps `.git/config`
	// untouched. The identity matches the one set by pkg/toolkit/toolkit.go.
	_, err := runner.Run(ctx, "git",
		"-c", "user.name=grafanabot",
		"-c", "user.email=bot@grafana.com",
		"cherry-pick", "-x", opts.SourceSHA,
	)
	if err != nil {
		if err := ResolveBettererConflict(ctx, runner); err == nil {
			return nil
		}

		runner.Run(ctx, "git", "cherry-pick", "--abort")

		return fmt.Errorf("error running git cherry-pick: %w", err)
	}

	return nil
}
