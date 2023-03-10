package toolkit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v50/github"
)

type Toolkit struct {
	ghClient *github.Client
	token    string
}

func Init(ctx context.Context) (*Toolkit, error) {
	tk := &Toolkit{}
	token := tk.GetInput("GITHUB_TOKEN", nil)
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("neither INPUT_GITHUB_TOKEN nor GITHUB_TOKEN set")
	}
	client := github.NewTokenClient(ctx, token)
	tk.ghClient = client
	tk.token = token
	return tk, nil
}

func (tk *Toolkit) GitHubClient() *github.Client {
	return tk.ghClient
}

type GetInputOptions struct {
	TrimWhitespace bool
}

func (tk *Toolkit) GetInput(name string, opts *GetInputOptions) string {
	if opts == nil {
		opts = &GetInputOptions{}
	}
	name = strings.ToUpper(name)
	name = strings.ReplaceAll(name, " ", "_")
	val := os.Getenv("INPUT_" + name)
	if opts.TrimWhitespace {
		val = strings.TrimSpace(val)
	}
	return val
}

func (tk *Toolkit) CloneRepository(ctx context.Context, targetFolder string, ownerAndRepo string) error {
	cloneURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", tk.token, ownerAndRepo)
	cmd := exec.CommandContext(ctx, "git", "clone", cloneURL, targetFolder)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.CommandContext(ctx, "git", "config", "user.email", "bot@grafana.com")
	cmd.Dir = targetFolder
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.CommandContext(ctx, "git", "config", "user.name", "grafanabot")
	cmd.Dir = targetFolder
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (tk *Toolkit) BranchExists(ctx context.Context, owner, repo, branch string) (bool, error) {
	_, _, err := tk.GitHubClient().Repositories.GetBranch(ctx, owner, repo, branch, true)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
