package pr

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"
)

type PRService interface {
	CreatePR(ctx context.Context, prID string, prName string, authorID string) (*domain.PullRequest, error)
	SetStatusMerged(ctx context.Context, prID string) (*domain.PullRequest, error)
}

type Handler struct {
	prService PRService
}

func New(prService PRService) *Handler {
	return &Handler{
		prService: prService,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	prGroup := router.Group("/pullRequest")
	{
		prGroup.POST("create", h.create)
		prGroup.POST("merge", h.merge)
	}
}
