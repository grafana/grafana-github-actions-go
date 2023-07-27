package changelog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/grafana/grafana-github-actions-go/pkg/ghgql"
	"github.com/grafana/grafana-github-actions-go/pkg/toolkit"
	"github.com/rs/zerolog"

	"github.com/google/go-github/v50/github"
)

const LabelEnterprise = "enterprise"
const LabelUI = "area/grafana/ui"
const LabelToolkit = "area/grafana/toolkit"
const LabelRuntime = "area/grafana/runtime"
const LabelBug = "type/bug"

func Build(ctx context.Context, version string, tk *toolkit.Toolkit) (*ChangelogBody, error) {
	logger := zerolog.Ctx(ctx)
	body := newChangelogBody()

	milestone, err := getMilestone(ctx, tk, "grafana/grafana", version)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve OSS milestone: %w", err)
	}
	enterpriseMilestone, err := getMilestone(ctx, tk, "grafana/grafana-enterprise", version)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve OSS milestone: %w", err)
	}

	ossIssues, err := tk.GitHubGQLClient().GetMilestonedPRsForChangelog(ctx, "grafana", "grafana", milestone.GetNumber())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve OSS issues: %w", err)
	}

	enterpriseIssues, err := tk.GitHubGQLClient().GetMilestonedPRsForChangelog(ctx, "grafana", "grafana-enterprise", enterpriseMilestone.GetNumber())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Enterprise issues: %w", err)
	}

	issues := make([]ghgql.PullRequest, 0, len(ossIssues)+len(enterpriseIssues))
	issues = append(issues, ossIssues...)
	issues = append(issues, enterpriseIssues...)

	// At this point check if the PR is already part of an older release.
	// Basically any milestone that was part of the stream and the previous one
	// released before the current milestone should be considered a potential
	// conflict.
	milestones, err := getHistoricalMilestones(ctx, tk, milestone, version)
	if err != nil {
		return nil, err
	}

	knownTitles := make(map[string]struct{})

	loader := NewLoader(tk.GitHubClient())
	parser := NewParser()
	for _, milestone := range milestones {
		logger.Info().Msgf("Considering %s for duplicates", milestone.GetTitle())
		msContent, err := loader.LoadContent(ctx, "grafana", "grafana", milestone.GetTitle(), &LoaderOptions{RemoveHeading: true})
		if err != nil {
			var noChangelogFound NoChangelogFound
			if errors.As(err, &noChangelogFound) {
				logger.Warn().Msgf("No changelog found for %s", noChangelogFound.Version)
				continue
			}
			return nil, err
		}
		sections, err := parser.Parse(ctx, bytes.NewBufferString(msContent))
		if err != nil {
			return nil, err
		}
		for _, section := range sections {
			for _, entry := range section.Entries {
				knownTitles[entry.Title] = struct{}{}
			}
		}
	}

	body.Version = version
	if !milestone.GetDueOn().IsZero() {
		body.ReleaseDate = milestone.GetDueOn().Format("2006-01-02")
	} else if !milestone.GetClosedAt().IsZero() {
		body.ReleaseDate = milestone.GetClosedAt().Format("2006-01-02")
	}
	numDups := 0
	for _, i := range issues {
		// If the PR already seems to be present in a previous release, we can
		// skip it here:
		newTitle := PreparePRTitle(i)
		if _, found := knownTitles[strings.TrimSpace(newTitle)]; found {
			logger.Debug().Msgf("`%s` (#%d) was already mentioned in a previous release", i.GetTitle(), i.GetNumber())
			numDups++
			continue
		}
		addToBody(body, i)
	}
	logger.Info().Msgf("%d duplicates skipped", numDups)
	return body, nil
}

// getHistoricalMilestones retrieves all the milestones of the current and
// previous release stream that were closed n-days before the milestone
// matching `version`.
func getHistoricalMilestones(ctx context.Context, tk *toolkit.Toolkit, currentMilestone *github.Milestone, version string) ([]*github.Milestone, error) {
	logger := zerolog.Ctx(ctx)
	result := make([]*github.Milestone, 0, 10)
	if strings.HasSuffix(version, ".x") {
		version = strings.Replace(version, ".x", ".0", 1)
	}
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, err
	}
	allMilestones, err := getAllMilestones(ctx, tk, "grafana", "grafana")
	if err != nil {
		return nil, err
	}
	// Now we need to find the previous minor release so that we can then filter based on that:
	previousMinorVersion, err := getPreviousMinorRelease(ctx, allMilestones, *v)
	if err != nil {
		return nil, err
	}
	logger.Info().Msgf("Previous minor version: %s", previousMinorVersion)

	// Now we need to find all the milestones for that particular minor version
	// that have been closed before the date of the version-milestone:
	currentDueDate := getMilestoneDate(currentMilestone)
	// If the milestone hasn't been closed yet (e.g. relevant for previewing
	// new releases), then we assume that it was closed just now:
	if currentDueDate.IsZero() {
		currentDueDate = time.Now()
	}
	for _, otherMilestone := range allMilestones {
		if !isInMinorRelease(otherMilestone, previousMinorVersion) {
			continue
		}
		otherDueDate := getMilestoneDate(otherMilestone)
		if otherDueDate.IsZero() {
			continue
		}
		if otherDueDate.Before(currentDueDate) {
			result = append(result, otherMilestone)
		}
	}
	return result, nil
}

