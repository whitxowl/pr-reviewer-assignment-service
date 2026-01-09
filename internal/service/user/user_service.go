package user

import (
	"context"
	"errors"
	"log/slog"

	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"
	serviceErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/errors"
	storageErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/storage/errors"
)

type UserStorage interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
}

type Service struct {
	log         *slog.Logger
	userStorage UserStorage
}

func New(log *slog.Logger, userStorage UserStorage) *Service {
	return &Service{
		log:         log,
		userStorage: userStorage,
	}
}

func (s *Service) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	const op = "storage.user.SetIsActive"

	log := s.log.With(
		slog.String("op", op),
		slog.String("user_id", userID),
	)

	user, err := s.userStorage.SetIsActive(ctx, userID, isActive)
	if errors.Is(err, storageErr.ErrUserNotFound) {
		log.DebugContext(ctx, "user not found", "error", err)
		return nil, serviceErr.ErrUserNotFound
	}
	if err != nil {
		log.ErrorContext(ctx, "failed to set is_active", "error", err)
	}

	log.InfoContext(ctx, "user set is_active", "user", user)

	return user, err
}
