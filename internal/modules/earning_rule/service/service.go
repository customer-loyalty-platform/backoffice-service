package service

import (
	"context"
	"encoding/json"
	"errors"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/earning_rule/model"
	"backoffice-service/internal/modules/earning_rule/repository"
	"backoffice-service/internal/pkg/validator"
)

type EarningRuleService interface {
	GetRules(ctx context.Context, pagination domain.PaginationQuery, status string) ([]model.DynamicRule, int64, error)
	GetRuleByID(ctx context.Context, id string) (*model.DynamicRule, error)
	CreateRule(ctx context.Context, req model.CreateRuleRequest, adminID string) (*model.CreateRuleResponse, error)
	UpdateRule(ctx context.Context, id string, req model.UpdateRuleRequest, adminID string) (*model.DynamicRule, error)
	UpdateRuleStatus(ctx context.Context, id string, status string, adminID string) (*model.DynamicRule, error)
	DeleteRule(ctx context.Context, id string) error
}

type earningRuleService struct {
	repo      repository.EarningRuleRepository
	validator *validator.JSONSchemaValidator
}

func NewEarningRuleService(repo repository.EarningRuleRepository, v *validator.JSONSchemaValidator) EarningRuleService {
	return &earningRuleService{
		repo:      repo,
		validator: v,
	}
}

func (s *earningRuleService) GetRules(ctx context.Context, pagination domain.PaginationQuery, status string) ([]model.DynamicRule, int64, error) {
	pagination.Normalize()
	return s.repo.FindAll(ctx, pagination, status)
}

func (s *earningRuleService) GetRuleByID(ctx context.Context, id string) (*model.DynamicRule, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *earningRuleService) CreateRule(ctx context.Context, req model.CreateRuleRequest, adminID string) (*model.CreateRuleResponse, error) {
	// Validate JSON schema if and-conditions exist
	if len(req.ConditionsJSON.And) > 0 {
		condBytes, err := json.Marshal(req.ConditionsJSON)
		if err == nil && s.validator != nil {
			if err := s.validator.ValidateJSONB(condBytes); err != nil {
				return nil, errors.New("conditions schema validation failed: " + err.Error())
			}
		}
	}

	r := &model.DynamicRule{
		RuleCode:       req.RuleCode,
		RuleName:       req.RuleName,
		Priority:       req.Priority,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		ConditionsJSON: req.ConditionsJSON,
		ActionJSON:     req.ActionJSON,
	}

	if err := s.repo.Create(ctx, r, adminID); err != nil {
		return nil, err
	}

	return &model.CreateRuleResponse{
		RuleID:   r.ID,
		RuleCode: r.RuleCode,
		State:    r.Status,
	}, nil
}

func (s *earningRuleService) UpdateRule(ctx context.Context, id string, req model.UpdateRuleRequest, adminID string) (*model.DynamicRule, error) {
	r := &model.DynamicRule{
		RuleName:       req.RuleName,
		Priority:       req.Priority,
		ConditionsJSON: req.ConditionsJSON,
		ActionJSON:     req.ActionJSON,
	}

	if err := s.repo.Update(ctx, id, r, adminID); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *earningRuleService) UpdateRuleStatus(ctx context.Context, id string, status string, adminID string) (*model.DynamicRule, error) {
	return s.repo.UpdateStatus(ctx, id, status, adminID)
}

func (s *earningRuleService) DeleteRule(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

