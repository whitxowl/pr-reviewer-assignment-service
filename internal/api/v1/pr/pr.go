package pr

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	serviceErr "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/service/errors"
)

func (h *Handler) create(c *gin.Context) {
	var req CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid request body: " + err.Error(),
			},
		})
		return
	}

	pr, err := h.prService.CreatePR(c.Request.Context(), req.PRID, req.PRName, req.AuthorID)
	if errors.Is(err, serviceErr.ErrAuthorNotCorrect) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: ErrorDetail{
				Code:    "NOT_FOUND",
				Message: "resource not found",
			},
		})
		return
	}
	if errors.Is(err, serviceErr.ErrPRExists) {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error: ErrorDetail{
				Code:    "PR_EXISTS",
				Message: "PR id already exists",
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

	response := ToCreatePRResponse(pr)

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) merge(c *gin.Context) {
	var req MergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid request body: " + err.Error(),
			},
		})
		return
	}

	pr, err := h.prService.SetStatusMerged(c.Request.Context(), req.PRID)
	if errors.Is(err, serviceErr.ErrPRNotFound) {
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

	response := ToMergeResponse(pr)

	c.JSON(http.StatusOK, response)
}

func (h *Handler) reassign(c *gin.Context) {
	var req ReassignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid request body: " + err.Error(),
			},
		})
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(c.Request.Context(), req.PRID, req.OldReviewerID)
	if errors.Is(err, serviceErr.ErrPRNotFound) || errors.Is(err, serviceErr.ErrReviewerNotFound) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: ErrorDetail{
				Code:    "NOT_FOUND",
				Message: "resource not found",
			},
		})
		return
	}
	if errors.Is(err, serviceErr.ErrPRMerged) {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error: ErrorDetail{
				Code:    "PR_MERGED",
				Message: "cannot reassign on merged PR",
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

	response := ToReassignResponse(pr, newReviewerID)

	c.JSON(http.StatusOK, response)
}
