package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/tier/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TierRepository interface {
	FindAll(ctx context.Context) ([]model.TierDefinition, error)
	FindByID(ctx context.Context, id string) (*model.TierDefinition, error)
	Create(ctx context.Context, t *model.TierDefinition, createdBy string) error
	Update(ctx context.Context, id string, t *model.TierDefinition) error
	Delete(ctx context.Context, id string) error
}

type tierRepository struct {
	pool *pgxpool.Pool
}

func NewTierRepository(pool *pgxpool.Pool) TierRepository {
	return &tierRepository{pool: pool}
}

func (r *tierRepository) FindAll(ctx context.Context) ([]model.TierDefinition, error) {
	query := `
		SELECT 
			t.id, t.tier_name, t.sort_order, 'POINTS_EARNED', t.min_points, t.multiplier, 
			t.perks, t.created_at, t.updated_at,
			(SELECT COUNT(*) FROM schema_ledger.members m WHERE m.current_tier_id  = t.id) as member_count
		FROM schema_backoffice.tier_definitions t
		ORDER BY t.sort_order ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []model.TierDefinition
	for rows.Next() {
		var item model.TierDefinition
		var perksBytes []byte
		err := rows.Scan(
			&item.ID,
			&item.TierName,
			&item.Rank,
			&item.QualificationType,
			&item.MinThreshold,
			&item.Multiplier,
			&perksBytes,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.MemberCount,
		)
		if err != nil {
			return nil, err
		}

		if len(perksBytes) > 0 {
			_ = json.Unmarshal(perksBytes, &item.PerksConfig)
		}
		if item.PerksConfig == nil {
			item.PerksConfig = []model.TierPerk{}
		}

		tiers = append(tiers, item)
	}

	return tiers, nil
}

func (r *tierRepository) FindByID(ctx context.Context, id string) (*model.TierDefinition, error) {
	query := `
		SELECT 
			t.id, t.tier_name, t.sort_order, 'POINTS_EARNED', t.min_points, t.multiplier, 
			t.perks, t.created_at, t.updated_at,
			(SELECT COUNT(*) FROM schema_ledger.members m WHERE m.tier_code = t.tier_code) as member_count
		FROM schema_backoffice.tier_definitions t
		WHERE t.id = $1 LIMIT 1`

	var item model.TierDefinition
	var perksBytes []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.TierName,
		&item.Rank,
		&item.QualificationType,
		&item.MinThreshold,
		&item.Multiplier,
		&perksBytes,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.MemberCount,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if len(perksBytes) > 0 {
		_ = json.Unmarshal(perksBytes, &item.PerksConfig)
	}
	if item.PerksConfig == nil {
		item.PerksConfig = []model.TierPerk{}
	}

	return &item, nil
}

func (r *tierRepository) Create(ctx context.Context, t *model.TierDefinition, createdBy string) error {
	perksBytes, _ := json.Marshal(t.PerksConfig)

	query := `
		INSERT INTO schema_backoffice.tier_definitions 
		(tier_code, tier_name, min_points, multiplier, perks, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id, created_at, updated_at`

	tierCode := t.TierName
	now := time.Now()

	return r.pool.QueryRow(
		ctx, query,
		tierCode, t.TierName, t.MinThreshold, t.Multiplier, perksBytes, t.Rank, now,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *tierRepository) Update(ctx context.Context, id string, t *model.TierDefinition) error {
	perksBytes, _ := json.Marshal(t.PerksConfig)

	query := `
		UPDATE schema_backoffice.tier_definitions 
		SET tier_name = $1, min_points = $2, multiplier = $3, perks = $4, sort_order = $5, updated_at = NOW()
		WHERE id = $6`

	res, err := r.pool.Exec(ctx, query, t.TierName, t.MinThreshold, t.Multiplier, perksBytes, t.Rank, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	t.ID = id
	t.UpdatedAt = time.Now()
	return nil
}

func (r *tierRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM schema_backoffice.tier_definitions WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
