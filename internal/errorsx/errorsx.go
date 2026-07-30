package errorsx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func BadRequest(message string) *AppError {
	return &AppError{Code: 40000, Message: message}
}

func Unauthorized(message string) *AppError {
	return &AppError{Code: 40100, Message: message}
}

func Forbidden(message string) *AppError {
	return &AppError{Code: 40300, Message: message}
}

func NotFound(message string) *AppError {
	return &AppError{Code: 40400, Message: message}
}

func Conflict(message string) *AppError {
	return &AppError{Code: 40900, Message: message}
}

func Internal(message string) *AppError {
	return &AppError{Code: 50000, Message: message}
}

func GinResponse(c *gin.Context, err error) {
	appErr := Internal(LocalizeMessage("internal server error"))
	var candidate *AppError
	if errors.As(err, &candidate) {
		appErr = candidate
	}
	errorCode := ""
	if appErr.Message == "account_exists_multiple_tenants" {
		errorCode = appErr.Message
	}
	appErr.Message = LocalizeMessage(appErr.Message)
	body := gin.H{
		"code":    appErr.Code,
		"message": appErr.Message,
		"data":    nil,
	}
	if errorCode != "" {
		body["error"] = errorCode
	}
	c.JSON(httpStatus(appErr.Code), body)
}

func httpStatus(code int) int {
	switch code {
	case 40000:
		return http.StatusBadRequest
	case 40100:
		return http.StatusUnauthorized
	case 40300:
		return http.StatusForbidden
	case 40400:
		return http.StatusNotFound
	case 40900:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
