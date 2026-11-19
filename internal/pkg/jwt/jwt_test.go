package jwt_test

import (
	"testing"
	"time"

	pkgJWT "backoffice-service/internal/pkg/jwt"
)

func TestJWTManager_GenerateAndVerify(t *testing.T) {
	manager := pkgJWT.NewJWTManager("test-secret-key-123", 10*time.Minute, 1*time.Hour)

	adminID := "admin-uuid-001"
	username := "superadmin"
	email := "admin@loyalty.local"
	role := "SUPER_ADMIN"
	permissions := []string{"ALL_ACCESS"}

	accessToken, refreshToken, err := manager.GenerateTokens(adminID, username, email, role, permissions)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	if accessToken == "" || refreshToken == "" {
		t.Fatalf("Expected non-empty tokens")
	}

	claims, err := manager.VerifyToken(accessToken)
	if err != nil {
		t.Fatalf("Failed to verify access token: %v", err)
	}

	if claims.AdminID != adminID {
		t.Errorf("Expected AdminID %s, got %s", adminID, claims.AdminID)
	}
	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}
	if len(claims.Permissions) != 1 || claims.Permissions[0] != "ALL_ACCESS" {
		t.Errorf("Expected permissions [ALL_ACCESS], got %v", claims.Permissions)
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	manager := pkgJWT.NewJWTManager("test-secret-key-123", 10*time.Minute, 1*time.Hour)

	_, err := manager.VerifyToken("invalid.token.string")
	if err == nil {
		t.Errorf("Expected invalid token to fail verification")
	}
}

