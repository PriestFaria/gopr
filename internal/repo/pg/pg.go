package pg

import "gopr/internal/repo"

var (
	_ repo.User        = &UserRepo{}
	_ repo.Team        = &TeamRepo{}
	_ repo.PullRequest = &PullRequestRepo{}
)
