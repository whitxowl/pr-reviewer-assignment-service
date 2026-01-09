package team

import (
	"context"
	"fmt"

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

func (s *Storage) TeamExists(ctx context.Context, teamName string) (bool, error) {
	const op = "storage.team.TeamExists"

	const query = "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)"

	var exists bool
	err := s.Db.QueryRow(ctx, query, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}
