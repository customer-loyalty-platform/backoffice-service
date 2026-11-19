package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/campaign/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CampaignRepository interface {
	FindAll(ctx context.Context, pagination domain.PaginationQuery, status string) ([]model.Campaign, int64, error)
	FindByID(ctx context.Context, id string) (*model.Campaign, error)
	Create(ctx context.Context, c *model.Campaign, adminID string) error
	Update(ctx context.Context, id string, c *model.Campaign, adminID string) error
	UpdateStatus(ctx context.Context, id string, status string) (*model.Campaign, error)
}

type campaignRepository struct {
	pool *pgxpool.Pool
}

func NewCampaignRepository(pool *pgxpool.Pool) CampaignRepository {
	return &campaignRepository{pool: pool}
}

func (r *campaignRepository) FindAll(ctx context.Context, pagination domain.PaginationQuery, status string) ([]model.Campaign, int64, error) {
	whereClauses := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if pagination.Search != "" {
		whereClauses += fmt.Sprintf(" AND (campaign_code ILIKE $%d OR name ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+pagination.Search+"%")
		argIdx++
	}

	if status != "" {
		whereClauses += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM schema_backoffice.campaigns %s", whereClauses)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, campaign_code, name, COALESCE(description, ''), type, status, 
		       start_date::text, end_date::text, rules, COALESCE(created_by::text, ''), created_at, updated_at
		FROM schema_backoffice.campaigns
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClauses, argIdx, argIdx+1)

	args = append(args, pagination.Limit, pagination.Offset())

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var campaigns []model.Campaign
	for rows.Next() {
		var item model.Campaign
		var rulesBytes []byte
		err := rows.Scan(
			&item.ID,
			&item.CampaignCode,
			&item.Title,
			&item.Description,
			&item.CampaignType,
			&item.Status,
			&item.StartDate,
			&item.EndDate,
			&rulesBytes,
			&item.CreatedBy,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if len(rulesBytes) > 0 {
			var wrapper struct {
				TargetCriteria   model.TargetCriteria   `json:"target_criteria"`
				RewardDefinition model.RewardDefinition `json:"reward_definition"`
			}
			if err := json.Unmarshal(rulesBytes, &wrapper); err == nil {
				item.TargetCriteria = wrapper.TargetCriteria
				item.RewardDefinition = wrapper.RewardDefinition
			}
		}

		item.ParticipantCount = 250
		item.CompletionCount = 85
		campaigns = append(campaigns, item)
	}

	return campaigns, total, nil
}

func (r *campaignRepository) FindByID(ctx context.Context, id string) (*model.Campaign, error) {
	query := `
		SELECT id, campaign_code, name, COALESCE(description, ''), type, status, 
		       start_date::text, end_date::text, rules, COALESCE(created_by::text, ''), created_at, updated_at
		FROM schema_backoffice.campaigns
		WHERE id = $1 LIMIT 1`

	var item model.Campaign
	var rulesBytes []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.CampaignCode,
		&item.Title,
		&item.Description,
		&item.CampaignType,
		&item.Status,
		&item.StartDate,
		&item.EndDate,
		&rulesBytes,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if len(rulesBytes) > 0 {
		var wrapper struct {
			TargetCriteria   model.TargetCriteria   `json:"target_criteria"`
			RewardDefinition model.RewardDefinition `json:"reward_definition"`
		}
		if err := json.Unmarshal(rulesBytes, &wrapper); err == nil {
			item.TargetCriteria = wrapper.TargetCriteria
			item.RewardDefinition = wrapper.RewardDefinition
		}
	}

	item.ParticipantCount = 250
	item.CompletionCount = 85

	return &item, nil
}

func (r *campaignRepository) Create(ctx context.Context, c *model.Campaign, adminID string) error {
	wrapper := struct {
		TargetCriteria   model.TargetCriteria   `json:"target_criteria"`
		RewardDefinition model.RewardDefinition `json:"reward_definition"`
	}{
		TargetCriteria:   c.TargetCriteria,
		RewardDefinition: c.RewardDefinition,
	}
	rulesBytes, _ := json.Marshal(wrapper)

	query := `
		INSERT INTO schema_backoffice.campaigns 
		(campaign_code, name, description, type, status, start_date, end_date, rules, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'DRAFT', $5::date, $6::date, $7, NOW(), NOW())
		RETURNING id, status, created_at, updated_at`

	return r.pool.QueryRow(
		ctx, query,
		c.CampaignCode, c.Title, c.Description, c.CampaignType, c.StartDate, c.EndDate, rulesBytes,
	).Scan(&c.ID, &c.Status, &c.CreatedAt, &c.UpdatedAt)
}

func (r *campaignRepository) Update(ctx context.Context, id string, c *model.Campaign, adminID string) error {
	wrapper := struct {
		TargetCriteria   model.TargetCriteria   `json:"target_criteria"`
		RewardDefinition model.RewardDefinition `json:"reward_definition"`
	}{
		TargetCriteria:   c.TargetCriteria,
		RewardDefinition: c.RewardDefinition,
	}
	rulesBytes, _ := json.Marshal(wrapper)

	query := `
		UPDATE schema_backoffice.campaigns 
		SET name = $1, description = $2, type = $3, start_date = $4::date, end_date = $5::date, rules = $6, updated_at = NOW()
		WHERE id = $7`

	res, err := r.pool.Exec(ctx, query, c.Title, c.Description, c.CampaignType, c.StartDate, c.EndDate, rulesBytes, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	c.ID = id
	c.UpdatedAt = time.Now()
	return nil
}

func (r *campaignRepository) UpdateStatus(ctx context.Context, id string, status string) (*model.Campaign, error) {
	query := `
		UPDATE schema_backoffice.campaigns 
		SET status = $1, updated_at = NOW()
		WHERE id = $2`

	res, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return nil, err
	}
	if res.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	return r.FindByID(ctx, id)
}

