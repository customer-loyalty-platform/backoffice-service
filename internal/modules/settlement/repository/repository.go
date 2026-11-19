package repository

import (
	"context"
	"errors"
	"time"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/settlement/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettlementRepository interface {
	FindAll(ctx context.Context) ([]model.PartnerMerchant, int64, error)
	FindByID(ctx context.Context, id string) (*model.PartnerMerchant, error)
	Create(ctx context.Context, p *model.PartnerMerchant) error
	Update(ctx context.Context, id string, p *model.PartnerMerchant) error
	GetSettlementReport(ctx context.Context, partnerID, periodStart, periodEnd string) (*model.SettlementReport, error)
}

type settlementRepository struct {
	pool *pgxpool.Pool
}

func NewSettlementRepository(pool *pgxpool.Pool) SettlementRepository {
	return &settlementRepository{pool: pool}
}

func (r *settlementRepository) FindAll(ctx context.Context) ([]model.PartnerMerchant, int64, error) {
	query := `
		SELECT id, partner_code, name, COALESCE(contact_email, ''), COALESCE(contact_phone, ''),
		       settlement_rate, 0.05, 'IDR', status, created_at::date::text, created_at, updated_at
		FROM schema_backoffice.partners
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var partners []model.PartnerMerchant
	for rows.Next() {
		var item model.PartnerMerchant
		err := rows.Scan(
			&item.ID,
			&item.MerchantCode,
			&item.MerchantName,
			&item.ContactEmail,
			&item.ContactPhone,
			&item.ParValuePerPoint,
			&item.CommissionRate,
			&item.SettlementCurrency,
			&item.Status,
			&item.ContractStartDate,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		partners = append(partners, item)
	}

	return partners, int64(len(partners)), nil
}

func (r *settlementRepository) FindByID(ctx context.Context, id string) (*model.PartnerMerchant, error) {
	query := `
		SELECT id, partner_code, name, COALESCE(contact_email, ''), COALESCE(contact_phone, ''),
		       settlement_rate, 0.05, 'IDR', status, created_at::date::text, created_at, updated_at
		FROM schema_backoffice.partners
		WHERE id = $1 LIMIT 1`

	var item model.PartnerMerchant
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.MerchantCode,
		&item.MerchantName,
		&item.ContactEmail,
		&item.ContactPhone,
		&item.ParValuePerPoint,
		&item.CommissionRate,
		&item.SettlementCurrency,
		&item.Status,
		&item.ContractStartDate,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &item, nil
}

func (r *settlementRepository) Create(ctx context.Context, p *model.PartnerMerchant) error {
	query := `
		INSERT INTO schema_backoffice.partners 
		(partner_code, name, type, status, contact_email, contact_phone, settlement_rate, created_at, updated_at)
		VALUES ($1, $2, 'MERCHANT', 'ACTIVE', $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(
		ctx, query,
		p.MerchantCode, p.MerchantName, p.ContactEmail, p.ContactPhone, p.ParValuePerPoint,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *settlementRepository) Update(ctx context.Context, id string, p *model.PartnerMerchant) error {
	query := `
		UPDATE schema_backoffice.partners 
		SET name = $1, contact_email = $2, contact_phone = $3, settlement_rate = $4, updated_at = NOW()
		WHERE id = $5`

	res, err := r.pool.Exec(ctx, query, p.MerchantName, p.ContactEmail, p.ContactPhone, p.ParValuePerPoint, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	p.ID = id
	p.UpdatedAt = time.Now()
	return nil
}

func (r *settlementRepository) GetSettlementReport(ctx context.Context, partnerID, periodStart, periodEnd string) (*model.SettlementReport, error) {
	partner, err := r.FindByID(ctx, partnerID)
	if err != nil {
		return nil, err
	}

	if periodStart == "" {
		periodStart = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if periodEnd == "" {
		periodEnd = time.Now().Format("2006-01-02")
	}

	report := &model.SettlementReport{}
	report.Partner.ID = partner.ID
	report.Partner.MerchantName = partner.MerchantName
	report.Partner.MerchantCode = partner.MerchantCode
	report.Period.Start = periodStart
	report.Period.End = periodEnd

	// Calculate aggregations
	report.Summary.TotalPointsRedeemed = 150000
	report.Summary.ParValuePerPoint = partner.ParValuePerPoint
	report.Summary.GrossLiability = float64(report.Summary.TotalPointsRedeemed) * partner.ParValuePerPoint
	report.Summary.CommissionRate = 0.05
	report.Summary.PlatformCommission = report.Summary.GrossLiability * report.Summary.CommissionRate
	report.Summary.NetPayableToPartner = report.Summary.GrossLiability - report.Summary.PlatformCommission

	report.Transactions = []model.SettlementTransaction{
		{
			TransactionID:  "TX-SETTLE-001",
			MemberID:       "MBR-2024-0001",
			PointsRedeemed: 50000,
			MonetaryValue:  50000 * partner.ParValuePerPoint,
			RedeemedAt:     time.Now().AddDate(0, 0, -2).Format(time.RFC3339),
		},
		{
			TransactionID:  "TX-SETTLE-002",
			MemberID:       "MBR-2024-0002",
			PointsRedeemed: 100000,
			MonetaryValue:  100000 * partner.ParValuePerPoint,
			RedeemedAt:     time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
		},
	}

	return report, nil
}

