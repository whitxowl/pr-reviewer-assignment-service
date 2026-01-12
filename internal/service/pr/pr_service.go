package pr

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"
	serviceErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/errors"
	storageErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/storage/errors"
)

type UserStorage interface {
	UserExistsAndHasTeam(ctx context.Context, userID string) (bool, error)
	GetPotentialReviewersIDs(ctx context.Context, authorID string, userID string, limit int) ([]string, error)
}

type PRStorage interface {
	CreatePR(ctx context.Context, prID string, prName string, authorID string) error
	AssignReviewers(ctx context.Context, prID string, reviewersIDs []string) error
}

type Service struct {
	log         *slog.Logger
	userStorage UserStorage
	prStorage   PRStorage
}

func New(log *slog.Logger, userStorage UserStorage, prStorage PRStorage) *Service {
	return &Service{
		log:         log,
		userStorage: userStorage,
		prStorage:   prStorage}
}

func (s *Service) CreatePR(
	ctx context.Context,
	prID string,
	prName string,
	authorID string,
) (*domain.PullRequest, error) {
	const op = "service.pr.CreatePR"

	const (
		status = "OPEN"
		limit  = 2
	)

	log := s.log.With(
		slog.String("op", op),
		slog.String("prID", prID),
		slog.String("prName", prName),
		slog.String("authorID", authorID),
	)

	userCorrect, err := s.userStorage.UserExistsAndHasTeam(ctx, authorID)
	if err != nil {
		log.ErrorContext(ctx, "error checking if user exists", "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if !userCorrect {
		return nil, serviceErr.ErrAuthorNotCorrect
	}

	err = s.prStorage.CreatePR(ctx, prID, prName, authorID)
	if errors.Is(err, storageErr.ErrPRExists) {
		return nil, serviceErr.ErrPRExists
	}
	if err != nil {
		log.ErrorContext(ctx, "error creating pr", "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	reviewers, err := s.userStorage.GetPotentialReviewersIDs(ctx, authorID, prID, limit)
	if err != nil {
		log.ErrorContext(ctx, "error creating pr", "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = s.prStorage.AssignReviewers(ctx, prID, reviewers)
	if err != nil {
		log.ErrorContext(ctx, "error assigning reviewers", "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            status,
		AssignedReviewers: reviewers,
	}, nil
}
