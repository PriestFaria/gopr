package domain

import "time"

type PullRequestStatus string

var (
	PullRequestStatusOpen   PullRequestStatus = "open"
	PullRequestStatusClosed PullRequestStatus = "closed"
)

type PullRequest struct {
	Id       string `json:"id"`
	AuthorId string `json:"author_id"`
	Name     string `json:"name"`

	Status string `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	MergedAt  time.Time `json:"merged_at"`
}
