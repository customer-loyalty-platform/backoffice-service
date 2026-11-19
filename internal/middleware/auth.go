package middleware

import (
	"strings"

	pkgJWT "backoffice-service/internal/pkg/jwt"
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	CtxAdminID     = "admin_id"
	CtxUsername    = "admin_username"
	CtxEmail       = "admin_email"
	CtxRole        = "admin_role"
	CtxPermissions = "admin_permissions"
)

func Auth(jwtManager *pkgJWT.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization token is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "Invalid Authorization header format. Format: Bearer <token>")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := jwtManager.VerifyToken(tokenStr)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token: "+err.Error())
			c.Abort()
			return
		}

		c.Set(CtxAdminID, claims.AdminID)
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxEmail, claims.Email)
		c.Set(CtxRole, claims.Role)
		c.Set(CtxPermissions, claims.Permissions)

		c.Next()
	}
}
