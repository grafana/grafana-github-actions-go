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
	"github.com/grafana/grafana-github-actions-go/pkg/ghgql"
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

	logger.Info().Msgf("Operating inside %s", repo)

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
		toolkit.WithRegisteredInput("version_source_repository", "owner/repo of the repository to check for a package.json file"),
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

	repoOwner, repoName := splitRepo(repo)
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

	versionSourceRepoOverride := tk.MustGetInput(ctx, "version_source_repository")
	if versionSourceRepoOverride == "" {
		versionSourceRepoOverride = fmt.Sprintf("%s/%s", prTargetBranch.GetRepo().GetOwner().GetLogin(), prTargetBranch.GetRepo().GetName())
	}
	versionSourceOwner, versionSourceName := splitRepo(versionSourceRepoOverride)

	var targetMilestoneName string

	content, _, _, err := gh.Repositories.GetContents(ctx, versionSourceOwner, versionSourceName, "package.json", &github.RepositoryContentGetOptions{Ref: prTargetBranch.GetRef()})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to retrieve package.json of PR base")
		return
	}
	if cnt, err := content.GetContent(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to get package.json content")
		return
	} else {
		if v, err := versionFromPackage(cnt); err != nil {
			logger.Fatal().Err(err).Msg("Failed to determine base version")
			return
		} else {
			targetMilestoneName = v
		}
	}

	prMilestone := pr.GetMilestone()

	if prMilestone != nil {
		logger.Info().Msgf("Current milestone: %s", prMilestone.GetTitle())
	}
	logger.Info().Msgf("Target milestone name: %s", targetMilestoneName)
	milestone, err := tk.GitHubGQLClient().GetMilestoneByTitle(ctx, repoOwner, repoName, targetMilestoneName)
	if err != nil {
		logger.Fatal().Msgf("Failed to find milestone matching `%s`", targetMilestoneName)
	}
	if milestone == nil {
		logger.Fatal().Msgf("Milestone not found: `%s`", targetMilestoneName)
	}

	a := determineAction(ctx, pr, prMilestone, milestone)

	if doPreview {
		logger.Info().Msgf("The following action would be performed: %v", a)
		return
	}

	switch a.Type {
	case actionTypeNoop:
		logger.Info().Msgf("No action necessary.")
		return
	case actionTypeSetToMilestone:
		logger.Info().Msgf("Updating PR to %s", a.Milestone)
		if a.Milestone != nil {
			if _, resp, err := gh.Issues.Edit(ctx, repoOwner, repoName, prNumber, &github.IssueRequest{
				Milestone: &a.Milestone.Number,
			}); err != nil || resp.StatusCode >= 300 {
				logger.Fatal().Err(err).Msgf("Failed to update #%d with new milestone", prNumber)
				return
			}
		} else {
			if _, resp, err := gh.Issues.RemoveMilestone(ctx, repoOwner, repoName, prNumber); err != nil || resp.StatusCode >= 300 {
				logger.Fatal().Err(err).Msgf("Failed to remove milestone from #%d", prNumber)
				return
			}
		}
	}

	logger.Info().Msgf("PR updated")
}

type actionType int

const (
	actionTypeSetToMilestone = iota
	actionTypeNoop
)

type action struct {
	Type      actionType
	Milestone *ghgql.Milestone
}

func (a action) String() string {
	switch a.Type {
	case actionTypeNoop:
		return "no action"
	case actionTypeSetToMilestone:
		return fmt.Sprintf("set milestone to %s", a.Milestone)
	default:
		return "<unknown action>"
	}
}

func determineAction(ctx context.Context, pr *github.PullRequest, currentMilestone *github.Milestone, targetMilestone *ghgql.Milestone) action {
	logger := zerolog.Ctx(ctx)
	if pr.ClosedAt != nil && !pr.GetMerged() {
		// If this PR is closed but wasn't merged, then we remove the milestone:
		if currentMilestone != nil {
			logger.Info().Msg("PR is closed but was not merged. Unsetting the milestone.")
			return action{
				Type:      actionTypeSetToMilestone,
				Milestone: nil,
			}
		} else {
			logger.Info().Msg("PR is closed but was not merged. It has no milestone and so none was set.")
			return action{
				Type: actionTypeNoop,
			}
		}
	}
	prTargetBranchLabel := pr.GetBase().GetLabel()
	if prTargetBranchLabel != "grafana:main" && !strings.HasSuffix(prTargetBranchLabel, ".x") {
		logger.Info().Msgf("The PR is targeting branch %s, which does not match either main or a release branch. No action required.", prTargetBranchLabel)
		return action{
			Type: actionTypeNoop,
		}
	}
	if currentMilestone == nil {
		return action{
			Type:      actionTypeSetToMilestone,
			Milestone: targetMilestone,
		}
	}
	targetMilestoneTitle := targetMilestone.Title
	currentMilestoneTitle := currentMilestone.GetTitle()
	if targetMilestoneTitle == currentMilestoneTitle {
		logger.Info().Msg("The PR already has the correct milestone.")
		return action{
			Type: actionTypeNoop,
		}
	}
	if !strings.HasSuffix(currentMilestoneTitle, ".x") {
		logger.Info().Msg("The PR has a release milestone attached so no action required.")
		return action{
			Type: actionTypeNoop,
		}
	}
	return action{
		Type:      actionTypeSetToMilestone,
		Milestone: targetMilestone,
	}
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

func splitRepo(repo string) (string, string) {
	elems := strings.Split(repo, "/")
	repoOwner := elems[0]
	repoName := elems[1]
	return repoOwner, repoName
}
