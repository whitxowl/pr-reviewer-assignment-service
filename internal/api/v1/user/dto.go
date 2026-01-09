package user

import "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"

type SetIsActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active" default:"true"`
}

type SetIsActiveResponse struct {
	User UserResponse `json:"user"`
}

type UserResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ToSetIsActiveResponse(user *domain.User) *SetIsActiveResponse {
	response := SetIsActiveResponse{
		User: UserResponse{
			UserID:   user.UserID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	}

	return &response
}
