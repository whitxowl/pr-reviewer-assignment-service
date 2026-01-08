package team

import (
	"context"
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

func (s *Storage) CreateTeam(ctx context.Context, teamName string) error {
	const op = "storage.team.CreateTeam"

	const query = "INSERT INTO teams(team_name) VALUES ($1)"

	_, err := s.Db.Exec(ctx, query, teamName)
	if pg.IsUniqueViolationError(err) {
		return fmt.Errorf("%s: %w", op, storageErr.ErrTeamExists)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	return nil, nil
}
