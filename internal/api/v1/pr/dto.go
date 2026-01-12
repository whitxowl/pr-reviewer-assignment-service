package pr

import "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"

type CreatePRRequest struct {
	PRID     string `json:"pull_request_id" binding:"required"`
	PRName   string `json:"pull_request_name" binding:"required"`
	AuthorID string `json:"author_id" binding:"required"`
}

type CreatePRResponse struct {
	PR PRResponse `json:"pr"`
}

type PRResponse struct {
	PRID      string   `json:"pull_request_id"`
	PRName    string   `json:"pull_request_name"`
	AuthorID  string   `json:"author_id"`
	Status    string   `json:"status"`
	Reviewers []string `json:"assigned_reviewers"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ToCreatePRResponse(pr *domain.PullRequest) CreatePRResponse {
	return CreatePRResponse{
		PR: PRResponse{
			PRID:      pr.PullRequestID,
			PRName:    pr.PullRequestName,
			AuthorID:  pr.AuthorID,
			Status:    pr.Status,
			Reviewers: pr.AssignedReviewers,
		},
	}
}
