package domain

import "time"

type PullRequest struct {
	PullRequestID     string
	PullRequestName   string
	AuthorID          string
	Status            string // OPEN, MERGED
	AssignedReviewers []string
	CreatedAt         *time.Time
	MergedAt          *time.Time
}
