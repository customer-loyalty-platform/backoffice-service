package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func AuditLogger(masterPool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only audit state-modifying requests
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// Restore the body for handler
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		// If operation was successful (2xx status)
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			adminIDVal, _ := c.Get(CtxAdminID)
			adminID, _ := adminIDVal.(string)

			action := c.Request.Method + " " + c.FullPath()
			clientIP := c.ClientIP()
			resource := c.Request.URL.Path

			var payload interface{}
			if len(bodyBytes) > 0 {
				payload = string(bodyBytes)
			}

			// Fire and forget asynchronous audit log insert
			go func(aID, act, res, ip string, pl interface{}) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				query := `
					INSERT INTO schema_backoffice.admin_audit_logs 
					(admin_id, action, target_resource, new_value, ip_address, created_at)
					VALUES ($1, $2, $3, $4, $5, NOW())`

				var adminIDParam interface{}
				if aID != "" {
					adminIDParam = aID
				}

				_, err := masterPool.Exec(ctx, query, adminIDParam, act, res, pl, ip)
				if err != nil {
					log.Warn("failed to record async audit log",
						zap.String("action", act),
						zap.String("resource", res),
						zap.Error(err),
					)
				}
			}(adminID, action, resource, clientIP, payload)
		}
	}
}

