package toolkit

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v50/github"
)

type Toolkit struct {
	ghClient *github.Client
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
