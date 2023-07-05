package types

import "time"

type GQLMilestone struct {
	Title    string     `json:"title"`
	Number   int        `json:"number"`
	ClosedAt *time.Time `json:"closedAt"`
	DueOn    *time.Time `json:"dueOn"`
}
