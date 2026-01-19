package pr

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
	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/pr/mocks"
	storageErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/storage/errors"
)

func TestService_CreatePR(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name          string
		prID          string
		prName        string
		authorID      string
		setupMocks    func(*mocks.MockUserStorage, *mocks.MockPRStorage)
		expectedPR    *domain.PullRequest
		expectedError error
	}{
		{
			name:     "success - PR created with reviewers",
			prID:     "pr-123",
			prName:   "Add new feature",
			authorID: "u1",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				userStorage.EXPECT().
					UserExistsAndHasTeam(ctx, "u1").
					Return(true, nil).
					Once()

				prStorage.EXPECT().
					CreatePR(ctx, "pr-123", "Add new feature", "u1").
					Return(nil).
					Once()

				userStorage.EXPECT().
					GetPotentialReviewersIDs(ctx, "u1", "u1", 2).
					Return([]string{"u11", "u12"}, nil).
					Once()

				prStorage.EXPECT().
					AssignReviewers(ctx, "pr-123", []string{"u11", "u12"}).
					Return(nil).
					Once()
			},
			expectedPR: &domain.PullRequest{
				PullRequestID:     "pr-123",
				PullRequestName:   "Add new feature",
				AuthorID:          "u1",
				Status:            "OPEN",
				AssignedReviewers: []string{"u11", "u12"},
			},
			expectedError: nil,
		},
		{
			name:     "success - PR created with one reviewer",
			prID:     "pr-456",
			prName:   "Fix bug",
			authorID: "u2",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				userStorage.EXPECT().
					UserExistsAndHasTeam(ctx, "u2").
					Return(true, nil).
					Once()

				prStorage.EXPECT().
					CreatePR(ctx, "pr-456", "Fix bug", "u2").
					Return(nil).
					Once()

				userStorage.EXPECT().
					GetPotentialReviewersIDs(ctx, "u2", "u2", 2).
					Return([]string{"u13"}, nil).
					Once()

				prStorage.EXPECT().
					AssignReviewers(ctx, "pr-456", []string{"u13"}).
					Return(nil).
					Once()
			},
			expectedPR: &domain.PullRequest{
				PullRequestID:     "pr-456",
				PullRequestName:   "Fix bug",
				AuthorID:          "u2",
				Status:            "OPEN",
				AssignedReviewers: []string{"u13"},
			},
			expectedError: nil,
		},
		{
			name:     "error - author not correct",
			prID:     "pr-789",
			prName:   "Update docs",
			authorID: "invalid_author",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				userStorage.EXPECT().
					UserExistsAndHasTeam(ctx, "invalid_author").
					Return(false, nil).
					Once()
			},
			expectedPR:    nil,
			expectedError: serviceErr.ErrAuthorNotCorrect,
		},
		{
			name:     "error - user exists check fails",
			prID:     "pr-101",
			prName:   "Refactor code",
			authorID: "u3",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				userStorage.EXPECT().
					UserExistsAndHasTeam(ctx, "u3").
					Return(false, errors.New("database error")).
					Once()
			},
			expectedPR:    nil,
			expectedError: errors.New("service.pr.CreatePR: database error"),
		},
		{
			name:     "error - PR already exists",
			prID:     "pr-202",
			prName:   "New feature",
			authorID: "u4",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				userStorage.EXPECT().
					UserExistsAndHasTeam(ctx, "u4").
					Return(true, nil).
					Once()

				prStorage.EXPECT().
					CreatePR(ctx, "pr-202", "New feature", "u4").
					Return(storageErr.ErrPRExists).
					Once()
			},
			expectedPR:    nil,
			expectedError: serviceErr.ErrPRExists,
		},
		{
			name:     "error - create PR storage error",
			prID:     "pr-303",
			prName:   "Hot fix",
			authorID: "u5",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				userStorage.EXPECT().
					UserExistsAndHasTeam(ctx, "u5").
					Return(true, nil).
					Once()

				prStorage.EXPECT().
					CreatePR(ctx, "pr-303", "Hot fix", "u5").
					Return(errors.New("insert failed")).
					Once()
			},
			expectedPR:    nil,
			expectedError: errors.New("service.pr.CreatePR: insert failed"),
		},
		{
			name:     "error - get potential reviewers fails",
			prID:     "pr-404",
			prName:   "Performance improvement",
			authorID: "u6",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				userStorage.EXPECT().
					UserExistsAndHasTeam(ctx, "u6").
					Return(true, nil).
					Once()

				prStorage.EXPECT().
					CreatePR(ctx, "pr-404", "Performance improvement", "u6").
					Return(nil).
					Once()

				userStorage.EXPECT().
					GetPotentialReviewersIDs(ctx, "u6", "u6", 2).
					Return(nil, errors.New("query error")).
					Once()
			},
			expectedPR:    nil,
			expectedError: errors.New("service.pr.CreatePR: query error"),
		},
		{
			name:     "error - assign reviewers fails",
			prID:     "pr-505",
			prName:   "Security patch",
			authorID: "u7",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				userStorage.EXPECT().
					UserExistsAndHasTeam(ctx, "u7").
					Return(true, nil).
					Once()

				prStorage.EXPECT().
					CreatePR(ctx, "pr-505", "Security patch", "u7").
					Return(nil).
					Once()

				userStorage.EXPECT().
					GetPotentialReviewersIDs(ctx, "u7", "u7", 2).
					Return([]string{"u14", "u15"}, nil).
					Once()

				prStorage.EXPECT().
					AssignReviewers(ctx, "pr-505", []string{"u14", "u15"}).
					Return(errors.New("assignment failed")).
					Once()
			},
			expectedPR:    nil,
			expectedError: errors.New("service.pr.CreatePR: assignment failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			userStorage := mocks.NewMockUserStorage(t)
			prStorage := mocks.NewMockPRStorage(t)
			tt.setupMocks(userStorage, prStorage)

			service := New(log, userStorage, prStorage)

			// Act
			result, err := service.CreatePR(ctx, tt.prID, tt.prName, tt.authorID)

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
				assert.Equal(t, tt.expectedPR.PullRequestID, result.PullRequestID)
				assert.Equal(t, tt.expectedPR.PullRequestName, result.PullRequestName)
				assert.Equal(t, tt.expectedPR.AuthorID, result.AuthorID)
				assert.Equal(t, tt.expectedPR.Status, result.Status)
				assert.Equal(t, tt.expectedPR.AssignedReviewers, result.AssignedReviewers)
			}
		})
	}
}

