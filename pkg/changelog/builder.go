package changelog

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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

	ossIssues, err := getIssues(ctx, tk, "grafana/grafana", version)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve OSS issues: %w", err)
	}

	enterpriseIssues, err := getIssues(ctx, tk, "grafana/grafana-enterprise", version)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Enterprise issues: %w", err)
	}

	issues := make([]*github.Issue, 0, len(ossIssues)+len(enterpriseIssues))
	issues = append(issues, ossIssues...)
	issues = append(issues, enterpriseIssues...)

	body.Version = version
	if !milestone.GetClosedAt().IsZero() {
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
	PluginDevChanges   []*github.Issue
	Bugfixes           []*github.Issue
	Features           []*github.Issue
}

func (body *ChangelogBody) ToMarkdown() string {
	out := strings.Builder{}
	out.WriteString("# ")
	out.WriteString(body.Version)
	if body.ReleaseDate != "" {
		out.WriteString(" (")
		out.WriteString(body.ReleaseDate)
		out.WriteString(")")
	}
	out.WriteString("\n\n")
	if len(body.Features) > 0 {
		out.WriteString("### Features and enhancements\n\n")
		writeIssueLines(&out, body.Features)
		out.WriteString("\n")
	}
	if len(body.Bugfixes) > 0 {
		out.WriteString("### Bug fixes\n\n")
		writeIssueLines(&out, body.Bugfixes)
		out.WriteString("\n")
	}
	if len(body.BreakingChanges) > 0 {
		out.WriteString("### Breaking changes\n\n")
		for _, notice := range body.BreakingChanges {
			out.WriteString(notice)
			out.WriteString("\n\n")
		}
	}
	if len(body.DeprecationChanges) > 0 {
		out.WriteString("### Deprecations\n\n")
		for _, notice := range body.DeprecationChanges {
			out.WriteString(notice)
			out.WriteString("\n\n")
		}
	}
	if len(body.PluginDevChanges) > 0 {
		out.WriteString("### Plugin development fixes & changes\n\n")
		writeIssueLines(&out, body.PluginDevChanges)
		out.WriteString("\n")
	}
	return out.String()
}

func writeIssueLines(out *strings.Builder, issues []*github.Issue) {
	for _, issue := range issues {
		out.WriteString(issueAsMarkdown(issue))
	}
}

var titleHeadlinePattern = regexp.MustCompile(`^([^:]*:)`)

func issueAsMarkdown(issue *github.Issue) string {
	out := strings.Builder{}

	title := issue.GetTitle()
	title = titleHeadlinePattern.ReplaceAllString(title, "**$1**")
	title = strings.TrimSuffix(title, ".")

	out.WriteString("- ")
	out.WriteString(title)
	if issueHasLabel(issue, LabelEnterprise) || strings.HasSuffix(issue.GetRepositoryURL(), "grafana-enterprise") {
		out.WriteString(". (Enterprise)")
	} else {
		out.WriteString(". ")
		out.WriteString(getIssueLink(issue))
		if issue.IsPullRequest() && issue.User != nil {
			out.WriteString(", ")
			out.WriteString(getUserLink(issue.User))
		}
	}
	out.WriteString("\n")
	return out.String()
}

func getIssueLink(issue *github.Issue) string {
	num := strconv.Itoa(issue.GetNumber())
	out := strings.Builder{}
	out.WriteString("[#")
	out.WriteString(num)
	out.WriteString("]")
	out.WriteString("(https://github.com/grafana/grafana/issues/")
	out.WriteString(num)
	out.WriteString(")")
	return out.String()
}

func getUserLink(user *github.User) string {
	out := strings.Builder{}
	out.WriteString("[@")
	out.WriteString(user.GetLogin())
	out.WriteString("]")
	out.WriteString("(https://github.com/")
	out.WriteString(user.GetLogin())
	out.WriteString(")")
	return out.String()
}

func issueHasLabel(issue *github.Issue, label string) bool {
	if issue == nil || issue.Labels == nil {
		return false
	}
	for _, l := range issue.Labels {
		if l.GetName() == label {
			return true
		}
	}
	return false
}

func getBreakingChangeNotice(issue *github.Issue) string {
	return getNotice(issue, "Release notice breaking change")
}

func getDeprecationNotice(issue *github.Issue) string {
	return getNotice(issue, "Deprecation notice")
}

func getNotice(issue *github.Issue, sectionStart string) string {
	lines := strings.Split(issue.GetBody(), "\n")
	startFound := false
	result := strings.Builder{}
	lastLine := len(lines) - 1
	for idx, line := range lines {
		if startFound {
			l := strings.TrimSpace(line)
			if result.Len() > 0 {
				result.WriteString("\n")
			} else {
				// If there is a blank line right after the start, let's skip it:
				if l == "" {
					continue
				}
			}
			result.WriteString(l)
			if idx == lastLine {
				result.WriteString(" Issue ")
				result.WriteString(getIssueLink(issue))
			}
		}
		if strings.Contains(line, sectionStart) {
			startFound = true
		}
	}
	return result.String()
}

func newChangelogBody() *ChangelogBody {
	return &ChangelogBody{
		DeprecationChanges: make([]string, 0, 10),
		BreakingChanges:    make([]string, 0, 10),
		PluginDevChanges:   make([]*github.Issue, 0, 10),
		Bugfixes:           make([]*github.Issue, 0, 10),
		Features:           make([]*github.Issue, 0, 10),
	}
}

func addToBody(body *ChangelogBody, issue *github.Issue) {
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

func isBug(issue *github.Issue) bool {
	title := issue.GetTitle()
	if strings.Contains(strings.ToLower(title), "fix") {
		return true
	}
	if issueHasLabel(issue, LabelBug) {
		return true
	}
	return false
}

func getIssues(ctx context.Context, tk *toolkit.Toolkit, repo string, version string) ([]*github.Issue, error) {
	result := make([]*github.Issue, 0, 10)
	nextPage := 1
	for nextPage > 0 {
		opts := &github.SearchOptions{}
		opts.Page = nextPage
		tk.IncrRequestCount()
		issues, resp, err := tk.GitHubClient().Search.Issues(ctx, fmt.Sprintf(`repo:%s label:"add to changelog" is:pr is:closed milestone:%s`, repo, version), opts)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code from GitHub API: %s", resp.Status)
		}
		result = append(result, issues.Issues...)
		nextPage = resp.NextPage
	}
	return result, nil
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
		opts := github.MilestoneListOptions{}
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
