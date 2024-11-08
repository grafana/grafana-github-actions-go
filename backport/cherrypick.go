package main

import (
	"context"
	"errors"
	"fmt"
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

	if _, err := runner.Run(ctx, "git", "cherry-pick", "--continue"); err != nil {
		return err
	}

	return nil
}

func CreateCherryPickBranch(ctx context.Context, runner CommandRunner, branch string, opts BackportOpts) error {
	if _, err := runner.Run(ctx, "git", "fetch"); err != nil {
		return fmt.Errorf("error fetching: %w", err)
	}

	if stderr, err := runner.Run(ctx, "git", "switch", "--create", branch, "origin/"+opts.Target); err != nil {
		return fmt.Errorf("error creating branch: %w\nstderr:%s", err, stderr)
	}

	_, err := runner.Run(ctx, "git", "cherry-pick", "-x", opts.SourceSHA)
	if err != nil {
		if err := ResolveBettererConflict(ctx, runner); err == nil {
			return nil
		}
		return fmt.Errorf("error running git cherry-pick: %w", err)
	}

	return nil
}
