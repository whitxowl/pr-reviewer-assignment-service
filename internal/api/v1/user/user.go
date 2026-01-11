package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	serviceErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/errors"
)

func (h *Handler) setIsActive(c *gin.Context) {
	var req SetIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid request body: " + err.Error(),
			},
		})
		return
	}

	user, err := h.userService.SetIsActive(c.Request.Context(), req.UserID, req.IsActive)
	if errors.Is(err, serviceErr.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: ErrorDetail{
				Code:    "USER_NOT_FOUND",
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

	response := ToSetIsActiveResponse(user)

	c.JSON(http.StatusOK, response)
}

func (h *Handler) getReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid request body: team_name is required",
			},
		})
		return
	}

	prs, err := h.userService.GetPRsReviewedBy(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "internal server error",
			},
		})
		return
	}

	response := ToGetReviewedResponse(userID, prs)

	c.JSON(http.StatusOK, response)
}
