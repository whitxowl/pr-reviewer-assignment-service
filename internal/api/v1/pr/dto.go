package pr

import (
	"time"

	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"
)

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

type MergeRequest struct {
	PRID string `json:"pull_request_id" binding:"required"`
}

type MergeResponse struct {
	PR PRMergedResponse `json:"pr"`
}

type PRMergedResponse struct {
	PRID      string     `json:"pull_request_id"`
	PRName    string     `json:"pull_request_name"`
	AuthorID  string     `json:"author_id"`
	Status    string     `json:"status"`
	Reviewers []string   `json:"assigned_reviewers"`
	MergedAt  *time.Time `json:"merged_at"`
}

type ReassignRequest struct {
	PRID          string `json:"pull_request_id" binding:"required"`
	OldReviewerID string `json:"old_reviewer_id" binding:"required"`
}

type ReassignResponse struct {
	PR PRReassignResponse `json:"pr"`
}

type PRReassignResponse struct {
	PRID       string   `json:"pull_request_id"`
	PRName     string   `json:"pull_request_name"`
	AuthorID   string   `json:"author_id"`
	Status     string   `json:"status"`
	Reviewers  []string `json:"assigned_reviewers"`
	ReplacedBy string   `json:"replaced_by"`
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

func ToMergeResponse(pr *domain.PullRequest) MergeResponse {
	return MergeResponse{
		PR: PRMergedResponse{
			PRID:      pr.PullRequestID,
			PRName:    pr.PullRequestName,
			AuthorID:  pr.AuthorID,
			Status:    pr.Status,
			Reviewers: pr.AssignedReviewers,
			MergedAt:  pr.MergedAt,
		},
	}
}

func ToReassignResponse(pr *domain.PullRequest, newReviewerID string) ReassignResponse {
	return ReassignResponse{
		PR: PRReassignResponse{
			PRID:       pr.PullRequestID,
			PRName:     pr.PullRequestName,
			AuthorID:   pr.AuthorID,
			Status:     pr.Status,
			Reviewers:  pr.AssignedReviewers,
			ReplacedBy: newReviewerID,
		},
	}
}
