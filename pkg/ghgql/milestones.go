package ghgql

import (
	"context"

	"github.com/rs/zerolog"
)

type Milestone struct {
	Number int
	Title  string
	Closed bool
}

func (m Milestone) String() string {
	return m.Title
}

func (c *Client) GetMilestoneByTitle(ctx context.Context, repoOwner string, repoName string, title string) (*Milestone, error) {
	logger := zerolog.Ctx(ctx)
	resp, err := getMilestonesWithTitle(ctx, c.gql, repoOwner, repoName, title)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		logger.Warn().Msgf("No response for milestone with title `%s`", title)
		return nil, nil
	}
	for _, candidate := range resp.GetRepository().Milestones.Nodes {
		if title == candidate.GetTitle() {
			return &Milestone{
				Number: candidate.Number,
				Title:  candidate.Title,
				Closed: candidate.Closed,
			}, nil
		}
	}
	logger.Warn().Msgf("No milestone with title `%s` found", title)
	return nil, nil
}
