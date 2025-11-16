package pg

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}
