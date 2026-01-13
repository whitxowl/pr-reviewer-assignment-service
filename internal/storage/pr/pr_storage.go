package pr

import (
	"context"
	"errors"
	"fmt"

	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"
	storageErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/storage/errors"
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

const (
	statusMerged = "MERGED"
)

func (s *Storage) GetPRsReviewedBy(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	const op = "storage.pr.GetPRsReviewedBy"

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
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		prs = append(prs, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return prs, nil
}

func (s *Storage) CreatePR(ctx context.Context, prID string, prName string, authorID string) error {
	const op = "storage.pr.CreatePR"

	const query = `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id)
		VALUES ($1, $2, $3)
	`

	_, err := s.Db.Exec(ctx, query, prID, prName, authorID)
	if pg.IsUniqueViolationError(err) {
		return fmt.Errorf("%s: %w", op, storageErr.ErrPRExists)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) AssignReviewers(ctx context.Context, prID string, reviewersIDs []string) error {
	const op = "storage.pr.AssignReviewers"

	if len(reviewersIDs) == 0 {
		return nil
	}

	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}
	defer tx.Rollback(ctx)

	const query = "INSERT INTO pull_request_reviewers VALUES ($1, $2)"

	batch := &pg.Batch{}
	for _, reviewerID := range reviewersIDs {
		batch.Queue(query, prID, reviewerID)
	}
	batchResults := tx.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range reviewersIDs {
		_, err = batchResults.Exec()
		if err != nil {
			e := batchResults.Close()

			return errors.Join(e, fmt.Errorf("%s: %w", op, err))
		}
	}

	if err = batchResults.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) SetStatusMerged(ctx context.Context, prID string) (*domain.PullRequest, error) {
	const op = "storage.pr.SetStatusMerged"

	const query = `
        UPDATE pull_requests 
        SET status = 'MERGED', 
            merged_at = COALESCE(merged_at, NOW())
        WHERE pull_request_id = $1
        RETURNING pull_request_id, 
                  pull_request_name, 
                  author_id, 
                  status,
                  merged_at
    `

	var pr domain.PullRequest
	err := s.Db.QueryRow(ctx, query, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.MergedAt,
	)
	if pg.IsNoRowsError(err) {
		return nil, fmt.Errorf("%s: %w", op, storageErr.ErrPRNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	reviewers, err := s.getReviewersByPRID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (s *Storage) GetPRAuthorID(ctx context.Context, prID string) (string, error) {
	const op = "storage.pr.GetPRAuthorID"

	const query = "SELECT author_id FROM pull_requests WHERE pull_request_id = $1"

	var authorID string
	err := s.Db.QueryRow(ctx, query, prID).Scan(&authorID)
	if pg.IsNoRowsError(err) {
		return "", fmt.Errorf("%s: %w", op, storageErr.ErrPRNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return authorID, nil
}

func (s *Storage) ReassignReviewer(
	ctx context.Context,
	prID string,
	oldReviewerID string,
	newReviewerID string,
) (*domain.PullRequest, error) {
	const op = "storage.pr.ReassignReviewer"

	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	if err := s.checkPRStatus(ctx, tx, prID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.removeReviewerTx(ctx, tx, prID, oldReviewerID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if newReviewerID != "" {
		if err := s.addReviewerTx(ctx, tx, prID, newReviewerID); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	pr, err := s.getPRWithReviewersTx(ctx, tx, prID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pr, nil
}

func (s *Storage) checkPRStatus(ctx context.Context, tx pg.Tx, prID string) error {
	const op = "storage.pr.checkPRStatus"

	const query = "SELECT status FROM pull_requests WHERE pull_request_id = $1"

	var status string
	err := tx.QueryRow(ctx, query, prID).Scan(&status)
	if pg.IsNoRowsError(err) {
		return fmt.Errorf("%s: %w", op, storageErr.ErrPRNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if status == statusMerged {
		return fmt.Errorf("%s: %w", op, storageErr.ErrPRMerged)
	}

	return nil
}

func (s *Storage) removeReviewerTx(ctx context.Context, tx pg.Tx, prID string, reviewerID string) error {
	const op = "storage.pr.removeReviewerTx"

	const query = `
        DELETE FROM pull_request_reviewers
        WHERE pull_request_id = $1 AND user_id = $2
    `

	result, err := tx.Exec(ctx, query, prID, reviewerID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storageErr.ErrReviewerNotFound)
	}

	return nil
}

func (s *Storage) addReviewerTx(ctx context.Context, tx pg.Tx, prID string, reviewerID string) error {
	const op = "storage.pr.addReviewerTx"

	const query = `
        INSERT INTO pull_request_reviewers (pull_request_id, user_id)
        VALUES ($1, $2)
    `

	_, err := tx.Exec(ctx, query, prID, reviewerID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) getPRWithReviewersTx(ctx context.Context, tx pg.Tx, prID string) (*domain.PullRequest, error) {
	const op = "storage.pr.getPRWithReviewersTx"

	const query = `
        SELECT pull_request_id, pull_request_name, author_id, status
        FROM pull_requests
        WHERE pull_request_id = $1
    `

	var pr domain.PullRequest
	err := tx.QueryRow(ctx, query, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	const getReviewersQuery = `
        SELECT user_id
        FROM pull_request_reviewers
        WHERE pull_request_id = $1
    `

	rows, err := tx.Query(ctx, getReviewersQuery, prID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		reviewers = append(reviewers, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (s *Storage) getReviewersByPRID(ctx context.Context, prID string) ([]string, error) {
	const op = "storage.pr.GetReviewersByPRID"

	const query = `
        SELECT user_id 
        FROM pull_request_reviewers 
        WHERE pull_request_id = $1
    `

	rows, err := s.Db.Query(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		reviewers = append(reviewers, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return reviewers, nil
}
