package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/changelog"
	"github.com/grafana/grafana-github-actions-go/pkg/toolkit"
	"github.com/spf13/pflag"
)

// LatestString returns a string that the GitHub API expects.
// Their docs say that this string can be one of: "true", "false", or "legacy".
func LatestString(latest bool) *string {
	if latest {
		return github.String("true")
	}

	return github.String("false")
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var ownerRepo string
	var doPreview bool
	var listInputs bool
	var version string

	pflag.StringVar(&ownerRepo, "repo", os.Getenv("GITHUB_REPOSITORY"), "owner/repo pair for a repository on GitHub")
	pflag.BoolVar(&doPreview, "preview", false, "Only determine the milestone but don't set it")
	pflag.BoolVar(&listInputs, "list-inputs", false, "Show a list of all available inputs")
	pflag.Parse()

	log.Info("starting GitHub release")

	version = strings.TrimPrefix(pflag.Arg(0), "v")
	tag := fmt.Sprintf("v%s", version)

	tk, err := toolkit.Init(
		ctx,
		toolkit.WithRegisteredInput("latest", "`true` for marking the release as latest, otherwise not"),
	)
	if err != nil {
		log.Error("failed to initialize toolkit", "error", err)
		panic("failed to initialize toolkit")
	}

	if listInputs {
		tk.ShowInputList()
		return
	}
	defer func() {
		if err := tk.SubmitUsageMetrics(ctx); err != nil {
			log.Warn("failed to submit usage metrics", "error", err)
		}
	}()

	log = log.With("tag", tag, "repo", ownerRepo)
	latest := tk.MustGetBoolInput(ctx, "latest")

	elems := strings.Split(ownerRepo, "/")
	owner := elems[0]
	repo := elems[1]

	gh := tk.GitHubClient()

	changelogContent, err := retrieveChangelog(ctx, gh, owner, repo, version)
	if err != nil {
		panic(fmt.Sprintf("failed to retrieve changelog for %s", version))
	}

	releaseTitle := version

	if doPreview {
		fmt.Println("no release will be created but this is what it would look like:")
		fmt.Printf("TITLE: %s\n\n%s\n", releaseTitle, changelogContent)
		return
	}

	newRelease := &github.RepositoryRelease{
		TagName:    github.String(tag),
		Name:       github.String(releaseTitle),
		Body:       github.String(changelogContent),
		MakeLatest: LatestString(latest),
	}

	url, err := CreateRelease(ctx, gh.Repositories, owner, repo, tag, newRelease)
	if err != nil {
		panic(err)
	}

	log.Info("release available", "url", url)
}

func retrieveChangelog(ctx context.Context, gh *github.Client, owner string, repo string, version string) (string, error) {
	loader := changelog.NewLoader(gh)
	output, err := loader.LoadContent(ctx, owner, repo, version, &changelog.LoaderOptions{
		RemoveHeading: true,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`[Download page](https://grafana.com/grafana/download/%s)
[What's new highlights](https://grafana.com/docs/grafana/latest/whatsnew/)

%s`, version, output), nil
}
