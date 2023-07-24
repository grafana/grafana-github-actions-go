package changelog

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/grafana/grafana-github-actions-go/pkg/ghgql"
	"github.com/grafana/grafana-github-actions-go/pkg/toolkit"

	"github.com/google/go-github/v50/github"
)

const LabelEnterprise = "enterprise"
const LabelUI = "area/grafana/ui"
const LabelToolkit = "area/grafana/toolkit"
const LabelRuntime = "area/grafana/runtime"
const LabelBug = "type/bug"

func Build(ctx context.Context, version string, tk *toolkit.Toolkit) (*ChangelogBody, error) {
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
	milestones, err := getHistoricalMilestones(ctx, tk, version)
	if err != nil {
		return nil, err
	}

	knownTitles := make(map[string]struct{})

	loader := NewLoader(tk.GitHubClient())
	parser := NewParser()
	for _, milestone := range milestones {
		msContent, err := loader.LoadContent(ctx, "grafana", "grafana", milestone.GetTitle(), &LoaderOptions{RemoveHeading: true})
		if err != nil {
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
	for _, i := range issues {
		// If the PR already seems to be present in a previous release, we can
		// skip it here:
		newTitle := PreparePRTitle(i)
		if _, found := knownTitles[newTitle]; found {
			continue
		}
		addToBody(body, i)
	}
	return body, nil
}

// getHistoricalMilestones retrieves all the milestones of the current and
// previous release stream that were closed n-days before the milestone
// matching `version`.
func getHistoricalMilestones(ctx context.Context, tk *toolkit.Toolkit, version string) ([]github.Milestone, error) {
	// TODO: Provide implementation
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
