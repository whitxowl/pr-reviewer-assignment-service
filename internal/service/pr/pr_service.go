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
	SetStatusMerged(ctx context.Context, prID string) (*domain.PullRequest, error)
	GetPRAuthorID(ctx context.Context, prID string) (string, error)
	ReassignReviewer(ctx context.Context, prID string, oldReviewerID string, newReviewerID string) (*domain.PullRequest, error)
}

const (
	statusOpen       = "OPEN"
	firstAssignLimit = 2
	reassignLimit    = 1
)

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
		log.DebugContext(ctx, "author or team not found", "error", err)
		return nil, serviceErr.ErrAuthorNotCorrect
	}

	err = s.prStorage.CreatePR(ctx, prID, prName, authorID)
	if errors.Is(err, storageErr.ErrPRExists) {
		log.DebugContext(ctx, "pr already exists", "error", err)
		return nil, serviceErr.ErrPRExists
	}
	if err != nil {
		log.ErrorContext(ctx, "error creating pr", "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	reviewers, err := s.userStorage.GetPotentialReviewersIDs(ctx, authorID, authorID, firstAssignLimit)
	if err != nil {
		log.ErrorContext(ctx, "error finding potential reviewers", "error", err)
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
		Status:            statusOpen,
		AssignedReviewers: reviewers,
	}, nil
}

func (s *Service) SetStatusMerged(ctx context.Context, prID string) (*domain.PullRequest, error) {
	const op = "service.pr.SetStatusMerged"

	log := s.log.With(
		slog.String("op", op),
		slog.String("prID", prID),
	)

	pr, err := s.prStorage.SetStatusMerged(ctx, prID)
	if errors.Is(err, storageErr.ErrPRNotFound) {
		log.DebugContext(ctx, "pr not found", "error", err)
		return nil, serviceErr.ErrPRNotFound
	}
	if err != nil {
		log.ErrorContext(ctx, "error setting status merged", "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pr, nil
}

func (s *Service) ReassignReviewer(
	ctx context.Context,
	prID string,
	oldReviewerID string,
) (*domain.PullRequest, string, error) {
	const op = "service.pr.ReassignReviewer"

	log := s.log.With(
		slog.String("op", op),
		slog.String("prID", prID),
		slog.String("oldReviewerID", oldReviewerID),
	)

	authorID, err := s.prStorage.GetPRAuthorID(ctx, prID)
	if errors.Is(err, storageErr.ErrPRNotFound) {
		log.DebugContext(ctx, "pr not found", "error", err)
		return nil, "", serviceErr.ErrPRNotFound
	}
	if err != nil {
		log.ErrorContext(ctx, "error getting pr", "error", err)
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	reviewers, err := s.userStorage.GetPotentialReviewersIDs(ctx, authorID, oldReviewerID, reassignLimit)
	if err != nil {
		log.ErrorContext(ctx, "error finding potential reviewers", "error", err)
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	var newReviewerID string
	if len(reviewers) == 0 {
		newReviewerID = ""
	} else {
		newReviewerID = reviewers[0]
	}

	pr, err := s.prStorage.ReassignReviewer(ctx, prID, oldReviewerID, newReviewerID)
	if errors.Is(err, storageErr.ErrPRNotFound) {
		log.DebugContext(ctx, "pr not found", "error", err)
		return nil, "", serviceErr.ErrPRNotFound
	}
	if errors.Is(err, storageErr.ErrPRMerged) {
		log.DebugContext(ctx, "pr already merged", "error", err)
		return nil, "", serviceErr.ErrPRMerged
	}
	if errors.Is(err, storageErr.ErrReviewerNotFound) {
		log.DebugContext(ctx, "reviewer not assigned to this pr", "error", err)
		return nil, "", serviceErr.ErrReviewerNotFound
	}
	if err != nil {
		log.ErrorContext(ctx, "error reassigning reviewer", "error", err)
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	log.InfoContext(ctx, "reviewer reassigned successfully",
		"oldReviewer", oldReviewerID,
		"newReviewer", newReviewerID)

	return pr, newReviewerID, nil
}
