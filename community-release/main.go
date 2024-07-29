package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/changelog"
	"github.com/grafana/grafana-github-actions-go/pkg/community"
	"github.com/grafana/grafana-github-actions-go/pkg/toolkit"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

const inputCommunityAPIKey = "COMMUNITY_API_KEY"
const inputCommunityAPIUsername = "COMMUNITY_API_USERNAME"
const inputCommunityCategoryID = "COMMUNITY_CATEGORY_ID"
const inputCommunityBaseURL = "COMMUNITY_BASE_URL"
const inputVersion = "VERSION"
const defaultCategoryID = "9"
const defaultBaseURL = "https://community.grafana.com/"

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	ctx = logger.WithContext(ctx)

	var repo string
	var doPreview bool
	var listInputs bool

	pflag.StringVar(&repo, "repo", os.Getenv("GITHUB_REPOSITORY"), "owner/repo pair for a repository on GitHub")
	pflag.BoolVar(&doPreview, "preview", false, "Print the community post instead of posting it")
	pflag.BoolVar(&listInputs, "list-inputs", false, "Show a list of all available inputs")
	pflag.Parse()

	logger.Info().Msgf("Operating inside %s", repo)

	tk, err := toolkit.Init(
		ctx,
		toolkit.WithRegisteredInput(inputVersion, "Version number to generate the changelog for"),
		toolkit.WithRegisteredInput(inputCommunityAPIKey, "API token for the Discourse community"),
		toolkit.WithRegisteredInput(inputCommunityAPIUsername, "API username for the Discourse community"),
		toolkit.WithRegisteredInput(inputCommunityCategoryID, "Discourse category ID for the changelog post"),
		toolkit.WithRegisteredInput(inputCommunityBaseURL, "URL where the Discourse community can be found"),
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

	version := tk.MustGetInput(ctx, inputVersion)

	elems := strings.Split(repo, "/")
	repoOwner := elems[0]
	repoName := elems[1]

	gh := tk.GitHubClient()

	changelogContent, err := retrieveChangelog(ctx, gh, repoOwner, repoName, version)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Failed to retrieve changelog for %s", version)
	}

	logger.Info().Msgf("Changelog received with %d characters", utf8.RuneCountInString(changelogContent))

	releaseTitle := fmt.Sprintf("Changelog: Updates in Grafana %s", version)

	if doPreview {
		logger.Info().Msgf("No post will be created but this is what it would look like:")
		fmt.Printf("TITLE: %s\n\n%s\n", releaseTitle, changelogContent)
		return
	}

	key := tk.MustGetInput(ctx, inputCommunityAPIKey)
	username := tk.MustGetInput(ctx, inputCommunityAPIUsername)
	communityBaseURL := tk.MustGetInput(ctx, inputCommunityBaseURL)
	rawCommunityCategoryID := tk.MustGetInput(ctx, inputCommunityCategoryID)
	if key == "" {
		logger.Fatal().Msgf("No %s provided", tk.GetInputEnvName(inputCommunityAPIKey))
	}
	if username == "" {
		logger.Fatal().Msgf("No %s provided", tk.GetInputEnvName(inputCommunityAPIUsername))
	}
	if communityBaseURL == "" {
		communityBaseURL = defaultBaseURL
	}
	if rawCommunityCategoryID == "" {
		rawCommunityCategoryID = defaultCategoryID
	}
	communityCategoryID, err := strconv.Atoi(rawCommunityCategoryID)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Failed to parse %s", tk.GetInputEnvName(inputCommunityCategoryID))
	}

	logger.Info().Msgf("Posting to the community board in category %d", communityCategoryID)
	comm := community.New(
		community.CommunityWithBaseURL(communityBaseURL),
		community.CommunityWithAPICredentials(username, key),
	)
	if _, err := comm.CreateOrUpdatePost(ctx, community.PostInput{
		Title:    releaseTitle,
		Author:   username,
		Body:     changelogContent,
		Category: communityCategoryID,
	}, &community.PostOptions{
		FallbackBody: fallbackChangelog(version),
	}); err != nil {
		logger.Fatal().Err(err).Msg("Failed to post to the forums")
	}

}

func retrieveChangelog(ctx context.Context, gh *github.Client, repoOwner string, repoName string, version string) (string, error) {
	loader := changelog.NewLoader(gh)
	output, err := loader.LoadContent(ctx, repoOwner, repoName, version, &changelog.LoaderOptions{
		RemoveHeading: true,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`%s

%s`, output, changelogFooter(version)), nil
}

func changelogFooter(version string) string {
	return fmt.Sprintf(`[Download page](https://grafana.com/grafana/download/%s)
[What's new highlights](https://grafana.com/docs/grafana/latest/whatsnew/)`, version)
}

func fallbackChangelog(version string) string {
	return fmt.Sprintf(`[Full changelog](https://github.com/grafana/grafana/releases/tag/v%s)
%s`, version, changelogFooter(version))
}
