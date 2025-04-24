package main

import (
	"context"
	"errors"
	"fmt"
)

func DeepenUntil(ctx context.Context, runner CommandRunner, sha string, size int, maxFetches int) error {
	for i := 0; i < maxFetches; i++ {
		if _, err := runner.Run(ctx, "git", "fetch", fmt.Sprintf("--deepen=%d", size)); err != nil {
			return err
		}

		if _, err := runner.Run(ctx, "git", "rev-parse", "--verify", sha); err == nil {
			return nil
		}
	}

	return errors.New("commit not found")
}
