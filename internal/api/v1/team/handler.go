package team

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"
)

type TeamService interface {
	CreateTeam(ctx context.Context, team domain.Team) error
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
}

type Handler struct {
	teamService TeamService
}

func New(teamService TeamService) *Handler {
	return &Handler{
		teamService: teamService,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	teamGroup := router.Group("/team")
	{
		teamGroup.POST("/add", h.add)
		teamGroup.GET("/get", h.get)
	}
}
