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

type GetReviewedResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestResponse `json:"pull_requests"`
}

type PullRequestResponse struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull-request-name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ToSetIsActiveResponse(user *domain.User) SetIsActiveResponse {
	response := SetIsActiveResponse{
		User: UserResponse{
			UserID:   user.UserID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	}

	return response
}

func ToGetReviewedResponse(userID string, prs []*domain.PullRequest) GetReviewedResponse {
	response := GetReviewedResponse{
		UserID:       userID,
		PullRequests: make([]PullRequestResponse, len(prs)),
	}

	for i, pr := range prs {
		response.PullRequests[i] = PullRequestResponse{
			PullRequestID:   pr.PullRequestID,
			PullRequestName: pr.PullRequestName,
			AuthorID:        pr.AuthorID,
			Status:          pr.Status,
		}
	}

	return response
}
