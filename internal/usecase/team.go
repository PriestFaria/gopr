package usecase

import (
	context "context"
	"fmt"

	"gopr/internal/domain"
	"gopr/internal/repo"

	"github.com/google/uuid"
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

	err := t.teamRepo.Create(ctx, team)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	var members []*domain.User

	for _, m := range input.Members {
		user := &domain.User{
			Id:       m.UserID,
			Username: m.Username,
			TeamId:   team.Id,
			IsActive: m.IsActive,
		}

		err := t.userRepo.Create(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to create user %s: %w", m.UserID, err)
		}

		members = append(members, user)
	}

	return &domain.TeamWithMembers{
		Team:    team,
		Members: members,
	}, nil
}

func (t *Team) GetTeam(ctx context.Context, teamID string) (*domain.TeamWithMembers, error) {
	team, err := t.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	users, err := t.userRepo.ListByTeam(ctx, teamID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to load team users: %w", err)
	}

	return &domain.TeamWithMembers{
		Team:    team,
		Members: users,
	}, nil
}
