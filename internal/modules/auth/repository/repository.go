package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/auth/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository interface {
	FindByUsername(ctx context.Context, username string) (*model.AdminUser, error)
	FindByID(ctx context.Context, id string) (*model.AdminUser, error)
	UpdateLastLogin(ctx context.Context, id string) error
}

type authRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) AuthRepository {
	return &authRepository{pool: pool}
}

func (r *authRepository) FindByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, permissions, is_active, last_login_at, created_at, updated_at
		FROM schema_backoffice.admin_users
		WHERE username = $1 LIMIT 1`

	var user model.AdminUser
	var permsBytes []byte

	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&permsBytes,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if len(permsBytes) > 0 {
		_ = json.Unmarshal(permsBytes, &user.Permissions)
	}

	return &user, nil
}

func (r *authRepository) FindByID(ctx context.Context, id string) (*model.AdminUser, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, permissions, is_active, last_login_at, created_at, updated_at
		FROM schema_backoffice.admin_users
		WHERE id = $1 LIMIT 1`

	var user model.AdminUser
	var permsBytes []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&permsBytes,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if len(permsBytes) > 0 {
		_ = json.Unmarshal(permsBytes, &user.Permissions)
	}

	return &user, nil
}

func (r *authRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `
		UPDATE schema_backoffice.admin_users
		SET last_login_at = $1, updated_at = $1
		WHERE id = $2`

	_, err := r.pool.Exec(ctx, query, time.Now(), id)
	return err
}

