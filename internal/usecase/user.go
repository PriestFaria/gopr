package usecase

import (
	"context"
	"fmt"

	"gopr/internal/domain"
	"gopr/internal/repo"
)

type User struct {
	userRepo repo.User
	teamRepo repo.Team
}

func NewUser(userRepo repo.User, teamRepo repo.Team) *User {
	return &User{
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

func (u *User) SetActive(ctx context.Context, userID string, active bool) (*domain.User, string, error) {
	if err := u.userRepo.UpdateIsActive(ctx, userID, active); err != nil {
		return nil, "", fmt.Errorf("failed to update activity: %w", err)
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load user: %w", err)
	}

	var teamName string
	if user.TeamId != "" {
		team, err := u.teamRepo.GetByID(ctx, user.TeamId)
		if err != nil {
			return nil, "", fmt.Errorf("failed to load team: %w", err)
		}
		teamName = team.Name
	}

	return user, teamName, nil
}
