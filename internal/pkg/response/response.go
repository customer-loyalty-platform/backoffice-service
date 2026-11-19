package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type StandardResponse struct {
	Status    string      `json:"status"`
	Timestamp string      `json:"timestamp"`
	Success   bool        `json:"success,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorBody  `json:"error,omitempty"`
	Pagination *PaginationBody `json:"pagination,omitempty"`
}


type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PaginationBody struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}


func Success(c *gin.Context, httpStatus int, data interface{}) {
	c.JSON(httpStatus, StandardResponse{
		Status:    "SUCCESS",
		Timestamp: time.Now().Format(time.RFC3339),
		Success:   true,
		Data:      data,
	})
}
	// No fallback needed – only the unified response structures are used.


func SuccessWithMeta(c *gin.Context, httpStatus int, data interface{}, pagination *PaginationBody) {
	c.JSON(httpStatus, StandardResponse{
		Status:    "SUCCESS",
		Timestamp: time.Now().Format(time.RFC3339),
		Success:   true,
		Data:      data,
		Pagination: pagination,
	})
}

func Error(c *gin.Context, httpStatus int, code string, message string) {
	c.JSON(httpStatus, StandardResponse{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

