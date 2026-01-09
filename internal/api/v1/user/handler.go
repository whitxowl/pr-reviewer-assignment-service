package user

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"
)

type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
}

type Handler struct {
	userService UserService
}

func New(userService UserService) *Handler {
	return &Handler{
		userService: userService,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	usersGroup := router.Group("/users")
	{
		usersGroup.POST("/setIsActive", h.setIsActive)
	}
}
