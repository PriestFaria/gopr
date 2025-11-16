package repo

import (
	"context"
	"gopr/internal/domain"
)

type User interface {
	Create(ctx context.Context, user *domain.User) error

	GetByID(ctx context.Context, id string) (*domain.User, error)

	UpdateIsActive(ctx context.Context, id string, isActive bool) error

	ListByTeam(ctx context.Context, teamID string, onlyActive bool) ([]*domain.User, error)
}

type Team interface {
	Create(ctx context.Context, team *domain.Team) error

	GetByID(ctx context.Context, id string) (*domain.Team, error)

	GetByName(ctx context.Context, name string) (*domain.Team, error)
}

type PullRequest interface {
	Create(ctx context.Context, pr *domain.PullRequest) error

	GetByID(ctx context.Context, id string) (*domain.PullRequest, error)

	UpdateStatusMerged(ctx context.Context, id string) error

	AddReviewer(ctx context.Context, prID, reviewerID string) error
	RemoveReviewer(ctx context.Context, prID, reviewerID string) error
	ListReviewers(ctx context.Context, prID string) ([]string, error)

	ListByReviewer(ctx context.Context, reviewerID string) ([]*domain.PullRequest, error)
}
