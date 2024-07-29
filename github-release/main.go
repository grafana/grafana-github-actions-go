package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/changelog"
	"github.com/grafana/grafana-github-actions-go/pkg/ghgql"
	"github.com/grafana/grafana-github-actions-go/pkg/toolkit"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

var makeLatestTrue = "true"
var makeLatestFalse = "false"

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	ctx = logger.WithContext(ctx)

	var repo string
	var doPreview bool
	var listInputs bool
	var version string

	pflag.StringVar(&repo, "repo", os.Getenv("GITHUB_REPOSITORY"), "owner/repo pair for a repository on GitHub")
	pflag.BoolVar(&doPreview, "preview", false, "Only determine the milestone but don't set it")
	pflag.BoolVar(&listInputs, "list-inputs", false, "Show a list of all available inputs")
	pflag.Parse()

	logger.Info().Msgf("Operating inside %s", repo)

	version = pflag.Arg(0)
	tag := fmt.Sprintf("v%s", version)

	tk, err := toolkit.Init(
		ctx,
		toolkit.WithRegisteredInput("latest", "`true` for marking the release as latest, otherwise not"),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize toolkit")
	}

	if listInputs {
		tk.ShowInputList()
		return
	}
	defer func() {
		if err := tk.SubmitUsageMetrics(ctx); err != nil {
			logger.Warn().Err(err).Msg("Failed to submit usage metrics")
		}
	}()

	markLatest := tk.MustGetBoolInput(ctx, "latest")

	elems := strings.Split(repo, "/")
	repoOwner := elems[0]
	repoName := elems[1]

	gh := tk.GitHubClient()

	tagExists, err := verifyTagExists(ctx, gh, repoOwner, repoName, tag)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Failed to verify that tag `%s` exists", version)
	}

	if !tagExists {
		logger.Fatal().Err(err).Msgf("Tag `%s` does not exist", version)
	}

	changelogContent, err := retrieveChangelog(ctx, gh, repoOwner, repoName, version)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Failed to retrieve changelog for %s", version)
	}

	releaseTitle := version

	existingRelease, resp, err := gh.Repositories.GetReleaseByTag(ctx, repoOwner, repoName, tag)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			logger.Info().Msgf("No existing release found for tag `%s`", tag)
		} else {
			logger.Fatal().Err(err).Msgf("Failed to check for existing release with tag `%s`", tag)
		}
	}

	if doPreview {
		logger.Info().Msgf("No release will be created but this is what it would look like:")
		fmt.Printf("TITLE: %s\n\n%s\n", releaseTitle, changelogContent)
		return
	}

	if existingRelease != nil {
		logger.Info().Msgf("Updating existing release")
		existingRelease.Name = &releaseTitle
		existingRelease.Body = &changelogContent
		if markLatest {
			existingRelease.MakeLatest = &makeLatestTrue
		} else {
			existingRelease.MakeLatest = &makeLatestFalse
		}
		if rel, _, err := gh.Repositories.EditRelease(ctx, repoOwner, repoName, existingRelease.GetID(), existingRelease); err != nil {
			logger.Fatal().Err(err).Msgf("Failed to update existing release")
		} else {
			logger.Info().Msgf("Release updated: %s", rel.GetHTMLURL())
		}
	} else {
		logger.Info().Msgf("Creating new release.")
		newRelease := &github.RepositoryRelease{}
		newRelease.TagName = &tag
		newRelease.Name = &releaseTitle
		newRelease.Body = &changelogContent
		if markLatest {
			newRelease.MakeLatest = &makeLatestTrue
		} else {
			newRelease.MakeLatest = &makeLatestFalse
		}
		if rel, _, err := gh.Repositories.CreateRelease(ctx, repoOwner, repoName, newRelease); err != nil {
			logger.Fatal().Err(err).Msgf("Failed to create new release")
		} else {
			logger.Info().Msgf("Release created: %s", rel.GetHTMLURL())
		}
	}
}

func generateReleaseTitle(ctx context.Context, version string, milestone *ghgql.Milestone) string {
	date := milestone.ClosedAt
	if !milestone.DueOn.IsZero() {
		date = milestone.DueOn
	}
	if date.IsZero() {
		date = time.Now()
	}
	return fmt.Sprintf("%s (%s)", version, date.Format("2006-01-02"))
}

func verifyTagExists(ctx context.Context, gh *github.Client, repoOwner string, repoName string, tag string) (bool, error) {
	opts := github.ListOptions{}
	opts.Page = 1
	for {
		tags, resp, err := gh.Repositories.ListTags(ctx, repoOwner, repoName, &opts)
		if err != nil {
			return false, err
		}
		for _, t := range tags {
			if t.GetName() == tag {
				return true, nil
			}
		}
		if resp.NextPage <= opts.Page {
			break
		}
		opts.Page = resp.NextPage
	}
	return false, nil
}

func retrieveChangelog(ctx context.Context, gh *github.Client, repoOwner string, repoName string, version string) (string, error) {
	loader := changelog.NewLoader(gh)
	output, err := loader.LoadContent(ctx, repoOwner, repoName, version, &changelog.LoaderOptions{
		RemoveHeading: true,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`[Download page](https://grafana.com/grafana/download/%s)
[What's new highlights](https://grafana.com/docs/grafana/latest/whatsnew/)

%s`, version, output), nil
}
