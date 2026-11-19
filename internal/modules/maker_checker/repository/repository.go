package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/maker_checker/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MakerCheckerRepository interface {
	FindPending(ctx context.Context, pagination domain.PaginationQuery, status string) ([]model.PendingAction, int64, error)
	FindByID(ctx context.Context, id string) (*model.PendingAction, error)
	Approve(ctx context.Context, id, checkerID string) error
	Reject(ctx context.Context, id, checkerID, reason string) error
}

type makerCheckerRepository struct {
	pool *pgxpool.Pool
}

func NewMakerCheckerRepository(pool *pgxpool.Pool) MakerCheckerRepository {
	return &makerCheckerRepository{pool: pool}
}

func (r *makerCheckerRepository) FindPending(ctx context.Context, pagination domain.PaginationQuery, status string) ([]model.PendingAction, int64, error) {
	whereClauses := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		whereClauses += fmt.Sprintf(" AND p.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM schema_backoffice.admin_pending_actions p %s", whereClauses)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT 
			p.id, p.action_type, p.target_resource, p.payload, p.status, 
			p.checker_note, p.created_at, p.resolved_at,
			m.id, m.username, m.role,
			COALESCE(c.id::text, ''), COALESCE(c.username, '')
		FROM schema_backoffice.admin_pending_actions p
		JOIN schema_backoffice.admin_users m ON p.maker_id = m.id
		LEFT JOIN schema_backoffice.admin_users c ON p.checker_id = c.id
		%s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClauses, argIdx, argIdx+1)

	args = append(args, pagination.Limit, pagination.Offset())

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.PendingAction
	for rows.Next() {
		var item model.PendingAction
		var payloadBytes []byte
		var checkerID, checkerUsername string

		err := rows.Scan(
			&item.ID,
			&item.ActionType,
			&item.TargetResource,
			&payloadBytes,
			&item.Status,
			&item.RejectionReason,
			&item.CreatedAt,
			&item.ResolvedAt,
			&item.Maker.ID,
			&item.Maker.Username,
			&item.Maker.Role,
			&checkerID,
			&checkerUsername,
		)
		if err != nil {
			return nil, 0, err
		}

		if len(payloadBytes) > 0 {
			_ = json.Unmarshal(payloadBytes, &item.Payload)
		}

		if checkerID != "" {
			item.Checker = &model.CheckerInfo{
				ID:       checkerID,
				Username: checkerUsername,
			}
		}

		items = append(items, item)
	}

	return items, total, nil
}

func (r *makerCheckerRepository) FindByID(ctx context.Context, id string) (*model.PendingAction, error) {
	query := `
		SELECT 
			p.id, p.action_type, p.target_resource, p.payload, p.status, 
			p.checker_note, p.created_at, p.resolved_at,
			m.id, m.username, m.role,
			COALESCE(c.id::text, ''), COALESCE(c.username, '')
		FROM schema_backoffice.admin_pending_actions p
		JOIN schema_backoffice.admin_users m ON p.maker_id = m.id
		LEFT JOIN schema_backoffice.admin_users c ON p.checker_id = c.id
		WHERE p.id = $1 LIMIT 1`

	var item model.PendingAction
	var payloadBytes []byte
	var checkerID, checkerUsername string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.ActionType,
		&item.TargetResource,
		&payloadBytes,
		&item.Status,
		&item.RejectionReason,
		&item.CreatedAt,
		&item.ResolvedAt,
		&item.Maker.ID,
		&item.Maker.Username,
		&item.Maker.Role,
		&checkerID,
		&checkerUsername,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if len(payloadBytes) > 0 {
		_ = json.Unmarshal(payloadBytes, &item.Payload)
	}

	if checkerID != "" {
		item.Checker = &model.CheckerInfo{
			ID:       checkerID,
			Username: checkerUsername,
		}
	}

	return &item, nil
}

func (r *makerCheckerRepository) Approve(ctx context.Context, id, checkerID string) error {
	query := `
		UPDATE schema_backoffice.admin_pending_actions 
		SET status = 'APPROVED', checker_id = $1, resolved_at = NOW() 
		WHERE id = $2 AND status = 'PENDING'`

	res, err := r.pool.Exec(ctx, query, checkerID, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("action not found or already resolved")
	}
	return nil
}

func (r *makerCheckerRepository) Reject(ctx context.Context, id, checkerID, reason string) error {
	query := `
		UPDATE schema_backoffice.admin_pending_actions 
		SET status = 'REJECTED', checker_id = $1, checker_note = $2, resolved_at = NOW() 
		WHERE id = $3 AND status = 'PENDING'`

	res, err := r.pool.Exec(ctx, query, checkerID, reason, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("action not found or already resolved")
	}
	return nil
}
