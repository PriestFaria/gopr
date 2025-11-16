package pg

import (
	"context"
	"gopr/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepo struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func (t *TeamRepo) Create(ctx context.Context, createTeam *domain.CreateTeam) (string, error) {
	//TODO implement me
	panic("implement me")
}
