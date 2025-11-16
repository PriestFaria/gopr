package pg

import (
	"context"
	"errors"
	"fmt"
	"gopr/internal/domain"
	"gopr/internal/repo"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		db:   db,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO "users"(id, username, team_id, is_active, created_at, updated_at)
         VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		user.Id,
		user.Username,
		user.TeamId,
		user.IsActive,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User

	err := r.db.QueryRow(ctx,
		`SELECT id, username, team_id, is_active, created_at, updated_at
         FROM "users"
         WHERE id = $1`,
		id,
	).Scan(
		&u.Id,
		&u.Username,
		&u.TeamId,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user: %w", err)
	}

	return &u, nil
}

func (r *UserRepo) UpdateIsActive(ctx context.Context, id string, isActive bool) error {
	res, err := r.db.Exec(ctx,
		`UPDATE "users"
         SET is_active = $1, updated_at = NOW()
         WHERE id = $2`,
		isActive,
		id,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if res.RowsAffected() == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *UserRepo) ListByTeam(ctx context.Context, teamID string, onlyActive bool) ([]*domain.User, error) {
	builder := r.psql.
		Select("id", "username", "team_id", "is_active", "created_at", "updated_at").
		From(`"users"`).
		Where(sq.Eq{"team_id": teamID})

	if onlyActive {
		builder = builder.Where(sq.Eq{"is_active": true})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql (list users by team): %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query users by team: %w", err)
	}
	defer rows.Close()

	var result []*domain.User

	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.Id,
			&u.Username,
			&u.TeamId,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		result = append(result, &u)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}

	return result, nil
}
