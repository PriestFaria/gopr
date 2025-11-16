package usecase

import (
	context "context"
	"fmt"

	"gopr/internal/repo"
)

type User struct {
	userRepo repo.User
}

func NewUser(userRepo repo.User) *User {
	return &User{
		userRepo: userRepo,
	}
}

func (u *User) SetActive(ctx context.Context, userID string, active bool) error {
	err := u.userRepo.UpdateIsActive(ctx, userID, active)
	if err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}
	return nil
}
