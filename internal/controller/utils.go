package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	ierrors "github.com/kimbasn/printly/internal/errors"
	"github.com/kimbasn/printly/internal/dto"
)

// handleServiceError centralizes error handling for the user controller.
// It maps service-layer errors to appropriate HTTP status codes and responses.
func HandleServiceError(ctx *gin.Context, err error, defaultMessage string) {
	switch {
	case errors.Is(err, ierrors.ErrUserNotFound):
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, ierrors.ErrUserAlreadyExists):
		ctx.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, ierrors.ErrPrintCenterNotFound):
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, ierrors.ErrOrderNotFound):
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: defaultMessage})
	}
}