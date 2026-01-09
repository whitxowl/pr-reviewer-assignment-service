package team

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"
	serviceErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/errors"
	storageErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/storage/errors"
)

type TeamStorage interface {
	CreateTeam(ctx context.Context, teamName string) error
	TeamExists(ctx context.Context, teamName string) (bool, error)
}

type UserStorage interface {
	UpsertUsers(ctx context.Context, users []*domain.User) error
	GetUsersByTeamName(ctx context.Context, teamName string) ([]*domain.User, error)
}

type Service struct {
	log         *slog.Logger
	teamStorage TeamStorage
	userStorage UserStorage
}

func New(log *slog.Logger, teamStorage TeamStorage, userStorage UserStorage) *Service {
	return &Service{
		log:         log,
		teamStorage: teamStorage,
		userStorage: userStorage,
	}
}

func (s *Service) CreateTeam(ctx context.Context, team domain.Team) error {
	const op = "service.team.CreateTeam"

	log := s.log.With(
		slog.String("op", op),
		slog.String("teamName", team.TeamName),
	)

	err := s.teamStorage.CreateTeam(ctx, team.TeamName)
	if errors.Is(err, storageErr.ErrTeamExists) {
		log.DebugContext(ctx, "team already exists")
		return serviceErr.ErrTeamExists
	}
	if err != nil {
		log.ErrorContext(ctx, "error creating team", "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = s.userStorage.UpsertUsers(ctx, team.Members); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.InfoContext(ctx, "team created successfully")

	return nil
}

func (s *Service) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	const op = "service.team.GetTeam"

	log := s.log.With(
		slog.String("op", op),
		slog.String("teamName", teamName),
	)

	exists, err := s.teamStorage.TeamExists(ctx, teamName)
	if err != nil {
		log.ErrorContext(ctx, "error checking if team exists", "error", err)
	}
	if !exists {
		return nil, serviceErr.ErrTeamNotFound
	}

	users, err := s.userStorage.GetUsersByTeamName(ctx, teamName)
	if err != nil {
		log.ErrorContext(ctx, "error getting users", "error", err)
	}

	return &domain.Team{
		TeamName: teamName,
		Members:  users,
	}, nil
}
