package user

import (
	"context"
	"errors"
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

func (s *Storage) UpsertUsers(ctx context.Context, users []*domain.User) error {
	const op = "storage.user.UpsertUsers"

	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}
	defer tx.Rollback(ctx)

	const query = `
		INSERT INTO users (user_id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
            ON CONFLICT (user_id)
            DO UPDATE SET
                username = EXCLUDED.username,
                team_name = EXCLUDED.team_name,
                is_active = EXCLUDED.is_active
	`

	batch := &pg.Batch{}
	for _, member := range users {
		batch.Queue(query, member.UserID, member.Username, member.TeamName, member.IsActive)
	}
	batchResults := tx.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range users {
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

func (s *Storage) GetUsersByTeamName(ctx context.Context, teamName string) ([]*domain.User, error) {
	const op = "storage.user.GetUsersByTeamName"

	const query = "SELECT user_id, username, is_active FROM users WHERE team_name = $1"
	rows, err := s.Db.Query(ctx, query, teamName)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(&user.UserID, &user.Username, &user.IsActive)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}
