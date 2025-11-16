package usecase

import (
	"context"
	"gopr/cmd/config"
	"gopr/internal/repo/pg"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Cases struct {
	Team        *Team
	User        *User
	PullRequest *PullRequest
}

func Setup(ctx context.Context, cfg *config.Config, db *pgxpool.Pool) Cases {
	teamRepo := pg.NewTeamRepo(db)
	userRepo := pg.NewUserRepo(db)
	prRepo := pg.NewPullRequestRepo(db)

	return Cases{
		Team:        NewTeam(teamRepo, userRepo),
		User:        NewUser(userRepo, teamRepo),
		PullRequest: NewPullRequest(prRepo, userRepo),
	}
}
