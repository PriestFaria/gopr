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

type PullRequestRepo struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewPullRequestRepo(db *pgxpool.Pool) *PullRequestRepo {
	return &PullRequestRepo{
		db:   db,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *PullRequestRepo) Create(ctx context.Context, pr *domain.PullRequest) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO pull_request(id, author_id, name, status, created_at)
         VALUES ($1, $2, $3, $4, NOW())`,
		pr.Id,
		pr.AuthorId,
		pr.Name,
		pr.Status,
	)
	if err != nil {
		return fmt.Errorf("insert pull_request: %w", err)
	}
	return nil
}

func (r *PullRequestRepo) GetByID(ctx context.Context, id string) (*domain.PullRequest, error) {
	var pr domain.PullRequest

	err := r.db.QueryRow(ctx,
		`SELECT id, author_id, name, status, created_at, merged_at
         FROM pull_request
         WHERE id = $1`,
		id,
	).Scan(
		&pr.Id,
		&pr.AuthorId,
		&pr.Name,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select pull_request: %w", err)
	}

	return &pr, nil
}

func (r *PullRequestRepo) UpdateStatusMerged(ctx context.Context, id string) error {
	res, err := r.db.Exec(ctx,
		`UPDATE pull_request
         SET status = 'merged',
             merged_at = COALESCE(merged_at, NOW())
         WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("update merged: %w", err)
	}

	if res.RowsAffected() == 0 {
		return repo.ErrNotFound
	}

	return nil
}

func (r *PullRequestRepo) AddReviewer(ctx context.Context, prID, reviewerID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO pull_request_reviewer(pull_request_id, reviewer_id)
         VALUES ($1, $2)`,
		prID, reviewerID,
	)
	if err != nil {
		return fmt.Errorf("insert reviewer: %w", err)
	}
	return nil
}

func (r *PullRequestRepo) RemoveReviewer(ctx context.Context, prID, reviewerID string) error {
	res, err := r.db.Exec(ctx,
		`DELETE FROM pull_request_reviewer
         WHERE pull_request_id = $1 AND reviewer_id = $2`,
		prID, reviewerID,
	)
	if err != nil {
		return fmt.Errorf("delete reviewer: %w", err)
	}

	if res.RowsAffected() == 0 {
		return repo.ErrNotFound
	}

	return nil
}

func (r *PullRequestRepo) ListReviewers(ctx context.Context, prID string) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT reviewer_id
         FROM pull_request_reviewer
         WHERE pull_request_id = $1`,
		prID,
	)
	if err != nil {
		return nil, fmt.Errorf("query reviewers: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			return nil, fmt.Errorf("scan reviewer: %w", err)
		}
		ids = append(ids, rid)
	}

	return ids, nil
}

func (r *PullRequestRepo) ListByReviewer(ctx context.Context, reviewerID string) ([]*domain.PullRequest, error) {
	sql, args, err := r.psql.
		Select(
			"pr.id",
			"pr.author_id",
			"pr.name",
			"pr.status",
			"pr.created_at",
			"pr.merged_at",
		).
		From("pull_request AS pr").
		Join("pull_request_reviewer AS rr ON rr.pull_request_id = pr.id").
		Where(sq.Eq{"rr.reviewer_id": reviewerID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql listByReviewer: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query listByReviewer: %w", err)
	}
	defer rows.Close()

	var res []*domain.PullRequest

	for rows.Next() {
		var pr domain.PullRequest
		err := rows.Scan(
			&pr.Id,
			&pr.AuthorId,
			&pr.Name,
			&pr.Status,
			&pr.CreatedAt,
			&pr.MergedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan listByReviewer: %w", err)
		}
		res = append(res, &pr)
	}

	return res, nil
}
