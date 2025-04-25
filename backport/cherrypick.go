package main

import (
	"context"
	"errors"
	"fmt"
	"time"
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
	// 1. Ensure that we have the commit in the local history to cherry-pick
	if _, err := runner.Run(ctx, "git", "fetch", fmt.Sprintf("--shallow-since=\"%s\"", opts.SourceCommitDate.Add(-1*24*time.Minute).Format("2006-01-02"))); err != nil {
		return fmt.Errorf("error fetching source commit: %w", err)
	}

	// 2. Ensure that the backport branch is in the local history.
	if _, err := runner.Run(ctx, "git", "fetch", "origin", fmt.Sprintf("%[1]s:refs/remotes/origin/%[1]s", opts.Target.Name)); err != nil {
		return fmt.Errorf("error fetching target branch: %w", err)
	}

	if _, err := runner.Run(ctx, "git", "checkout", "-b", branch, "--track", "origin/"+opts.Target.Name); err != nil {
		return fmt.Errorf("error creating branch: %w", err)
	}

	_, err := runner.Run(ctx, "git", "cherry-pick", "-x", opts.SourceSHA)
	if err != nil {
		if err := ResolveBettererConflict(ctx, runner); err == nil {
			return nil
		}

		runner.Run(ctx, "git", "cherry-pick", "--abort")

		return fmt.Errorf("error running git cherry-pick: %w", err)
	}

	return nil
}