func TestService_SetStatusMerged(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	now := time.Now()

	tests := []struct {
		name          string
		prID          string
		setupMocks    func(*mocks.MockPRStorage)
		expectedPR    *domain.PullRequest
		expectedError error
	}{
		{
			name: "success - PR merged",
			prID: "pr-123",
			setupMocks: func(prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					SetStatusMerged(ctx, "pr-123").
					Return(&domain.PullRequest{
						PullRequestID:     "pr-123",
						PullRequestName:   "Add feature",
						AuthorID:          "u1",
						Status:            "MERGED",
						AssignedReviewers: []string{"u11", "u12"},
						CreatedAt:         &now,
						MergedAt:          &now,
					}, nil).
					Once()
			},
			expectedPR: &domain.PullRequest{
				PullRequestID:     "pr-123",
				PullRequestName:   "Add feature",
				AuthorID:          "u1",
				Status:            "MERGED",
				AssignedReviewers: []string{"u11", "u12"},
				CreatedAt:         &now,
				MergedAt:          &now,
			},
			expectedError: nil,
		},
		{
			name: "error - PR not found",
			prID: "pr-999",
			setupMocks: func(prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					SetStatusMerged(ctx, "pr-999").
					Return(nil, storageErr.ErrPRNotFound).
					Once()
			},
			expectedPR:    nil,
			expectedError: serviceErr.ErrPRNotFound,
		},
		{
			name: "error - storage error",
			prID: "pr-456",
			setupMocks: func(prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					SetStatusMerged(ctx, "pr-456").
					Return(nil, errors.New("update failed")).
					Once()
			},
			expectedPR:    nil,
			expectedError: errors.New("service.pr.SetStatusMerged: update failed"),
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
			result, err := service.SetStatusMerged(ctx, tt.prID)

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
				assert.Equal(t, tt.expectedPR.PullRequestID, result.PullRequestID)
				assert.Equal(t, tt.expectedPR.Status, result.Status)
			}
		})
	}
}

