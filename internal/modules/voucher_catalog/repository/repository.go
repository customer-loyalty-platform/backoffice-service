package repository

import (
	"context"
	"errors"
	"fmt"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/voucher_catalog/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VoucherRepository interface {
	FindBatches(ctx context.Context, pagination domain.PaginationQuery) ([]model.VoucherBatch, int64, error)
	FindBatchByID(ctx context.Context, id string) (*model.VoucherBatch, error)
	CreateBatch(ctx context.Context, b *model.VoucherBatch) error
	ImportCodes(ctx context.Context, batchID string, codes []string) (int, int, error)
	FindInventoryByBatch(ctx context.Context, batchID string, pagination domain.PaginationQuery, status string) ([]model.VoucherInventoryItem, int64, error)
	RevokeVouchers(ctx context.Context, batchID string, voucherIDs []string, reason string) error
}

type voucherRepository struct {
	pool *pgxpool.Pool
}

func NewVoucherRepository(pool *pgxpool.Pool) VoucherRepository {
	return &voucherRepository{pool: pool}
}

func (r *voucherRepository) FindBatches(ctx context.Context, pagination domain.PaginationQuery) ([]model.VoucherBatch, int64, error) {
	whereClauses := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if pagination.Search != "" {
		whereClauses += fmt.Sprintf(" AND (batch_code ILIKE $%d OR title ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+pagination.Search+"%")
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM schema_burning.voucher_batches %s", whereClauses)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT 
			id, batch_code, title, points_cost,
			total_stock, available_stock,
			(total_stock - available_stock) as claimed, 0 as revoked,
			created_at
		FROM schema_burning.voucher_batches
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClauses, argIdx, argIdx+1)

	args = append(args, pagination.Limit, pagination.Offset())

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var batches []model.VoucherBatch
	for rows.Next() {
		var item model.VoucherBatch
		var ptsCost float64
		err := rows.Scan(
			&item.ID,
			&item.BatchCode,
			&item.Title,
			&ptsCost,
			&item.TotalStock,
			&item.AvailableStock,
			&item.ClaimedStock,
			&item.RevokedCount,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		item.PointsCost = int64(ptsCost)
		item.Status = "ACTIVE"
		if item.AvailableStock == 0 {
			item.Status = "DEPLETED"
		}
		batches = append(batches, item)
	}

	return batches, total, nil
}

func (r *voucherRepository) FindBatchByID(ctx context.Context, id string) (*model.VoucherBatch, error) {
	query := `
		SELECT 
			id, batch_code, title, points_cost,
			total_stock, available_stock,
			(total_stock - available_stock) as claimed, 0 as revoked,
			created_at
		FROM schema_burning.voucher_batches
		WHERE id = $1 LIMIT 1`

	var item model.VoucherBatch
	var ptsCost float64
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.BatchCode,
		&item.Title,
		&ptsCost,
		&item.TotalStock,
		&item.AvailableStock,
		&item.ClaimedStock,
		&item.RevokedCount,
		&item.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	item.PointsCost = int64(ptsCost)
	item.Status = "ACTIVE"
	if item.AvailableStock == 0 {
		item.Status = "DEPLETED"
	}

	return &item, nil
}

func (r *voucherRepository) CreateBatch(ctx context.Context, b *model.VoucherBatch) error {
	query := `
		INSERT INTO schema_burning.voucher_batches 
		(batch_code, title, points_cost, total_stock, available_stock, created_at)
		VALUES ($1, $2, $3, 0, 0, NOW())
		RETURNING id, created_at`

	return r.pool.QueryRow(
		ctx, query,
		b.BatchCode, b.Title, b.PointsCost,
	).Scan(&b.ID, &b.CreatedAt)
}

func (r *voucherRepository) ImportCodes(ctx context.Context, batchID string, codes []string) (int, int, error) {
	imported := 0
	skipped := 0

	for _, code := range codes {
		query := `
			INSERT INTO schema_burning.voucher_inventories 
			(batch_id, voucher_code, status)
			VALUES ($1, $2, 'AVAILABLE')
			ON CONFLICT DO NOTHING`

		tag, err := r.pool.Exec(ctx, query, batchID, code)
		if err != nil {
			skipped++
			continue
		}
		if tag.RowsAffected() > 0 {
			imported++
		} else {
			skipped++
		}
	}

	// Update batch quantities
	updateBatch := `
		UPDATE schema_burning.voucher_batches 
		SET total_stock = total_stock + $1,
		    available_stock = available_stock + $1
		WHERE id = $2`

	_, _ = r.pool.Exec(ctx, updateBatch, imported, batchID)

	return imported, skipped, nil
}

func (r *voucherRepository) FindInventoryByBatch(ctx context.Context, batchID string, pagination domain.PaginationQuery, status string) ([]model.VoucherInventoryItem, int64, error) {
	whereClauses := "WHERE batch_id = $1"
	args := []interface{}{batchID}
	argIdx := 2

	if status != "" {
		whereClauses += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM schema_burning.voucher_inventories %s", whereClauses)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, batch_id, voucher_code, COALESCE(status, 'AVAILABLE'), claimed_by::text, claimed_at::text
		FROM schema_burning.voucher_inventories
		%s
		ORDER BY id DESC
		LIMIT $%d OFFSET $%d`, whereClauses, argIdx, argIdx+1)

	args = append(args, pagination.Limit, pagination.Offset())

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.VoucherInventoryItem
	for rows.Next() {
		var item model.VoucherInventoryItem
		err := rows.Scan(
			&item.ID,
			&item.BatchID,
			&item.VoucherCode,
			&item.Status,
			&item.ClaimedBy,
			&item.ClaimedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, nil
}

func (r *voucherRepository) RevokeVouchers(ctx context.Context, batchID string, voucherIDs []string, reason string) error {
	query := `
		UPDATE schema_burning.voucher_inventories 
		SET status = 'REVOKED' 
		WHERE batch_id = $1 AND id = ANY($2)`

	_, err := r.pool.Exec(ctx, query, batchID, voucherIDs)
	return err
}
