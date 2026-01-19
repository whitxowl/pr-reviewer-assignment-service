package user

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"
	serviceErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/errors"
	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/user/mocks"
	storageErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/storage/errors"
)

func TestService_SetIsActive(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name          string
		userID        string
		isActive      bool
		setupMocks    func(*mocks.MockUserStorage)
		expectedUser  *domain.User
		expectedError error
	}{
		{
			name:     "success - activate user",
			userID:   "u1",
			isActive: true,
			setupMocks: func(userStorage *mocks.MockUserStorage) {
				userStorage.EXPECT().
					SetIsActive(ctx, "u1", true).
					Return(&domain.User{
						UserID:   "u1",
						Username: "john_doe",
						TeamName: "backend",
						IsActive: true,
					}, nil).
					Once()
			},
			expectedUser: &domain.User{
				UserID:   "u1",
				Username: "john_doe",
				TeamName: "backend",
				IsActive: true,
			},
			expectedError: nil,
		},
		{
			name:     "success - deactivate user",
			userID:   "u2",
			isActive: false,
			setupMocks: func(userStorage *mocks.MockUserStorage) {
				userStorage.EXPECT().
					SetIsActive(ctx, "u2", false).
					Return(&domain.User{
						UserID:   "u2",
						Username: "jane_smith",
						TeamName: "frontend",
						IsActive: false,
					}, nil).
					Once()
			},
			expectedUser: &domain.User{
				UserID:   "u2",
				Username: "jane_smith",
				TeamName: "frontend",
				IsActive: false,
			},
			expectedError: nil,
		},
		{
			name:     "error - user not found",
			userID:   "nonexistent",
			isActive: true,
			setupMocks: func(userStorage *mocks.MockUserStorage) {
				userStorage.EXPECT().
					SetIsActive(ctx, "nonexistent", true).
					Return(nil, storageErr.ErrUserNotFound).
					Once()
			},
			expectedUser:  nil,
			expectedError: serviceErr.ErrUserNotFound,
		},
		{
			name:     "error - storage error",
			userID:   "u3",
			isActive: true,
			setupMocks: func(userStorage *mocks.MockUserStorage) {
				userStorage.EXPECT().
					SetIsActive(ctx, "u3", true).
					Return(nil, errors.New("database connection error")).
					Once()
			},
			expectedUser:  nil,
			expectedError: errors.New("storage.user.SetIsActive: database connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			userStorage := mocks.NewMockUserStorage(t)
			prStorage := mocks.NewMockPRStorage(t)
			tt.setupMocks(userStorage)

			service := New(log, userStorage, prStorage)

			// Act
			result, err := service.SetIsActive(ctx, tt.userID, tt.isActive)

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
				assert.Equal(t, tt.expectedUser.UserID, result.UserID)
				assert.Equal(t, tt.expectedUser.Username, result.Username)
				assert.Equal(t, tt.expectedUser.TeamName, result.TeamName)
				assert.Equal(t, tt.expectedUser.IsActive, result.IsActive)
			}
		})
	}
}

