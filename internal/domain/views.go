package domain

type TeamWithMembers struct {
	Team    *Team   `json:"team"`
	Members []*User `json:"members"`
}

type PullRequestWithReviewers struct {
	PR        *PullRequest `json:"pr"`
	Reviewers []string     `json:"assigned_reviewers"`
}

type UserReviews struct {
	UserID string         `json:"user_id"`
	PRs    []*PullRequest `json:"pull_requests"`
}
