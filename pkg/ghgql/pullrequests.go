package ghgql

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghgql/types"
	"github.com/rs/zerolog"
)

type PullRequest struct {
	Number             *int
	Title              *string
	Body               *string
	Labels             []string
	AuthorResourcePath *string
	AuthorLogin        *string
	RepoOwner          *string
	RepoName           *string
	HeadRefName        *string
	Milestone          *types.GQLMilestone
}

func (pr *PullRequest) IsBackport() bool {
	return hasLabel(pr, "backport")
}

func (pr *PullRequest) GetOriginalNumber() *int {
	pat := regexp.MustCompile("^backport-(\\d+)-to-v\\d+\\.\\d+\\.x$")
	match := pat.FindStringSubmatch(pr.GetHeadRefName())
	if len(match) < 1 {
		return nil
	}
	result, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return nil
	}
	res := int(result)
	return &res
}

func (pr *PullRequest) GetNumber() int {
	return *pr.Number
}

func (pr *PullRequest) GetTitle() string {
	if pr.Title == nil {
		return ""
	}
	return *pr.Title
}
func (pr *PullRequest) GetBody() string {
	if pr.Body == nil {
		return ""
	}
	return *pr.Body
}

func (pr *PullRequest) GetAuthorLogin() string {
	if pr.AuthorLogin == nil {
		return ""
	}
	return *pr.AuthorLogin
}

func (pr *PullRequest) GetAuthorResourcePath() string {
	if pr.AuthorResourcePath == nil {
		return ""
	}
	return *pr.AuthorResourcePath
}

func (pr *PullRequest) GetRepoOwner() string {
	if pr.RepoOwner == nil {
		return ""
	}
	return *pr.RepoOwner
}

func (pr *PullRequest) GetRepoName() string {
	if pr.RepoName == nil {
		return ""
	}
	return *pr.RepoName
}

func (pr *PullRequest) GetHeadRefName() string {
	if pr.HeadRefName == nil {
		return ""
	}
	return *pr.HeadRefName
}

type PullRequestFilterFunc func(pr *PullRequest) bool

func hasLabel(pr *PullRequest, label string) bool {
	for _, l := range pr.Labels {
		if l == label {
			return true
		}
	}
	return false
}

func (c *Client) GetMilestonedPRsForChangelog(ctx context.Context, repoOwner string, repoName string, milestone *github.Milestone, filter PullRequestFilterFunc) ([]PullRequest, error) {
	logger := zerolog.Ctx(ctx)
	cursor := ""
	directPRs := make([]PullRequest, 0, 30)
	relevantPRs := make(map[int]PullRequest)
	milestoneNumber := milestone.GetNumber()

	// First fetch all the PRs that are directly attached to this milestone:
	for {
		resp, err := getMilestonedPullRequests(ctx, c.gql, repoOwner, repoName, milestoneNumber, cursor)
		if err != nil {
			return nil, err
		}
		pageInfo := resp.Repository.Milestone.PullRequests.PageInfo
		for _, pr := range resp.Repository.Milestone.PullRequests.Nodes {
			labels := make([]string, len(pr.Labels.Nodes))
			for _, l := range pr.Labels.Nodes {
				labels = append(labels, l.Name)
			}
			author := pr.Author.GetLogin()
			authorResourcePath := pr.Author.GetResourcePath()
			number := pr.Number
			title := pr.Title
			body := pr.Body
			headRefName := pr.HeadRefName
			r := PullRequest{
				Number:             &number,
				Title:              &title,
				Body:               &body,
				Labels:             labels,
				RepoName:           &repoName,
				RepoOwner:          &repoOwner,
				AuthorLogin:        &author,
				AuthorResourcePath: &authorResourcePath,
				HeadRefName:        &headRefName,
				Milestone:          pr.Milestone,
			}
			fmt.Println(pr.Milestone)
			directPRs = append(directPRs, r)
			relevantPRs[number] = r
		}
		if !pageInfo.HasNextPage {
			break
		}
		cursor = pageInfo.EndCursor
	}

	// previousMinor = "9.5.x"

	// From these, let's find all the PRs that are backports and get their original PRs:
	for _, pr := range directPRs {
		if pr.IsBackport() {
			logger.Info().Msgf("Fetching original PR for #%d", pr.GetNumber())
			origPRNumber := pr.GetOriginalNumber()
			if origPRNumber == nil {
				logger.Warn().Msgf("Failed to determine original PR of #%d", pr.GetNumber())
				continue
			}
			// TODO: Get this single PR
			_, err := c.fetchSinglePR(ctx, repoOwner, repoName, *origPRNumber, relevantPRs)
			if err != nil {
				logger.Warn().Err(err).Msgf("Failed to fetch original PR of #%d", pr.GetNumber())
				continue
			}

			// Now also fetch the backport for the previous minor
			//
			// The problem is that we cannot easily get from a PR to its
			// backports except through the branch name and that might not
			// return a unique result:

		} else {
			if !hasLabel(&pr, "no-backport") {
				// If this wants to backport to at least one version, what versions are those?
				fmt.Println(pr.Labels)
			}
		}
	}
	sort.Slice(directPRs, func(i, j int) bool {
		a := directPRs[i]
		b := directPRs[j]
		return b.GetNumber() < a.GetNumber()
	})
	return directPRs, nil
}

func (c *Client) fetchSinglePR(ctx context.Context, repoOwner string, repoName string, prNumber int, knownPRs map[int]PullRequest) (PullRequest, error) {
	resp, err := getPullRequest(ctx, c.gql, repoOwner, repoName, prNumber)
	if err != nil {
		return PullRequest{}, fmt.Errorf("failed to get #%d: %w", prNumber, err)
	}
	origPR := resp.GetRepository().PullRequest
	labels := make([]string, len(origPR.Labels.Nodes))
	for _, l := range origPR.Labels.Nodes {
		labels = append(labels, l.Name)
	}
	author := origPR.Author.GetLogin()
	authorResourcePath := origPR.Author.GetResourcePath()
	number := origPR.Number
	title := origPR.Title
	body := origPR.Body
	headRefName := origPR.HeadRefName
	r := PullRequest{
		Number:             &number,
		Title:              &title,
		Body:               &body,
		Labels:             labels,
		RepoName:           &repoName,
		RepoOwner:          &repoOwner,
		AuthorLogin:        &author,
		AuthorResourcePath: &authorResourcePath,
		HeadRefName:        &headRefName,
		Milestone:          origPR.Milestone,
	}
	fmt.Println(origPR.Milestone)
	knownPRs[number] = r
	return r, nil
}
