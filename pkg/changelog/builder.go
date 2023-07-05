package changelog

import (
	"context"
	"fmt"
	"strings"

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

type Entry struct {
	Title                     string
	PullRequestNumber         int
	OriginalPullRequestNumber int
	Labels                    []string
}

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

	filter := func(pr *ghgql.PullRequest) bool {
		for _, l := range pr.Labels {
			if l == "backport" || l == "no-backport" || l == "backport v10.0.x" {
				return true
			}
		}
		return false
	}

	ossIssues, err := tk.GitHubGQLClient().GetMilestonedPRsForChangelog(ctx, "grafana", "grafana", milestone, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve OSS issues: %w", err)
	}

	enterpriseIssues, err := tk.GitHubGQLClient().GetMilestonedPRsForChangelog(ctx, "grafana", "grafana-enterprise", enterpriseMilestone, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Enterprise issues: %w", err)
	}

	issues := make([]ghgql.PullRequest, 0, len(ossIssues)+len(enterpriseIssues))
	issues = append(issues, ossIssues...)
	issues = append(issues, enterpriseIssues...)

	logger.Info().Msgf("%d issues in total", len(issues))

	body.Version = version
	if !milestone.GetDueOn().IsZero() {
		body.ReleaseDate = milestone.GetDueOn().Format("2006-01-02")
	} else if !milestone.GetClosedAt().IsZero() {
		body.ReleaseDate = milestone.GetClosedAt().Format("2006-01-02")
	}
	for _, i := range issues {
		addToBody(body, i)
	}
	return body, nil
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
