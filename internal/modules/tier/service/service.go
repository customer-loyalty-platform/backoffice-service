package service

import (
	"context"

	"backoffice-service/internal/modules/tier/model"
	"backoffice-service/internal/modules/tier/repository"
)

type TierService interface {
	GetTiers(ctx context.Context) ([]model.TierDefinition, error)
	CreateTier(ctx context.Context, req model.CreateTierRequest, adminID string) (*model.TierDefinition, error)
	UpdateTier(ctx context.Context, id string, req model.UpdateTierRequest) (*model.TierDefinition, error)
	DeleteTier(ctx context.Context, id string) error
}

type tierService struct {
	repo repository.TierRepository
}

func NewTierService(repo repository.TierRepository) TierService {
	return &tierService{repo: repo}
}

func (s *tierService) GetTiers(ctx context.Context) ([]model.TierDefinition, error) {
	return s.repo.FindAll(ctx)
}

func (s *tierService) CreateTier(ctx context.Context, req model.CreateTierRequest, adminID string) (*model.TierDefinition, error) {
	tier := &model.TierDefinition{
		TierName:          req.TierName,
		Rank:              req.Rank,
		QualificationType: req.QualificationType,
		MinThreshold:      req.MinThreshold,
		Multiplier:        req.Multiplier,
		PerksConfig:       req.PerksConfig,
	}

	if err := s.repo.Create(ctx, tier, adminID); err != nil {
		return nil, err
	}

	return tier, nil
}

func (s *tierService) UpdateTier(ctx context.Context, id string, req model.UpdateTierRequest) (*model.TierDefinition, error) {
	tier := &model.TierDefinition{
		TierName:          req.TierName,
		Rank:              req.Rank,
		QualificationType: req.QualificationType,
		MinThreshold:      req.MinThreshold,
		Multiplier:        req.Multiplier,
		PerksConfig:       req.PerksConfig,
	}

	if err := s.repo.Update(ctx, id, tier); err != nil {
		return nil, err
	}

	return tier, nil
}

func (s *tierService) DeleteTier(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