func TestService_ReassignReviewer(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	now := time.Now()

	tests := []struct {
		name          string
		prID          string
		oldReviewerID string
		setupMocks    func(*mocks.MockUserStorage, *mocks.MockPRStorage)
		expectedPR    *domain.PullRequest
		expectedNewID string
		expectedError error
	}{
		{
			name:          "success - reviewer reassigned",
			prID:          "pr-123",
			oldReviewerID: "u11",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRAuthorID(ctx, "pr-123").
					Return("u1", nil).
					Once()

				userStorage.EXPECT().
					GetPotentialReviewersIDs(ctx, "u1", "u11", 1).
					Return([]string{"u13"}, nil).
					Once()

				prStorage.EXPECT().
					ReassignReviewer(ctx, "pr-123", "u11", "u13").
					Return(&domain.PullRequest{
						PullRequestID:     "pr-123",
						PullRequestName:   "Feature",
						AuthorID:          "u1",
						Status:            "OPEN",
						AssignedReviewers: []string{"u12", "u13"},
						CreatedAt:         &now,
					}, nil).
					Once()
			},
			expectedPR: &domain.PullRequest{
				PullRequestID:     "pr-123",
				PullRequestName:   "Feature",
				AuthorID:          "u1",
				Status:            "OPEN",
				AssignedReviewers: []string{"u12", "u13"},
				CreatedAt:         &now,
			},
			expectedNewID: "u13",
			expectedError: nil,
		},
		{
			name:          "success - no replacement found",
			prID:          "pr-456",
			oldReviewerID: "u15",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRAuthorID(ctx, "pr-456").
					Return("u2", nil).
					Once()

				userStorage.EXPECT().
					GetPotentialReviewersIDs(ctx, "u2", "u15", 1).
					Return([]string{}, nil).
					Once()

				prStorage.EXPECT().
					ReassignReviewer(ctx, "pr-456", "u15", "").
					Return(&domain.PullRequest{
						PullRequestID:     "pr-456",
						PullRequestName:   "Bug fix",
						AuthorID:          "u2",
						Status:            "OPEN",
						AssignedReviewers: []string{"u14"},
						CreatedAt:         &now,
					}, nil).
					Once()
			},
			expectedPR: &domain.PullRequest{
				PullRequestID:     "pr-456",
				PullRequestName:   "Bug fix",
				AuthorID:          "u2",
				Status:            "OPEN",
				AssignedReviewers: []string{"u14"},
				CreatedAt:         &now,
			},
			expectedNewID: "",
			expectedError: nil,
		},
		{
			name:          "error - PR not found on get author",
			prID:          "pr-999",
			oldReviewerID: "u11",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRAuthorID(ctx, "pr-999").
					Return("", storageErr.ErrPRNotFound).
					Once()
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: serviceErr.ErrPRNotFound,
		},
		{
			name:          "error - get author storage error",
			prID:          "pr-789",
			oldReviewerID: "u12",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRAuthorID(ctx, "pr-789").
					Return("", errors.New("query failed")).
					Once()
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: errors.New("service.pr.ReassignReviewer: query failed"),
		},
		{
			name:          "error - get potential reviewers fails",
			prID:          "pr-321",
			oldReviewerID: "u13",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRAuthorID(ctx, "pr-321").
					Return("u3", nil).
					Once()

				userStorage.EXPECT().
					GetPotentialReviewersIDs(ctx, "u3", "u13", 1).
					Return(nil, errors.New("database error")).
					Once()
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: errors.New("service.pr.ReassignReviewer: database error"),
		},
		{
			name:          "error - PR not found on reassign",
			prID:          "pr-654",
			oldReviewerID: "u14",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRAuthorID(ctx, "pr-654").
					Return("u4", nil).
					Once()

				userStorage.EXPECT().
					GetPotentialReviewersIDs(ctx, "u4", "u14", 1).
					Return([]string{"u16"}, nil).
					Once()

				prStorage.EXPECT().
					ReassignReviewer(ctx, "pr-654", "u14", "u16").
					Return(nil, storageErr.ErrPRNotFound).
					Once()
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: serviceErr.ErrPRNotFound,
		},
		{
			name:          "error - PR already merged",
			prID:          "pr-987",
			oldReviewerID: "u17",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRAuthorID(ctx, "pr-987").
					Return("u5", nil).
					Once()

				userStorage.EXPECT().
					GetPotentialReviewersIDs(ctx, "u5", "u17", 1).
					Return([]string{"u18"}, nil).
					Once()

				prStorage.EXPECT().
					ReassignReviewer(ctx, "pr-987", "u17", "u18").
					Return(nil, storageErr.ErrPRMerged).
					Once()
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: serviceErr.ErrPRMerged,
		},
		{
			name:          "error - reviewer not found",
			prID:          "pr-555",
			oldReviewerID: "u19",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRAuthorID(ctx, "pr-555").
					Return("u6", nil).
					Once()

				userStorage.EXPECT().
					GetPotentialReviewersIDs(ctx, "u6", "u19", 1).
					Return([]string{"u110"}, nil).
					Once()

				prStorage.EXPECT().
					ReassignReviewer(ctx, "pr-555", "u19", "u110").
					Return(nil, storageErr.ErrReviewerNotFound).
					Once()
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: serviceErr.ErrReviewerNotFound,
		},
		{
			name:          "error - reassign storage error",
			prID:          "pr-888",
			oldReviewerID: "u111",
			setupMocks: func(userStorage *mocks.MockUserStorage, prStorage *mocks.MockPRStorage) {
				prStorage.EXPECT().
					GetPRAuthorID(ctx, "pr-888").
					Return("u7", nil).
					Once()

				userStorage.EXPECT().
					GetPotentialReviewersIDs(ctx, "u7", "u111", 1).
					Return([]string{"u112"}, nil).
					Once()

				prStorage.EXPECT().
					ReassignReviewer(ctx, "pr-888", "u111", "u112").
					Return(nil, errors.New("update failed")).
					Once()
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: errors.New("service.pr.ReassignReviewer: update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			userStorage := mocks.NewMockUserStorage(t)
			prStorage := mocks.NewMockPRStorage(t)
			tt.setupMocks(userStorage, prStorage)

			service := New(log, userStorage, prStorage)

			// Act
			resultPR, resultNewID, err := service.ReassignReviewer(ctx, tt.prID, tt.oldReviewerID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, resultPR)
				assert.Empty(t, resultNewID)
				if !errors.Is(tt.expectedError, err) {
					assert.Contains(t, err.Error(), tt.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resultPR)
				assert.Equal(t, tt.expectedPR.PullRequestID, resultPR.PullRequestID)
				assert.Equal(t, tt.expectedPR.Status, resultPR.Status)
				assert.Equal(t, tt.expectedNewID, resultNewID)
			}
		})
	}
}
