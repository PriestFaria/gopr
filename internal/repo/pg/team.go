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

type TeamRepo struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewTeamRepo(db *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{
		db:   db,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *TeamRepo) Create(ctx context.Context, team *domain.Team) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO team(id, name)
         VALUES ($1, $2)`,
		team.Id,
		team.Name,
	)
	if err != nil {
		return fmt.Errorf("insert team: %w", err)
	}
	return nil
}

func (r *TeamRepo) GetByID(ctx context.Context, id string) (*domain.Team, error) {
	var t domain.Team

	err := r.db.QueryRow(ctx,
		`SELECT id, name
         FROM team
         WHERE id = $1`,
		id,
	).Scan(&t.Id, &t.Name)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select team by id: %w", err)
	}

	return &t, nil
}

func (r *TeamRepo) GetByName(ctx context.Context, name string) (*domain.Team, error) {
	var t domain.Team

	err := r.db.QueryRow(ctx,
		`SELECT id, name
         FROM team
         WHERE name = $1`,
		name,
	).Scan(&t.Id, &t.Name)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select team by name: %w", err)
	}

	return &t, nil
}
