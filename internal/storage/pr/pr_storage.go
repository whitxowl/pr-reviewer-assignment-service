package pr

import (
	"context"
	"fmt"

	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"
	pg "github.com/whitxowl/pr-reviewer-assignment-service.git/pkg/postgres"
)

type Storage struct {
	Db pg.DB
}

func New(db pg.DB) *Storage {
	return &Storage{
		Db: db,
	}
}

func (s *Storage) GetPRsReviewedBy(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	const op = "storage.user.GetPRsReviewedBy"

	const query = `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN pull_request_reviewers prr
			ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
	`

	rows, err := s.Db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var prs []*domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		prs = append(prs, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return prs, nil
}
