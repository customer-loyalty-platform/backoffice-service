package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/member/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MemberRepository interface {
	FindAll(ctx context.Context, pagination domain.PaginationQuery, statusFilter string, tierFilter string) ([]model.MemberListItem, int64, error)
	FindByID(ctx context.Context, id string) (*model.MemberDetail, error)
	UpdateStatus(ctx context.Context, id string, newStatus string) (string, error)
	CreatePendingAction(ctx context.Context, actionType, resource, targetID, makerID string, payload interface{}) (string, error)
	FindTransactionsByMemberID(ctx context.Context, memberID string, pagination domain.PaginationQuery) ([]model.PointLedgerEntry, int64, error)
}

type memberRepository struct {
	master  *pgxpool.Pool
	replica *pgxpool.Pool
}

func NewMemberRepository(master, replica *pgxpool.Pool) MemberRepository {
	return &memberRepository{master: master, replica: replica}
}

func (r *memberRepository) FindAll(ctx context.Context, pagination domain.PaginationQuery, statusFilter string, tierFilter string) ([]model.MemberListItem, int64, error) {
	whereClauses := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if pagination.Search != "" {
		whereClauses += fmt.Sprintf(" AND (m.account_number ILIKE $%d)", argIdx)
		args = append(args, "%"+pagination.Search+"%")
		argIdx++
	}

	if statusFilter != "" {
		whereClauses += fmt.Sprintf(" AND m.status = $%d", argIdx)
		args = append(args, statusFilter)
		argIdx++
	}

	if tierFilter != "" {
		whereClauses += fmt.Sprintf(" AND t.tier_code = $%d", argIdx)
		args = append(args, tierFilter)
		argIdx++
	}

	// Count query
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM schema_ledger.members m
		LEFT JOIN schema_backoffice.tier_definitions t ON m.current_tier_id = t.id
		%s`, whereClauses)

	var total int64
	err := r.replica.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT 
			m.id, m.account_number, COALESCE(m.status, 'ACTIVE'), 
			COALESCE(t.tier_code, 'BRONZE'), COALESCE(t.tier_name, 'Bronze Tier'), COALESCE(t.sort_order, 1),
			COALESCE(m.created_at, NOW()), COALESCE(m.updated_at, NOW()),
			COALESCE(b.total_points, 0), COALESCE(b.locked_points, 0), COALESCE(b.expiring_points, 0)
		FROM schema_ledger.members m
		LEFT JOIN schema_ledger.member_balances b ON m.id = b.member_id
		LEFT JOIN schema_backoffice.tier_definitions t ON m.current_tier_id = t.id
		%s
		ORDER BY m.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClauses, argIdx, argIdx+1)

	args = append(args, pagination.Limit, pagination.Offset())

	rows, err := r.replica.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.MemberListItem
	for rows.Next() {
		var item model.MemberListItem
		var tierCode, tierName string
		var rank int
		var totalPts, lockedPts, expPts float64
		err := rows.Scan(
			&item.ID,
			&item.AccountNumber,
			&item.Status,
			&tierCode,
			&tierName,
			&rank,
			&item.CreatedAt,
			&item.UpdatedAt,
			&totalPts,
			&lockedPts,
			&expPts,
		)
		if err != nil {
			return nil, 0, err
		}

		item.FullName = "Member " + item.AccountNumber
		item.Email = item.AccountNumber + "@loyalty.local"
		item.Phone = "+6281" + item.AccountNumber[len(item.AccountNumber)-7:]
		item.CurrentTier = model.CurrentTier{
			ID:       tierCode,
			TierName: tierName,
			Rank:     rank,
		}
		item.Balance.TotalPoints = int64(totalPts)
		item.Balance.LockedPoints = int64(lockedPts)
		item.Balance.ExpiringPoints = int64(expPts)

		items = append(items, item)
	}

	return items, total, nil
}

func (r *memberRepository) FindByID(ctx context.Context, id string) (*model.MemberDetail, error) {
	query := `
		SELECT 
			m.id, m.account_number, COALESCE(m.status, 'ACTIVE'), 
			COALESCE(t.tier_code, 'BRONZE'), COALESCE(t.tier_name, 'Bronze Tier'), COALESCE(t.sort_order, 1),
			COALESCE(m.created_at, NOW()), COALESCE(m.updated_at, NOW()),
			COALESCE(b.total_points, 0), COALESCE(b.locked_points, 0), COALESCE(b.expiring_points, 0)
		FROM schema_ledger.members m
		LEFT JOIN schema_ledger.member_balances b ON m.id = b.member_id
		LEFT JOIN schema_backoffice.tier_definitions t ON m.current_tier_id = t.id
		WHERE m.id = $1 LIMIT 1`

	var detail model.MemberDetail
	var tierCode, tierName string
	var rank int
	var totalPts, lockedPts, expPts float64

	err := r.replica.QueryRow(ctx, query, id).Scan(
		&detail.ID,
		&detail.AccountNumber,
		&detail.Status,
		&tierCode,
		&tierName,
		&rank,
		&detail.CreatedAt,
		&detail.UpdatedAt,
		&totalPts,
		&lockedPts,
		&expPts,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	detail.FullName = "Member " + detail.AccountNumber
	detail.Email = detail.AccountNumber + "@loyalty.local"
	detail.Phone = "+6281" + detail.AccountNumber[len(detail.AccountNumber)-7:]
	detail.CurrentTier = model.CurrentTier{
		ID:       tierCode,
		TierName: tierName,
		Rank:     rank,
	}
	detail.Balance.TotalPoints = int64(totalPts)
	detail.Balance.LockedPoints = int64(lockedPts)
	detail.Balance.ExpiringPoints = int64(expPts)
	detail.LifetimeEarned = detail.Balance.TotalPoints
	detail.LifetimeRedeemed = 0
	detail.TierQualifyingPoints = detail.Balance.TotalPoints
	detail.TierEvaluationDate = time.Now().AddDate(0, 6, 0).Format("2006-01-02")
	detail.LinkedAccounts = []model.LinkedAccount{
		{
			ID:                "la-" + detail.ID,
			AccountType:       "GOPAY",
			AccountIdentifier: detail.Phone,
			LinkedAt:          detail.CreatedAt.Format(time.RFC3339),
		},
	}

	// Fetch recent transactions
	txQuery := `
		SELECT id, member_id, type, amount, balance_after, COALESCE(reference_id, ''), COALESCE(idempotency_key, ''), created_at
		FROM schema_ledger.point_ledgers
		WHERE member_id = $1
		ORDER BY created_at DESC
		LIMIT 10`

	txRows, err := r.replica.Query(ctx, txQuery, id)
	if err == nil {
		defer txRows.Close()
		for txRows.Next() {
			var tx model.PointLedgerEntry
			var amt, bal float64
			if err := txRows.Scan(
				&tx.ID,
				&tx.MemberID,
				&tx.Type,
				&amt,
				&bal,
				&tx.ReferenceID,
				&tx.IdempotencyKey,
				&tx.CreatedAt,
			); err == nil {
				tx.Amount = int64(amt)
				tx.BalanceAfter = int64(bal)
				detail.RecentTransactions = append(detail.RecentTransactions, tx)
			}
		}
	}

	return &detail, nil
}

func (r *memberRepository) UpdateStatus(ctx context.Context, id string, newStatus string) (string, error) {
	var prevStatus string
	queryGet := `SELECT COALESCE(status, 'ACTIVE') FROM schema_ledger.members WHERE id = $1`
	err := r.master.QueryRow(ctx, queryGet, id).Scan(&prevStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", err
	}

	queryUpdate := `
		UPDATE schema_ledger.members 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2`

	_, err = r.master.Exec(ctx, queryUpdate, newStatus, id)
	if err != nil {
		return "", err
	}

	return prevStatus, nil
}

func (r *memberRepository) CreatePendingAction(ctx context.Context, actionType, resource, targetID, makerID string, payload interface{}) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	query := `
		INSERT INTO schema_backoffice.admin_pending_actions 
		(action_type, target_resource, target_id, payload, status, maker_id, created_at)
		VALUES ($1, $2, $3, $4, 'PENDING', $5, NOW())
		RETURNING id`

	var actionID string
	err = r.master.QueryRow(ctx, query, actionType, resource, targetID, payloadBytes, makerID).Scan(&actionID)
	if err != nil {
		return "", err
	}

	return actionID, nil
}

func (r *memberRepository) FindTransactionsByMemberID(ctx context.Context, memberID string, pagination domain.PaginationQuery) ([]model.PointLedgerEntry, int64, error) {
	countQuery := `SELECT COUNT(*) FROM schema_ledger.point_ledgers WHERE member_id = $1`
	var total int64
	_ = r.replica.QueryRow(ctx, countQuery, memberID).Scan(&total)

	dataQuery := `
		SELECT id, member_id, type, amount, balance_after, COALESCE(reference_id, ''), COALESCE(idempotency_key, ''), created_at
		FROM schema_ledger.point_ledgers
		WHERE member_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.replica.Query(ctx, dataQuery, memberID, pagination.Limit, pagination.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []model.PointLedgerEntry
	for rows.Next() {
		var entry model.PointLedgerEntry
		var amt, bal float64
		if err := rows.Scan(
			&entry.ID,
			&entry.MemberID,
			&entry.Type,
			&amt,
			&bal,
			&entry.ReferenceID,
			&entry.IdempotencyKey,
			&entry.CreatedAt,
		); err == nil {
			entry.Amount = int64(amt)
			entry.BalanceAfter = int64(bal)
			entries = append(entries, entry)
		}
	}

	return entries, total, nil
}
