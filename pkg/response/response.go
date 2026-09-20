package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperr "banking/pkg/errors"
	"banking/pkg/utils"
)

// SuccessBody adalah struktur standar untuk response sukses.
type SuccessBody struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PaginatedBody adalah struktur standar untuk response sukses dengan pagination.
type PaginatedBody struct {
	Success    bool              `json:"success"`
	Message    string            `json:"message"`
	Data       any               `json:"data"`
	Pagination *utils.Pagination `json:"pagination"`
}

// ErrorBody adalah struktur standar untuk response error.
type ErrorBody struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Error   string            `json:"error,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// Success mengirimkan response sukses dengan status HTTP, pesan, dan data yang diberikan.
func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, SuccessBody{Success: true, Message: message, Data: data})
}


func OK(c *gin.Context, message string, data any) {
	Success(c, http.StatusOK, message, data)
}

func Created(c *gin.Context, message string, data any) {
	Success(c, http.StatusCreated, message, data)
}

func Paginated(c *gin.Context, message string, data any, p *utils.Pagination) {
	c.JSON(http.StatusOK, PaginatedBody{Success: true, Message: message, Data: data, Pagination: p})
}

// Error memetakan AppError menjadi HTTP response yang konsisten.
func Error(c *gin.Context, err error) {
	appErr := apperr.As(err)
	c.AbortWithStatusJSON(appErr.Status, ErrorBody{
		Success: false,
		Message: appErr.Message,
		Error:   appErr.Code,
	})
}

func ValidationError(c *gin.Context, errs map[string]string) {
	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, ErrorBody{
		Success: false,
		Message: "Validation failed",
		Error:   "VALIDATION_ERROR",
		Errors:  errs,
	})
}

func BadRequest(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, ErrorBody{
		Success: false,
		Message: message,
		Error:   "BAD_REQUEST",
	})
}
