package toolkit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghgql"
	"github.com/rs/zerolog"
)

const defaultMetricsAPIEndpoint = "https://graphite-us-central1.grafana.net/metrics"
const defaultMetricsAPIUsername = "6371"

type Toolkit struct {
	ghClient           *github.Client
	ghgqlClient        *ghgql.Client
	Token              string
	metricsAPIKey      string
	metricsAPIUsername string
	metricsAPIEndpoint string
	metricsNamePrefix  string
	requestCount       atomic.Int64
	registeredInputs   map[string]InputConfig
}

// IncrRequestCount increments an interval counter that is exposed as metric
// when calling the SubmitUsageMetrics method.
func (tk *Toolkit) IncrRequestCount() {
	tk.requestCount.Add(1)
}

type ToolkitOption func(tk *Toolkit)

func WithRegisteredInput(name, description string) ToolkitOption {
	return func(tk *Toolkit) {
		tk.registeredInputs[name] = InputConfig{
			Name:        name,
			Description: description,
		}
	}
}

func (tk *Toolkit) GetInputEnvName(name string) string {
	name = strings.ToUpper(name)
	name = strings.ReplaceAll(name, " ", "_")
	return fmt.Sprintf("INPUT_%s", name)
}

func Init(ctx context.Context, options ...ToolkitOption) (*Toolkit, error) {
	tk := &Toolkit{
		registeredInputs: make(map[string]InputConfig),
	}
	for _, opt := range options {
		opt(tk)
	}
	WithRegisteredInput("GITHUB_TOKEN", "Token used for interacting with the GitHub API")(tk)
	WithRegisteredInput("METRICS_API_USERNAME", "API username for metrics endpoint")(tk)
	WithRegisteredInput("METRICS_API_KEY", "API key (password) for metrics endpoint")(tk)
	WithRegisteredInput("METRICS_API_ENDPOINT", "API endpoint for metrics")(tk)
	token := tk.MustGetInput(ctx, "GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("neither INPUT_GITHUB_TOKEN nor GITHUB_TOKEN set")
	}
	metricsNamePrefix, err := determineMetricsNamePrefix(ctx)
	if err != nil {
		return nil, err
	}
	tk.metricsNamePrefix = metricsNamePrefix
	client := github.NewTokenClient(ctx, token)
	tk.ghClient = client
	tk.Token = token
	tk.ghgqlClient = ghgql.NewClient(token)
	tk.metricsAPIKey = tk.MustGetInput(ctx, "METRICS_API_KEY")
	tk.metricsAPIEndpoint = tk.MustGetInput(ctx, "METRICS_API_ENDPOINT")
	if tk.metricsAPIEndpoint == "" {
		tk.metricsAPIEndpoint = defaultMetricsAPIEndpoint
	}
	tk.metricsAPIUsername = tk.MustGetInput(ctx, "METRICS_API_USERNAME")
	if tk.metricsAPIUsername == "" {
		tk.metricsAPIUsername = defaultMetricsAPIUsername
	}
	return tk, nil
}

func (tk *Toolkit) GitHubClient() *github.Client {
	return tk.ghClient
}

func (tk *Toolkit) GitHubGQLClient() *ghgql.Client {
	return tk.ghgqlClient
}

type InputConfig struct {
	Name        string
	Description string
}

type GetInputOptions struct {
	TrimWhitespace bool
}

func (tk *Toolkit) GetInput(name string, opts *GetInputOptions) (string, error) {
	if opts == nil {
		opts = &GetInputOptions{}
	}
	if _, ok := tk.registeredInputs[name]; !ok {
		return "", fmt.Errorf("`%s` is not a registered input", name)
	}
	val := os.Getenv(tk.GetInputEnvName(name))
	if opts.TrimWhitespace {
		val = strings.TrimSpace(val)
	}
	return val, nil
}

func (tk *Toolkit) MustGetInput(ctx context.Context, name string) string {
	logger := zerolog.Ctx(ctx)
	input, err := tk.GetInput(name, nil)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Failed to retrieve input `%s`", name)
	}
	return input
}

func (tk *Toolkit) MustGetBoolInput(ctx context.Context, name string) bool {
	logger := zerolog.Ctx(ctx)
	input, err := tk.GetInput(name, nil)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Failed to retrieve input `%s`", name)
	}
	return input == "1"
}

func (tk *Toolkit) ShowInputList() {
	output := strings.Builder{}
	for _, input := range tk.registeredInputs {
		output.WriteString(input.Name)
		output.WriteString(" (")
		output.WriteString(tk.GetInputEnvName(input.Name))
		output.WriteString("):\n  ")
		output.WriteString(input.Description)
		output.WriteString("\n\n")
	}
	fmt.Fprint(os.Stdout, output.String())
}

func (tk *Toolkit) CloneRepository(ctx context.Context, targetFolder string, ownerAndRepo string) error {
	cloneURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", tk.Token, ownerAndRepo)
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
	tk.IncrRequestCount()
	_, _, err := tk.GitHubClient().Repositories.GetBranch(ctx, owner, repo, branch, true)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func determineMetricsNamePrefix(ctx context.Context) (string, error) {
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		return "", fmt.Errorf("GITHUB_REPOSITORY not set")
	}
	elems := strings.SplitN(repo, "/", 2)
	owner := elems[0]
	name := elems[1]
	if owner == "grafana" && name == "grafana" {
		return "gh_action", nil
	}
	return fmt.Sprintf("repo_stats.%s", name), nil
}
