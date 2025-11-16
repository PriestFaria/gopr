package domain

type Team struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type TeamAddInput struct {
	TeamName string               `json:"team_name"`
	Members  []TeamAddMemberInput `json:"members"`
}

type TeamAddMemberInput struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}
