package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/auth/model"
	"backoffice-service/internal/modules/auth/repository"
	pkgJWT "backoffice-service/internal/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

var defaultRolePermissions = map[string][]string{
	"SUPER_ADMIN": {"*"},
	"CAMPAIGN_SPEC": {
		"dashboard",
		"members:read",
		"tiers", "tiers:read", "tiers:write",
		"rules", "rules:read", "rules:write",
		"vouchers", "vouchers:read", "vouchers:write",
		"campaigns", "campaigns:read", "campaigns:write",
		"partners:read",
		"settlement:read",
	},
	"OPS_CS": {
		"dashboard",
		"members", "members:read", "members:write",
		"tiers:read",
		"rules:read",
		"vouchers:read",
		"approvals",
	},
	"FINANCE": {
		"dashboard",
		"members:read",
		"tiers:read",
		"rules:read",
		"vouchers:read",
		"settlement", "settlement:read", "settlement:write",
		"audit-logs", "audit-logs:read",
	},
}

type AuthService interface {
	Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error)
	RefreshToken(ctx context.Context, req model.RefreshTokenRequest) (*model.RefreshTokenResponse, error)
	GetMe(ctx context.Context, adminID string) (*model.MeResponse, error)
}

type authService struct {
	repo        repository.AuthRepository
	jwtManager  *pkgJWT.JWTManager
	tokenExpiry time.Duration
}

func NewAuthService(repo repository.AuthRepository, jwtManager *pkgJWT.JWTManager, tokenExpiry time.Duration) AuthService {
	return &authService{
		repo:        repo,
		jwtManager:  jwtManager,
		tokenExpiry: tokenExpiry,
	}
}

func (s *authService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, errors.New("admin account is inactive")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		fmt.Println(err)
		return nil, errors.New("invalid username or password")
	}

	permissions := user.Permissions
	if len(permissions) == 0 {
		permissions = defaultRolePermissions[user.Role]
	}

	accessToken, refreshToken, err := s.jwtManager.GenerateTokens(
		user.ID,
		user.Username,
		user.Email,
		user.Role,
		permissions,
	)
	if err != nil {
		return nil, err
	}

	_ = s.repo.UpdateLastLogin(ctx, user.ID)

	status := "ACTIVE"
	if !user.IsActive {
		status = "INACTIVE"
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenExpiry.Seconds()),
		Admin: model.AdminSummary{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
			Status:   status,
		},
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req model.RefreshTokenRequest) (*model.RefreshTokenResponse, error) {
	claims, err := s.jwtManager.VerifyToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.repo.FindByID(ctx, claims.AdminID)
	if err != nil {
		return nil, errors.New("admin user not found")
	}

	if !user.IsActive {
		return nil, errors.New("admin user is inactive")
	}

	permissions := user.Permissions
	if len(permissions) == 0 {
		permissions = defaultRolePermissions[user.Role]
	}

	newAccessToken, newRefreshToken, err := s.jwtManager.GenerateTokens(
		user.ID,
		user.Username,
		user.Email,
		user.Role,
		permissions,
	)
	if err != nil {
		return nil, err
	}

	return &model.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenExpiry.Seconds()),
	}, nil
}

func (s *authService) GetMe(ctx context.Context, adminID string) (*model.MeResponse, error) {
	user, err := s.repo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	permissions := user.Permissions
	if len(permissions) == 0 {
		permissions = defaultRolePermissions[user.Role]
	}

	status := "ACTIVE"
	if !user.IsActive {
		status = "INACTIVE"
	}

	return &model.MeResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Status:      status,
		Permissions: permissions,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}
