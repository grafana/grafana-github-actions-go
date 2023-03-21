package git

import (
	"context"
	"os"
	"os/exec"
)

type RepositoryClient struct {
	path string
}

func NewRepository(path string) *RepositoryClient {
	return &RepositoryClient{
		path: path,
	}
}

func (c *RepositoryClient) Exec(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = c.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *RepositoryClient) ListBranches(ctx context.Context) ([]string, error) {
	result := make([]string, 0, 10)
	return result, nil
}
