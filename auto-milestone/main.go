package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/toolkit"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	ctx = logger.WithContext(ctx)

	var repo string
	var doPreview bool
	var listInputs bool
	var prNumber int

	pflag.StringVar(&repo, "repo", os.Getenv("GITHUB_REPOSITORY"), "owner/repo pair for a repository on GitHub")
	pflag.BoolVar(&doPreview, "preview", false, "Only determine the milestone but don't set it")
	pflag.BoolVar(&listInputs, "list-inputs", false, "Show a list of all available inputs")
	pflag.Parse()

	rawPRNumber := pflag.Arg(0)
	if rawPRNumber == "" {
		logger.Fatal().Msg("No PR specified")
		return
	}
	if parsed, err := strconv.ParseInt(rawPRNumber, 10, 32); err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse PR number")
		return
	} else {
		prNumber = int(parsed)
	}

	// Determine the base-branch of that pull request
	tk, err := toolkit.Init(
		ctx,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize toolkit")
	}

	if listInputs {
		tk.ShowInputList()
		return
	}

	elems := strings.Split(repo, "/")
	repoOwner := elems[0]
	repoName := elems[1]

	gh := tk.GitHubClient()

	pr, _, err := gh.PullRequests.Get(ctx, repoOwner, repoName, prNumber)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to fetch pull-request")
	}
	prTargetBranch := pr.GetBase()
	if prTargetBranch == nil {
		logger.Fatal().Msg("The requested pull-request has no target branch")
		return
	}

	var targetMilestoneName string

	content, _, _, err := gh.Repositories.GetContents(ctx, prTargetBranch.GetRepo().GetOwner().GetLogin(), prTargetBranch.GetRepo().GetName(), "package.json", &github.RepositoryContentGetOptions{Ref: prTargetBranch.GetRef()})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to retrieve package.json of PR base")
		return
	}
	if cnt, err := content.GetContent(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to retrieve package.json of PR base")
		return
	} else {
		if v, err := versionFromPackage(cnt); err != nil {
			logger.Fatal().Err(err).Msg("Failed to determine base version")
			return
		} else {
			targetMilestoneName = v
		}
	}

	logger.Info().Msgf("Target milestone name: %s", targetMilestoneName)
	milestone, err := tk.GitHubGQLClient().GetMilestoneByTitle(ctx, repoOwner, repoName, targetMilestoneName)
	if err != nil {
		logger.Fatal().Msgf("Failed to find milestone matching `%s`", targetMilestoneName)
	}
	logger.Info().Msgf("Milestone number: %d", milestone.Number)
	if doPreview {
		return
	}

	// TODO: Check if the PR has a milestone. If it does and matches the one we
	// picked, there's nothing to do. Otherwise overwrite that milestone.
}

type packageJSON struct {
	Version string `json:"version"`
}

var versionPattern = regexp.MustCompile(`^(\d+)\.(\d+).(\d+)`)

func versionFromPackage(content string) (string, error) {
	pjson := packageJSON{}
	if err := json.Unmarshal([]byte(content), &pjson); err != nil {
		return "", err
	}
	v := pjson.Version
	match := versionPattern.FindStringSubmatch(v)
	if len(match) < 4 {
		return "", fmt.Errorf("unsupported version format")
	}
	return fmt.Sprintf("%s.%s.x", match[1], match[2]), nil
}
