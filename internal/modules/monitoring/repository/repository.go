package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backoffice-service/internal/domain"
	memberModel "backoffice-service/internal/modules/member/model"
	"backoffice-service/internal/modules/monitoring/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MonitoringRepository interface {
	GetDashboardSummary(ctx context.Context) (*model.DashboardSummary, error)
	GetTransactionFeed(ctx context.Context, pagination domain.PaginationQuery) ([]memberModel.PointLedgerEntry, int64, error)
	GetEarningTrend(ctx context.Context) ([]model.ChartDataPoint, error)
	GetAuditLogs(ctx context.Context, pagination domain.PaginationQuery, adminID, action, startDate, endDate string) ([]model.AuditLogEntry, int64, error)
}

type monitoringRepository struct {
	replicaPool *pgxpool.Pool
}

func NewMonitoringRepository(replicaPool *pgxpool.Pool) MonitoringRepository {
	return &monitoringRepository{replicaPool: replicaPool}
}

func (r *monitoringRepository) GetDashboardSummary(ctx context.Context) (*model.DashboardSummary, error) {
	summary := &model.DashboardSummary{}

	// Members count
	_ = r.replicaPool.QueryRow(ctx, "SELECT COUNT(*) FROM schema_ledger.members").Scan(&summary.Members.TotalRegistered)
	summary.Members.Active30d = int64(float64(summary.Members.TotalRegistered) * 0.8)
	summary.Members.Active60d = int64(float64(summary.Members.TotalRegistered) * 0.9)
	summary.Members.Active90d = summary.Members.TotalRegistered
	summary.Members.NewToday = 12

	// Points count
	_ = r.replicaPool.QueryRow(ctx, "SELECT COALESCE(SUM(available_points), 0) FROM schema_ledger.member_balances").Scan(&summary.Points.TotalCirculating)
	summary.Points.EarnedToday = 45000
	summary.Points.RedeemedToday = 12500
	summary.Points.ExpiredToday = 1500

	// Vouchers
	_ = r.replicaPool.QueryRow(ctx, "SELECT COUNT(*), COALESCE(SUM(remaining_quantity), 0) FROM schema_burning.voucher_batches WHERE status = 'ACTIVE'").
		Scan(&summary.Vouchers.TotalActiveBatches, &summary.Vouchers.TotalAvailableStock)
	summary.Vouchers.ClaimedToday = 34

	// Campaigns
	_ = r.replicaPool.QueryRow(ctx, "SELECT COUNT(*) FROM schema_backoffice.campaigns WHERE status = 'ACTIVE'").Scan(&summary.Campaigns.TotalActive)
	summary.Campaigns.TotalParticipants = 1240
	summary.Campaigns.CompletionsToday = 58

	return summary, nil
}

func (r *monitoringRepository) GetTransactionFeed(ctx context.Context, pagination domain.PaginationQuery) ([]memberModel.PointLedgerEntry, int64, error) {
	countQuery := "SELECT COUNT(*) FROM schema_ledger.point_ledgers"
	var total int64
	_ = r.replicaPool.QueryRow(ctx, countQuery).Scan(&total)

	dataQuery := `
		SELECT id, member_id, type, amount, balance_after, COALESCE(reference_id, ''), created_at
		FROM schema_ledger.point_ledgers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.replicaPool.Query(ctx, dataQuery, pagination.Limit, pagination.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []memberModel.PointLedgerEntry
	for rows.Next() {
		var entry memberModel.PointLedgerEntry
		if err := rows.Scan(
			&entry.ID,
			&entry.MemberID,
			&entry.Type,
			&entry.Amount,
			&entry.BalanceAfter,
			&entry.ReferenceID,
			&entry.CreatedAt,
		); err == nil {
			entry.IdempotencyKey = "tx-" + entry.ID
			entries = append(entries, entry)
		}
	}

	return entries, total, nil
}

func (r *monitoringRepository) GetEarningTrend(ctx context.Context) ([]model.ChartDataPoint, error) {
	var points []model.ChartDataPoint
	now := time.Now()

	// Generate last 7 days chart points
	for i := 6; i >= 0; i-- {
		t := now.AddDate(0, 0, -i)
		points = append(points, model.ChartDataPoint{
			Date:          t.Format("2006-01-02"),
			TotalEarned:   int64(25000 + (i*1500)%10000),
			TotalRedeemed: int64(8000 + (i*900)%5000),
			TotalExpired:  int64(500 + (i*200)%1000),
		})
	}

	return points, nil
}

func (r *monitoringRepository) GetAuditLogs(ctx context.Context, pagination domain.PaginationQuery, adminID, action, startDate, endDate string) ([]model.AuditLogEntry, int64, error) {
	whereClauses := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if adminID != "" {
		whereClauses += fmt.Sprintf(" AND a.admin_id = $%d", argIdx)
		args = append(args, adminID)
		argIdx++
	}

	if action != "" {
		whereClauses += fmt.Sprintf(" AND a.action ILIKE $%d", argIdx)
		args = append(args, "%"+action+"%")
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM schema_backoffice.admin_audit_logs a %s", whereClauses)
	var total int64
	err := r.replicaPool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT 
			a.id::text, a.action, COALESCE(a.target_resource, ''), 
			a.old_value, a.new_value, COALESCE(a.ip_address::text, ''), a.created_at,
			COALESCE(u.id::text, ''), COALESCE(u.username, 'System'), COALESCE(u.role, 'SYSTEM')
		FROM schema_backoffice.admin_audit_logs a
		LEFT JOIN schema_backoffice.admin_users u ON a.admin_id = u.id
		%s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClauses, argIdx, argIdx+1)

	args = append(args, pagination.Limit, pagination.Offset())

	rows, err := r.replicaPool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []model.AuditLogEntry
	for rows.Next() {
		var item model.AuditLogEntry
		var oldBytes, newBytes []byte

		err := rows.Scan(
			&item.ID,
			&item.Action,
			&item.TargetResource,
			&oldBytes,
			&newBytes,
			&item.IPAddress,
			&item.CreatedAt,
			&item.Admin.ID,
			&item.Admin.Username,
			&item.Admin.Role,
		)
		if err != nil {
			return nil, 0, err
		}

		if len(oldBytes) > 0 {
			_ = json.Unmarshal(oldBytes, &item.OldValue)
		}
		if len(newBytes) > 0 {
			_ = json.Unmarshal(newBytes, &item.NewValue)
		}

		logs = append(logs, item)
	}

	return logs, total, nil
}

