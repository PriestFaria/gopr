package pg

import (
	"context"
	"gopr/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PullRequestRepo struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func (p *PullRequestRepo) Create(ctx context.Context, createPullRequest *domain.CreatePullRequest) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PullRequestRepo) Reassign(ctx context.Context, reassignPullRequest *domain.ReassignPullRequest) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PullRequestRepo) Merge(ctx context.Context, mergePullRequest *domain.MergePullRequest) (string, error) {
	//TODO implement me
	panic("implement me")
}
