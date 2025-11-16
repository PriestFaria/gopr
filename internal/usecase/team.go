package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"gopr/internal/domain"
	"gopr/internal/repo"
)

type Team struct {
	teamRepo repo.Team
	userRepo repo.User
}

func NewTeam(teamRepo repo.Team, userRepo repo.User) *Team {
	return &Team{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (t *Team) AddTeam(ctx context.Context, input *domain.TeamAddInput) (*domain.TeamWithMembers, error) {
	team := &domain.Team{
		Id:   uuid.NewString(),
		Name: input.TeamName,
	}

	if err := t.teamRepo.Create(ctx, team); err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	members := make([]*domain.User, 0, len(input.Members))

	for _, m := range input.Members {
		user := &domain.User{
			Id:       m.UserID,
			Username: m.Username,
			TeamId:   team.Id,
			IsActive: m.IsActive,
		}

		if err := t.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user %s: %w", m.UserID, err)
		}

		members = append(members, user)
	}

	return &domain.TeamWithMembers{
		Team:    team,
		Members: members,
	}, nil
}

func (t *Team) GetTeam(ctx context.Context, teamName string) (*domain.TeamWithMembers, error) {
	team, err := t.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	users, err := t.userRepo.ListByTeam(ctx, team.Id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to load team users: %w", err)
	}

	return &domain.TeamWithMembers{
		Team:    team,
		Members: users,
	}, nil
}
