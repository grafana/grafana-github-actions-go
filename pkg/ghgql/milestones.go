package ghgql

import "context"

type Milestone struct {
	Number int
	Title  string
	Closed bool
}

func (m Milestone) String() string {
	return m.Title
}

func (c *Client) GetMilestoneByTitle(ctx context.Context, repoOwner string, repoName string, title string) (*Milestone, error) {
	resp, err := getMilestonesWithTitle(ctx, c.gql, repoOwner, repoName, title)
	if err != nil {
		return nil, err
	}
	if resp == nil {
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
	return nil, nil
}
