package domain

import "time"

type PullRequestStatus string

var (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusClosed PullRequestStatus = "MERGED"
)

type PullRequest struct {
	Id       string `json:"id"`
	AuthorId string `json:"author_id"`
	Name     string `json:"name"`

	Status string `json:"status"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	MergedAt  *time.Time `json:"merged_at"`
}

type CreatePullRequest struct {
	Id       string `json:"pull_request_id"`
	AuthorId string `json:"author_id"`
	Name     string `json:"pull_request_name"`
}

type ReassignPullRequest struct {
	Id            string `json:"pull_request_id"`
	OldReviewerId string `json:"old_reviewer_id"`
}

type MergePullRequest struct {
	Id string `json:"pull_request_id"`
}

type PullRequestReassignResponse struct {
	PullRequest *PullRequest `json:"pull_request"`
	NewReviewer string       `json:"new_reviewer"`
}
