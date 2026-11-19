package repository

import (
	"context"
	"errors"
	"time"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/admin_user/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminUserRepository interface {
	FindAll(ctx context.Context) ([]model.AdminUserItem, int64, error)
	FindByID(ctx context.Context, id string) (*model.AdminUserItem, error)
	Create(ctx context.Context, u *model.AdminUserItem, passwordHash string) error
	Update(ctx context.Context, id string, req model.UpdateAdminUserRequest) (*model.AdminUserItem, error)
}

type adminUserRepository struct {
	pool *pgxpool.Pool
}

func NewAdminUserRepository(pool *pgxpool.Pool) AdminUserRepository {
	return &adminUserRepository{pool: pool}
}

func (r *adminUserRepository) FindAll(ctx context.Context) ([]model.AdminUserItem, int64, error) {
	query := `
		SELECT id, username, email, role, is_active, created_at, updated_at
		FROM schema_backoffice.admin_users
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.AdminUserItem
	for rows.Next() {
		var u model.AdminUserItem
		var isActive bool
		err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.Role,
			&isActive,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if isActive {
			u.Status = "ACTIVE"
		} else {
			u.Status = "INACTIVE"
		}
		users = append(users, u)
	}

	return users, int64(len(users)), nil
}

func (r *adminUserRepository) FindByID(ctx context.Context, id string) (*model.AdminUserItem, error) {
	query := `
		SELECT id, username, email, role, is_active, created_at, updated_at
		FROM schema_backoffice.admin_users
		WHERE id = $1 LIMIT 1`

	var u model.AdminUserItem
	var isActive bool
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Role,
		&isActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if isActive {
		u.Status = "ACTIVE"
	} else {
		u.Status = "INACTIVE"
	}

	return &u, nil
}

func (r *adminUserRepository) Create(ctx context.Context, u *model.AdminUserItem, passwordHash string) error {
	query := `
		INSERT INTO schema_backoffice.admin_users 
		(username, email, password_hash, full_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, TRUE, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	fullName := u.Username
	return r.pool.QueryRow(
		ctx, query,
		u.Username, u.Email, passwordHash, fullName, u.Role,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *adminUserRepository) Update(ctx context.Context, id string, req model.UpdateAdminUserRequest) (*model.AdminUserItem, error) {
	currentUser, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	email := currentUser.Email
	if req.Email != nil {
		email = *req.Email
	}

	role := currentUser.Role
	if req.Role != nil {
		role = *req.Role
	}

	isActive := currentUser.Status == "ACTIVE"
	if req.Status != nil {
		isActive = *req.Status == "ACTIVE"
	}

	query := `
		UPDATE schema_backoffice.admin_users 
		SET email = $1, role = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4`

	res, err := r.pool.Exec(ctx, query, email, role, isActive, id)
	if err != nil {
		return nil, err
	}
	if res.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	currentUser.Email = email
	currentUser.Role = role
	if isActive {
		currentUser.Status = "ACTIVE"
	} else {
		currentUser.Status = "INACTIVE"
	}
	currentUser.UpdatedAt = time.Now()

	return currentUser, nil
}

