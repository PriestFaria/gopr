package usecase

import (
	"gopr/internal/repo"
)

type Cases struct {
	Team        *Team
	User        *User
	PullRequest *PullRequest
}

func Setup(
	teamRepo repo.Team,
	userRepo repo.User,
	prRepo repo.PullRequest,
) Cases {
	teamCase := NewTeam(teamRepo, userRepo)
	userCase := NewUser(userRepo)
	prCase := NewPullRequest(prRepo, userRepo)

	return Cases{
		Team:        teamCase,
		User:        userCase,
		PullRequest: prCase,
	}
}
