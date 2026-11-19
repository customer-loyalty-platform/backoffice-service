package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/earning_rule/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EarningRuleRepository interface {
	FindAll(ctx context.Context, pagination domain.PaginationQuery, status string) ([]model.DynamicRule, int64, error)
	FindByID(ctx context.Context, id string) (*model.DynamicRule, error)
	Create(ctx context.Context, r *model.DynamicRule, adminID string) error
	Update(ctx context.Context, id string, r *model.DynamicRule, adminID string) error
	UpdateStatus(ctx context.Context, id string, status string, adminID string) (*model.DynamicRule, error)
	Delete(ctx context.Context, id string) error
}

type earningRuleRepository struct {
	pool *pgxpool.Pool
}

func NewEarningRuleRepository(pool *pgxpool.Pool) EarningRuleRepository {
	return &earningRuleRepository{pool: pool}
}

func (r *earningRuleRepository) FindAll(ctx context.Context, pagination domain.PaginationQuery, status string) ([]model.DynamicRule, int64, error) {
	whereClauses := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if pagination.Search != "" {
		whereClauses += fmt.Sprintf(" AND (rule_code ILIKE $%d OR rule_name ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+pagination.Search+"%")
		argIdx++
	}

	if status != "" {
		whereClauses += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM schema_backoffice.dynamic_rules %s", whereClauses)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, rule_code, rule_name, status, priority, 
		       COALESCE(start_time::text, ''), COALESCE(end_time::text, ''), 
		       conditions_json, action_json, COALESCE(created_by::text, ''), created_at, updated_at
		FROM schema_backoffice.dynamic_rules
		%s
		ORDER BY priority ASC, created_at DESC
		LIMIT $%d OFFSET $%d`, whereClauses, argIdx, argIdx+1)

	args = append(args, pagination.Limit, pagination.Offset())

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rules []model.DynamicRule
	for rows.Next() {
		var item model.DynamicRule
		var condBytes, actBytes []byte
		err := rows.Scan(
			&item.ID,
			&item.RuleCode,
			&item.RuleName,
			&item.Status,
			&item.Priority,
			&item.StartTime,
			&item.EndTime,
			&condBytes,
			&actBytes,
			&item.CreatedBy,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if len(condBytes) > 0 {
			_ = json.Unmarshal(condBytes, &item.ConditionsJSON)
		}
		if len(actBytes) > 0 {
			_ = json.Unmarshal(actBytes, &item.ActionJSON)
		}

		rules = append(rules, item)
	}

	return rules, total, nil
}

func (r *earningRuleRepository) FindByID(ctx context.Context, id string) (*model.DynamicRule, error) {
	query := `
		SELECT id, rule_code, rule_name, status, priority, 
		       COALESCE(start_time::text, ''), COALESCE(end_time::text, ''), 
		       conditions_json, action_json, COALESCE(created_by::text, ''), created_at, updated_at
		FROM schema_backoffice.dynamic_rules
		WHERE id = $1 LIMIT 1`

	var item model.DynamicRule
	var condBytes, actBytes []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.RuleCode,
		&item.RuleName,
		&item.Status,
		&item.Priority,
		&item.StartTime,
		&item.EndTime,
		&condBytes,
		&actBytes,
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

	if len(condBytes) > 0 {
		_ = json.Unmarshal(condBytes, &item.ConditionsJSON)
	}
	if len(actBytes) > 0 {
		_ = json.Unmarshal(actBytes, &item.ActionJSON)
	}

	return &item, nil
}

func (r *earningRuleRepository) Create(ctx context.Context, rModel *model.DynamicRule, adminID string) error {
	condBytes, _ := json.Marshal(rModel.ConditionsJSON)
	actBytes, _ := json.Marshal(rModel.ActionJSON)

	query := `
		INSERT INTO schema_backoffice.dynamic_rules 
		(rule_code, rule_name, rule_type, status, priority, conditions_json, action_json, created_at, updated_at)
		VALUES ($1, $2, 'DYNAMIC_RULE', 'ACTIVE', $3, $4, $5, NOW(), NOW())
		RETURNING id, status, created_at, updated_at`

	return r.pool.QueryRow(
		ctx, query,
		rModel.RuleCode, rModel.RuleName, rModel.Priority, condBytes, actBytes,
	).Scan(&rModel.ID, &rModel.Status, &rModel.CreatedAt, &rModel.UpdatedAt)
}

func (r *earningRuleRepository) Update(ctx context.Context, id string, rModel *model.DynamicRule, adminID string) error {
	condBytes, _ := json.Marshal(rModel.ConditionsJSON)
	actBytes, _ := json.Marshal(rModel.ActionJSON)

	query := `
		UPDATE schema_backoffice.dynamic_rules 
		SET rule_name = $1, priority = $2, conditions_json = $3, action_json = $4, updated_at = NOW()
		WHERE id = $5`

	res, err := r.pool.Exec(ctx, query, rModel.RuleName, rModel.Priority, condBytes, actBytes, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	rModel.ID = id
	rModel.UpdatedAt = time.Now()
	return nil
}

func (r *earningRuleRepository) UpdateStatus(ctx context.Context, id string, status string, adminID string) (*model.DynamicRule, error) {
	query := `
		UPDATE schema_backoffice.dynamic_rules 
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

func (r *earningRuleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM schema_backoffice.dynamic_rules WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