func TestService_GetPRsReviewedBy(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	now := time.Now()
	mergedTime := now.Add(-24 * time.Hour)

	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*mocks.MockPRStorage)
		expectedPRs   []*domain.PullRequest
		expectedError error
	}{
		{
			name:   "success - user has reviewed PRs",
			userID: "u1",
			setupMocks: func(prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRsReviewedBy(ctx, "u1").
					Return([]*domain.PullRequest{
						{
							PullRequestID:     "pr1",
							PullRequestName:   "Add new feature",
							AuthorID:          "u11",
							Status:            "MERGED",
							AssignedReviewers: []string{"u1", "u2"},
							CreatedAt:         &now,
							MergedAt:          &mergedTime,
						},
						{
							PullRequestID:     "pr2",
							PullRequestName:   "Fix bug",
							AuthorID:          "u12",
							Status:            "OPEN",
							AssignedReviewers: []string{"u1"},
							CreatedAt:         &now,
							MergedAt:          nil,
						},
					}, nil).
					Once()
			},
			expectedPRs: []*domain.PullRequest{
				{
					PullRequestID:     "pr1",
					PullRequestName:   "Add new feature",
					AuthorID:          "u11",
					Status:            "MERGED",
					AssignedReviewers: []string{"u1", "u2"},
					CreatedAt:         &now,
					MergedAt:          &mergedTime,
				},
				{
					PullRequestID:     "pr2",
					PullRequestName:   "Fix bug",
					AuthorID:          "u12",
					Status:            "OPEN",
					AssignedReviewers: []string{"u1"},
					CreatedAt:         &now,
					MergedAt:          nil,
				},
			},
			expectedError: nil,
		},
		{
			name:   "success - user has no reviewed PRs",
			userID: "u3",
			setupMocks: func(prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRsReviewedBy(ctx, "u3").
					Return([]*domain.PullRequest{}, nil).
					Once()
			},
			expectedPRs:   []*domain.PullRequest{},
			expectedError: nil,
		},
		{
			name:   "success - user has single PR",
			userID: "u2",
			setupMocks: func(prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRsReviewedBy(ctx, "u2").
					Return([]*domain.PullRequest{
						{
							PullRequestID:     "pr3",
							PullRequestName:   "Update documentation",
							AuthorID:          "u13",
							Status:            "OPEN",
							AssignedReviewers: []string{"u2"},
							CreatedAt:         &now,
							MergedAt:          nil,
						},
					}, nil).
					Once()
			},
			expectedPRs: []*domain.PullRequest{
				{
					PullRequestID:     "pr3",
					PullRequestName:   "Update documentation",
					AuthorID:          "u13",
					Status:            "OPEN",
					AssignedReviewers: []string{"u2"},
					CreatedAt:         &now,
					MergedAt:          nil,
				},
			},
			expectedError: nil,
		},
		{
			name:   "error - storage error",
			userID: "u4",
			setupMocks: func(prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRsReviewedBy(ctx, "u4").
					Return(nil, errors.New("query execution failed")).
					Once()
			},
			expectedPRs:   nil,
			expectedError: errors.New("storage.user.GetPRsReviewedBy: query execution failed"),
		},
		{
			name:   "error - database connection error",
			userID: "u5",
			setupMocks: func(prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRsReviewedBy(ctx, "u5").
					Return(nil, errors.New("connection timeout")).
					Once()
			},
			expectedPRs:   nil,
			expectedError: errors.New("storage.user.GetPRsReviewedBy: connection timeout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			userStorage := mocks.NewMockUserStorage(t)
			prStorage := mocks.NewMockPRStorage(t)
			tt.setupMocks(prStorage)

			service := New(log, userStorage, prStorage)

			// Act
			result, err := service.GetPRsReviewedBy(ctx, tt.userID)

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
				assert.Equal(t, len(tt.expectedPRs), len(result))

				for i, expectedPR := range tt.expectedPRs {
					assert.Equal(t, expectedPR.PullRequestID, result[i].PullRequestID)
					assert.Equal(t, expectedPR.PullRequestName, result[i].PullRequestName)
					assert.Equal(t, expectedPR.AuthorID, result[i].AuthorID)
					assert.Equal(t, expectedPR.Status, result[i].Status)
					assert.Equal(t, expectedPR.AssignedReviewers, result[i].AssignedReviewers)

					if expectedPR.CreatedAt != nil && result[i].CreatedAt != nil {
						assert.Equal(t, expectedPR.CreatedAt.Unix(), result[i].CreatedAt.Unix())
					}

					if expectedPR.MergedAt != nil && result[i].MergedAt != nil {
						assert.Equal(t, expectedPR.MergedAt.Unix(), result[i].MergedAt.Unix())
					} else {
						assert.Equal(t, expectedPR.MergedAt, result[i].MergedAt)
					}
				}
			}
		})
	}
}
