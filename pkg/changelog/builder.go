package changelog

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
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

func (body *ChangelogBody) ToMarkdown(tk *toolkit.Toolkit) string {
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
		writeIssueLines(&out, tk, body.Features)
		out.WriteString("\n")
	}
	if len(body.Bugfixes) > 0 {
		out.WriteString("### Bug fixes\n\n")
		writeIssueLines(&out, tk, body.Bugfixes)
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
		writeIssueLines(&out, tk, body.PluginDevChanges)
		out.WriteString("\n")
	}
	return out.String()
}

func writeIssueLines(out *strings.Builder, tk *toolkit.Toolkit, issues []ghgql.PullRequest) {
	for _, issue := range issues {
		out.WriteString(issueAsMarkdown(issue, tk))
	}
}

var titleHeadlinePattern = regexp.MustCompile(`^([^:]*:)`)

func issueAsMarkdown(issue ghgql.PullRequest, tk *toolkit.Toolkit) string {
	ctx := context.Background()
	out := strings.Builder{}

	title := issue.GetTitle()
	title = stripReleaseStreamPrefix(title)
	title = titleHeadlinePattern.ReplaceAllString(title, "**$1**")
	title = strings.TrimSuffix(title, ".")

	out.WriteString("- ")
	out.WriteString(title)
	if issueHasLabel(issue, LabelEnterprise) || issue.GetRepoName() == "grafana-enterprise" {
		out.WriteString(". (Enterprise)")
	} else {
		out.WriteString(". ")
		out.WriteString(getIssueLink(issue))
		if issue.GetAuthorLogin() != "" {
			userLink, err := getUserLink(ctx, issue, tk)
			if err != nil {
			} else {
				out.WriteString(", ")
				out.WriteString(userLink)
			}
		}
	}
	out.WriteString("\n")
	return out.String()
}

var releaseStreamPrefixPattern = regexp.MustCompile(`^(\[[^]]+\]:?) (.*)$`)

func stripReleaseStreamPrefix(input string) string {
	if releaseStreamPrefixPattern.MatchString(input) {
		result := releaseStreamPrefixPattern.FindStringSubmatch(input)
		return result[2]
	}
	return input
}

func getIssueLink(issue ghgql.PullRequest) string {
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

func isBotUser(login string) bool {
	switch login {
	case "grafanabot":
		return true
	default:
		return false
	}
}

func getPRNumberFromBackportBranch(ref string) (int, error) {
	pat := regexp.MustCompile("^backport-(\\d+)-to-v\\d+\\.\\d+\\.x$")
	match := pat.FindStringSubmatch(ref)
	if len(match) < 1 {
		return -1, fmt.Errorf("no number found in ref")
	}
	result, err := strconv.ParseInt(match[1], 10, 64)
	return int(result), err

}

func getUserLink(ctx context.Context, issue ghgql.PullRequest, tk *toolkit.Toolkit) (string, error) {
	logger := zerolog.Ctx(ctx)
	user := issue.GetAuthorLogin()
	if isBotUser(user) {
		logger.Info().Msgf("PR#%d created by bot. Fetching original author from %s", issue.GetNumber(), issue.GetHeadRefName())
		// If this looks like a bot user, take the author of the original PR if
		// available:
		origPrNumber, err := getPRNumberFromBackportBranch(issue.GetHeadRefName())
		if err != nil {
			return "", err
		}
		origPR, _, err := tk.GitHubClient().PullRequests.Get(context.Background(), issue.GetRepoOwner(), issue.GetRepoName(), origPrNumber)
		if err != nil {
			return "", err
		}
		user = origPR.User.GetLogin()
	}
	out := strings.Builder{}
	out.WriteString("[@")
	out.WriteString(user)
	out.WriteString("]")
	out.WriteString("(https://github.com/")
	out.WriteString(user)
	out.WriteString(")")
	return out.String(), nil
}

func issueHasLabel(issue ghgql.PullRequest, label string) bool {
	for _, l := range issue.Labels {
		if l == label {
			return true
		}
	}
	return false
}

func getBreakingChangeNotice(issue ghgql.PullRequest) string {
	return getNotice(issue, "Release notice breaking change")
}

func getDeprecationNotice(issue ghgql.PullRequest) string {
	return getNotice(issue, "Deprecation notice")
}

func getNotice(issue ghgql.PullRequest, sectionStart string) string {
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
				// Trim tailing whitespaces before finalizing the output:
				output := strings.Builder{}
				output.WriteString(strings.TrimSpace(result.String()))
				result = strings.Builder{}
				result.WriteString(output.String())

				if strings.HasSuffix(output.String(), "```") {
					result.WriteString("\n")
				} else {
					result.WriteString(" ")
				}
				result.WriteString("Issue ")
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
