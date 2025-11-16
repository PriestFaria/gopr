package repo

import (
	"context"
	"gopr/internal/domain"
)

type User interface {
	Create(ctx context.Context, user *domain.CreateUser) (string, error)
	SetIsActive(ctx context.Context, isActive bool) error
}

type Team interface {
	Create(ctx context.Context, createTeam *domain.CreateTeam) (string, error)
}

type PullRequest interface {
	Create(ctx context.Context, createPullRequest *domain.CreatePullRequest) (string, error)
	Reassign(ctx context.Context, reassignPullRequest *domain.ReassignPullRequest) (string, error)
	Merge(ctx context.Context, mergePullRequest *domain.MergePullRequest) (string, error)
}
