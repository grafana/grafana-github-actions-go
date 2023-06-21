package ghgql

import (
	"context"
)

type PullRequest struct {
	Number      *int
	Title       *string
	Body        *string
	Labels      []string
	AuthorLogin *string
	RepoOwner   *string
	RepoName    *string
	HeadRefName *string
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
func (c *Client) GetMilestonedPRsForChangelog(ctx context.Context, repoOwner string, repoName string, milestoneNumber int) ([]PullRequest, error) {
	cursor := ""
	result := make([]PullRequest, 0, 30)
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
			number := pr.Number
			title := pr.Title
			body := pr.Body
			headRefName := pr.HeadRefName
			r := PullRequest{
				Number:      &number,
				Title:       &title,
				Body:        &body,
				Labels:      labels,
				RepoName:    &repoName,
				RepoOwner:   &repoOwner,
				AuthorLogin: &author,
				HeadRefName: &headRefName,
			}
			result = append(result, r)
		}
		if !pageInfo.HasNextPage {
			break
		}
		cursor = pageInfo.EndCursor
	}
	return result, nil
}
