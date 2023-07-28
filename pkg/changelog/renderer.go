package changelog

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/grafana/grafana-github-actions-go/pkg/ghgql"
	"github.com/grafana/grafana-github-actions-go/pkg/toolkit"
	"github.com/rs/zerolog"
)

// Renderer converts a changelog into a string that can then be posted e.g. to
// Discourse or to the main changelog file.
type Renderer interface {
	Render(context.Context, *ChangelogBody) (string, error)
}

// NewRenderer returns a renderer that produces Markdown as used by Discourse
// and the changelog.
func NewRenderer(tk *toolkit.Toolkit) Renderer {
	return &defaultRenderer{
		tk: tk,
	}
}

type defaultRenderer struct {
	tk *toolkit.Toolkit
}

func (r *defaultRenderer) Render(ctx context.Context, body *ChangelogBody) (string, error) {
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
		r.writeIssueLines(&out, body.Features)
		out.WriteString("\n")
	}
	if len(body.Bugfixes) > 0 {
		out.WriteString("### Bug fixes\n\n")
		r.writeIssueLines(&out, body.Bugfixes)
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
		r.writeIssueLines(&out, body.PluginDevChanges)
		out.WriteString("\n")
	}
	return out.String(), nil
}

func (r *defaultRenderer) writeIssueLines(out *strings.Builder, issues []ghgql.PullRequest) {
	for _, issue := range issues {
		out.WriteString(r.issueAsMarkdown(issue))
	}
}

var titleHeadlinePattern = regexp.MustCompile(`^([^:]*:)`)

func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// PreparePRTitle converts the title of the pull-request into a format that
// will then be used for rendering it. Since the output of this function can be
// used to match PRs from various releases it is public.
func PreparePRTitle(issue ghgql.PullRequest) string {
	out := strings.Builder{}
	title := issue.GetTitle()
	title = stripReleaseStreamPrefix(title)
	title = strings.TrimSuffix(title, ".")
	title = escapeMarkdown(title)
	out.WriteString(title)
	if issueHasLabel(issue, LabelEnterprise) || issue.GetRepoName() == "grafana-enterprise" {
		out.WriteString(". (Enterprise)")
	} else {
		out.WriteString(". ")
	}
	return out.String()
}

func (r *defaultRenderer) issueAsMarkdown(issue ghgql.PullRequest) string {
	ctx := context.Background()
	out := strings.Builder{}

	title := PreparePRTitle(issue)
	title = titleHeadlinePattern.ReplaceAllString(title, "**$1**")

	out.WriteString("- ")
	out.WriteString(title)
	if issueHasLabel(issue, LabelEnterprise) || issue.GetRepoName() == "grafana-enterprise" {
	} else {
		out.WriteString(r.getIssueLink(issue))
		if issue.GetAuthorLogin() != "" {
			userLink, err := r.getUserLink(ctx, issue)
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

func (r *defaultRenderer) getIssueLink(issue ghgql.PullRequest) string {
	return getIssueLink(issue)
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

func isBotUser(issue ghgql.PullRequest) bool {
	if strings.HasPrefix(issue.GetAuthorResourcePath(), "/apps/") {
		return true
	}
	switch issue.GetAuthorLogin() {
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

func (r *defaultRenderer) getUserLink(ctx context.Context, issue ghgql.PullRequest) (string, error) {
	logger := zerolog.Ctx(ctx)
	user := issue.GetAuthorLogin()
	if isBotUser(issue) {
		logger.Info().Msgf("PR#%d created by bot. Fetching original author from %s", issue.GetNumber(), issue.GetHeadRefName())
		// If this looks like a bot user, take the author of the original PR if
		// available:
		origPrNumber, err := getPRNumberFromBackportBranch(issue.GetHeadRefName())
		if err != nil {
			return "", err
		}
		origPR, _, err := r.tk.GitHubClient().PullRequests.Get(context.Background(), issue.GetRepoOwner(), issue.GetRepoName(), origPrNumber)
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
