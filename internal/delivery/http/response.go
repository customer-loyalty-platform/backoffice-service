package http

import (
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type StandardResponse = response.StandardResponse
type ErrorBody = response.ErrorBody
type PaginationBody = response.PaginationBody

func Success(c *gin.Context, httpStatus int, data interface{}) {
	response.Success(c, httpStatus, data)
}

func SuccessWithMeta(c *gin.Context, httpStatus int, data interface{}, pagination *PaginationBody) {
    response.SuccessWithMeta(c, httpStatus, data, pagination)
}

func Error(c *gin.Context, httpStatus int, code string, message string) {
	response.Error(c, httpStatus, code, message)
}

func BadRequest(c *gin.Context, message string) {
	response.BadRequest(c, message)
}

func Unauthorized(c *gin.Context, message string) {
	response.Unauthorized(c, message)
}

func Forbidden(c *gin.Context, message string) {
	response.Forbidden(c, message)
}

func NotFound(c *gin.Context, message string) {
	response.NotFound(c, message)
}

func InternalError(c *gin.Context, message string) {
	response.InternalError(c, message)
}