func getMilestoneDate(milestone *github.Milestone) time.Time {
	if milestone.DueOn != nil {
		return milestone.DueOn.Time
	}
	if milestone.ClosedAt != nil {
		return milestone.ClosedAt.Time
	}
	return time.Time{}
}

func isInMinorRelease(milestone *github.Milestone, version *semver.Version) bool {
	title := milestone.GetTitle()
	v, err := semver.NewVersion(title)
	if err != nil || v == nil {
		return false
	}
	return v.Major == version.Major && v.Minor == version.Minor
}

func getAllMilestones(ctx context.Context, tk *toolkit.Toolkit, owner, repo string) ([]*github.Milestone, error) {
	opts := &github.MilestoneListOptions{}
	opts.State = "all"
	opts.Page = 1
	result := make([]*github.Milestone, 0, 20)
	for {
		milestones, resp, err := tk.GitHubClient().Issues.ListMilestones(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		result = append(result, milestones...)
		if resp.NextPage <= opts.Page {
			break
		}
		opts.Page = resp.NextPage
	}
	return result, nil
}

// getPreviousMinorRelease tries to find the previous minor release of the provided version
func getPreviousMinorRelease(ctx context.Context, allMilestones []*github.Milestone, version semver.Version) (*semver.Version, error) {
	currentMinor := version
	currentMinor.Patch = 0
	// For this it should be enough to go through all the ".x" milestones:
	var candidateVersion *semver.Version
	for _, m := range allMilestones {
		if strings.HasSuffix(m.GetTitle(), ".x") {
			mTitle := m.GetTitle()
			mTitle = strings.Replace(mTitle, ".x", ".0", 1)
			mVersion, err := semver.NewVersion(mTitle)
			if err != nil {
				continue
			}
			if mVersion.LessThan(currentMinor) {
				if candidateVersion == nil || !mVersion.LessThan(*candidateVersion) {
					candidateVersion = mVersion
				}
			}
		}
	}
	if candidateVersion != nil {
		return candidateVersion, nil
	}
	return nil, nil
}

type ChangelogBody struct {
	Version            string
	ReleaseDate        string
	DeprecationChanges []string
	BreakingChanges    []string
	PluginDevChanges   []ghgql.PullRequest
	Bugfixes           []ghgql.PullRequest
	Features           []ghgql.PullRequest
}

func newChangelogBody() *ChangelogBody {
	return &ChangelogBody{
		DeprecationChanges: make([]string, 0, 10),
		BreakingChanges:    make([]string, 0, 10),
		PluginDevChanges:   make([]ghgql.PullRequest, 0, 10),
		Bugfixes:           make([]ghgql.PullRequest, 0, 10),
		Features:           make([]ghgql.PullRequest, 0, 10),
	}
}

func addToBody(body *ChangelogBody, issue ghgql.PullRequest) {
	if notice := getBreakingChangeNotice(issue); notice != "" {
		body.BreakingChanges = append(body.BreakingChanges, notice)
	}
	if notice := getDeprecationNotice(issue); notice != "" {
		body.DeprecationChanges = append(body.DeprecationChanges, notice)
	}

	if issueHasLabel(issue, LabelToolkit) || issueHasLabel(issue, LabelUI) || issueHasLabel(issue, LabelRuntime) {
		body.PluginDevChanges = append(body.PluginDevChanges, issue)
		return
	}

	if isBug(issue) {
		body.Bugfixes = append(body.Bugfixes, issue)
	} else {
		body.Features = append(body.Features, issue)
	}
}

func isBug(issue ghgql.PullRequest) bool {
	title := issue.GetTitle()
	if strings.Contains(strings.ToLower(title), "fix") {
		return true
	}
	if issueHasLabel(issue, LabelBug) {
		return true
	}
	return false
}

func getMilestone(ctx context.Context, tk *toolkit.Toolkit, repo string, version string) (*github.Milestone, error) {
	page := 1
	repoElems := strings.SplitN(repo, "/", 2)
	if len(repoElems) != 2 {
		return nil, fmt.Errorf("invalid repo provided: %s", repo)
	}
	owner := repoElems[0]
	repo = repoElems[1]
	for page > 0 {
		opts := github.MilestoneListOptions{State: "all"}
		opts.PerPage = 100
		opts.Page = page
		tk.IncrRequestCount()
		milestones, resp, err := tk.GitHubClient().Issues.ListMilestones(ctx, owner, repo, &opts)
		if err != nil {
			return nil, err
		}
		for _, ms := range milestones {
			if ms.GetTitle() == version {
				return ms, nil
			}
		}
		page = resp.NextPage
	}
	return nil, nil
}
