package team

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	serviceErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/errors"
)

func (h *Handler) add(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid request body: " + err.Error(),
			},
		})
		return
	}

	team := req.ToDomain()

	err := h.teamService.CreateTeam(c.Request.Context(), team)
	if errors.Is(err, serviceErr.ErrTeamExists) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "TEAM_EXISTS",
				Message: team.TeamName + " already exists",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "internal server error",
			},
		})
		return
	}

	response := ToTeamResponse(&team)

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) get(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid request body: team_name is required",
			},
		})
		return
	}

	team, err := h.teamService.GetTeam(c.Request.Context(), teamName)
	if errors.Is(err, serviceErr.ErrTeamNotFound) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: ErrorDetail{
				Code:    "NOT_FOUND",
				Message: "resource not found",
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "internal server error",
			},
		})
		return
	}

	response := ToGetTeamResponse(team)

	c.JSON(http.StatusOK, response)
}
