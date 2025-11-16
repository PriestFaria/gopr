package domain

import "time"

type User struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
	TeamId    string    `json:"team_id"`
	IsActive  bool      `json:"is_active"`
}

type CreateUser struct{}

type PatchUser struct{}

type FilterUser struct{}
