package team

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"
	serviceErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/errors"
	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/team/mocks"
	storageErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/storage/errors"
)

func TestService_CreateTeam(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name          string
		team          domain.Team
		setupMocks    func(*mocks.MockTeamStorage, *mocks.MockUserStorage)
		expectedError error
	}{
		{
			name: "success - team created",
			team: domain.Team{
				TeamName: "backend",
				Members: []*domain.User{
					{UserID: "u1", Username: "user1", TeamName: "backend"},
					{UserID: "u2", Username: "user2", TeamName: "backend"},
				},
			},
			setupMocks: func(teamStorage *mocks.MockTeamStorage, userStorage *mocks.MockUserStorage) {
				teamStorage.EXPECT().
					CreateTeam(ctx, "backend").
					Return(nil).
					Once()

				userStorage.EXPECT().
					UpsertUsers(ctx, mock.MatchedBy(func(users []*domain.User) bool {
						return len(users) == 2 && users[0].Username == "user1" && users[1].Username == "user2"
					})).
					Return(nil).
					Once()
			},
			expectedError: nil,
		},
		{
			name: "error - team already exists",
			team: domain.Team{
				TeamName: "existing-team",
				Members:  []*domain.User{},
			},
			setupMocks: func(teamStorage *mocks.MockTeamStorage, userStorage *mocks.MockUserStorage) {
				teamStorage.EXPECT().
					CreateTeam(ctx, "existing-team").
					Return(storageErr.ErrTeamExists).
					Once()
			},
			expectedError: serviceErr.ErrTeamExists,
		},
		{
			name: "error - storage error on create team",
			team: domain.Team{
				TeamName: "new-team",
				Members:  []*domain.User{},
			},
			setupMocks: func(teamStorage *mocks.MockTeamStorage, userStorage *mocks.MockUserStorage) {
				teamStorage.EXPECT().
					CreateTeam(ctx, "new-team").
					Return(errors.New("database error")).
					Once()
			},
			expectedError: errors.New("service.team.CreateTeam: database error"),
		},
		{
			name: "error - storage error on upsert users",
			team: domain.Team{
				TeamName: "new-team",
				Members: []*domain.User{
					{UserID: "u1", Username: "user1", TeamName: "new-team"},
				},
			},
			setupMocks: func(teamStorage *mocks.MockTeamStorage, userStorage *mocks.MockUserStorage) {
				teamStorage.EXPECT().
					CreateTeam(ctx, "new-team").
					Return(nil).
					Once()

				userStorage.EXPECT().
					UpsertUsers(ctx, mock.Anything).
					Return(errors.New("upsert error")).
					Once()
			},
			expectedError: errors.New("service.team.CreateTeam: upsert error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			teamStorage := mocks.NewMockTeamStorage(t)
			userStorage := mocks.NewMockUserStorage(t)
			tt.setupMocks(teamStorage, userStorage)

			service := New(log, teamStorage, userStorage)

			// Act
			err := service.CreateTeam(ctx, tt.team)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				if !errors.Is(tt.expectedError, err) {
					assert.Contains(t, err.Error(), tt.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetTeam(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name          string
		teamName      string
		setupMocks    func(*mocks.MockTeamStorage, *mocks.MockUserStorage)
		expectedTeam  *domain.Team
		expectedError error
	}{
		{
			name:     "success - team found with members",
			teamName: "backend",
			setupMocks: func(teamStorage *mocks.MockTeamStorage, userStorage *mocks.MockUserStorage) {
				teamStorage.EXPECT().
					TeamExists(ctx, "backend").
					Return(true, nil).
					Once()

				userStorage.EXPECT().
					GetUsersByTeamName(ctx, "backend").
					Return([]*domain.User{
						{UserID: "u1", Username: "user1", TeamName: "backend"},
						{UserID: "u2", Username: "user2", TeamName: "backend"},
					}, nil).
					Once()
			},
			expectedTeam: &domain.Team{
				TeamName: "backend",
				Members: []*domain.User{
					{UserID: "u1", Username: "user1", TeamName: "backend"},
					{UserID: "u2", Username: "user2", TeamName: "backend"},
				},
			},
			expectedError: nil,
		},
		{
			name:     "success - team found with no members",
			teamName: "empty-team",
			setupMocks: func(teamStorage *mocks.MockTeamStorage, userStorage *mocks.MockUserStorage) {
				teamStorage.EXPECT().
					TeamExists(ctx, "empty-team").
					Return(true, nil).
					Once()

				userStorage.EXPECT().
					GetUsersByTeamName(ctx, "empty-team").
					Return([]*domain.User{}, nil).
					Once()
			},
			expectedTeam: &domain.Team{
				TeamName: "empty-team",
				Members:  []*domain.User{},
			},
			expectedError: nil,
		},
		{
			name:     "error - team not found",
			teamName: "non-existent-team",
			setupMocks: func(teamStorage *mocks.MockTeamStorage, userStorage *mocks.MockUserStorage) {
				teamStorage.EXPECT().
					TeamExists(ctx, "non-existent-team").
					Return(false, nil).
					Once()
			},
			expectedTeam:  nil,
			expectedError: serviceErr.ErrTeamNotFound,
		},
		{
			name:     "error - storage error on team exists check",
			teamName: "backend",
			setupMocks: func(teamStorage *mocks.MockTeamStorage, userStorage *mocks.MockUserStorage) {
				teamStorage.EXPECT().
					TeamExists(ctx, "backend").
					Return(false, errors.New("database error")).
					Once()
			},
			expectedTeam:  nil,
			expectedError: errors.New("service.team.GetTeam: database error"),
		},
		{
			name:     "error - storage error on get users",
			teamName: "backend",
			setupMocks: func(teamStorage *mocks.MockTeamStorage, userStorage *mocks.MockUserStorage) {
				teamStorage.EXPECT().
					TeamExists(ctx, "backend").
					Return(true, nil).
					Once()

				userStorage.EXPECT().
					GetUsersByTeamName(ctx, "backend").
					Return(nil, errors.New("query error")).
					Once()
			},
			expectedTeam:  nil,
			expectedError: errors.New("service.team.GetTeam: query error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			teamStorage := mocks.NewMockTeamStorage(t)
			userStorage := mocks.NewMockUserStorage(t)
			tt.setupMocks(teamStorage, userStorage)

			service := New(log, teamStorage, userStorage)

			// Act
			result, err := service.GetTeam(ctx, tt.teamName)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				if !errors.Is(tt.expectedError, err) {
					assert.Contains(t, err.Error(), tt.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedTeam.TeamName, result.TeamName)
				assert.Equal(t, len(tt.expectedTeam.Members), len(result.Members))
				for i, member := range tt.expectedTeam.Members {
					assert.Equal(t, member.UserID, result.Members[i].UserID)
					assert.Equal(t, member.Username, result.Members[i].Username)
					assert.Equal(t, member.TeamName, result.Members[i].TeamName)
				}
			}
		})
	}
}
