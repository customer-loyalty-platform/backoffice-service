package middleware

import (
	"fmt"
	"net/http"

	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				errStr := fmt.Sprintf("%v", r)
				log.Error("panic recovered in HTTP handler",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("error", errStr),
				)
				response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected server error occurred")
				c.Abort()
			}
		}()
		c.Next()
	}
}
