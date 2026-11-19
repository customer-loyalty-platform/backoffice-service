package middleware

import (
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, _ := c.Get(CtxRole)
		role, _ := roleVal.(string)

		// SUPER_ADMIN has full access
		if role == "SUPER_ADMIN" {
			c.Next()
			return
		}

		permsVal, exists := c.Get(CtxPermissions)
		if !exists {
			response.Forbidden(c, "No permission context found in request")
			c.Abort()
			return
		}

		permissions, ok := permsVal.([]string)
		if !ok {
			response.Forbidden(c, "Malformed permission context")
			c.Abort()
			return
		}

		hasPerm := false
		for _, p := range permissions {
			if p == requiredPermission || p == "ALL_ACCESS" {
				hasPerm = true
				break
			}
		}

		if !hasPerm {
			response.Forbidden(c, "You do not have permission to execute this administrative action: "+requiredPermission)
			c.Abort()
			return
		}

		c.Next()
	}
}
