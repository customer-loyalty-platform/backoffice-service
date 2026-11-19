package service

import (
	"context"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/campaign/model"
	"backoffice-service/internal/modules/campaign/repository"
)

type CampaignService interface {
	GetCampaigns(ctx context.Context, pagination domain.PaginationQuery, status string) ([]model.Campaign, int64, error)
	GetCampaignByID(ctx context.Context, id string) (*model.Campaign, error)
	CreateCampaign(ctx context.Context, req model.CreateCampaignRequest, adminID string) (*model.Campaign, error)
	UpdateCampaign(ctx context.Context, id string, req model.UpdateCampaignRequest, adminID string) (*model.Campaign, error)
	UpdateStatus(ctx context.Context, id string, status string) (*model.Campaign, error)
}

type campaignService struct {
	repo repository.CampaignRepository
}

func NewCampaignService(repo repository.CampaignRepository) CampaignService {
	return &campaignService{repo: repo}
}

func (s *campaignService) GetCampaigns(ctx context.Context, pagination domain.PaginationQuery, status string) ([]model.Campaign, int64, error) {
	pagination.Normalize()
	return s.repo.FindAll(ctx, pagination, status)
}

func (s *campaignService) GetCampaignByID(ctx context.Context, id string) (*model.Campaign, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *campaignService) CreateCampaign(ctx context.Context, req model.CreateCampaignRequest, adminID string) (*model.Campaign, error) {
	c := &model.Campaign{
		CampaignCode:     req.CampaignCode,
		Title:            req.Title,
		Description:      req.Description,
		BannerURL:        req.BannerURL,
		CampaignType:     req.CampaignType,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		TargetCriteria:   req.TargetCriteria,
		RewardDefinition: req.RewardDefinition,
	}

	if err := s.repo.Create(ctx, c, adminID); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *campaignService) UpdateCampaign(ctx context.Context, id string, req model.UpdateCampaignRequest, adminID string) (*model.Campaign, error) {
	c := &model.Campaign{
		CampaignCode:     req.CampaignCode,
		Title:            req.Title,
		Description:      req.Description,
		BannerURL:        req.BannerURL,
		CampaignType:     req.CampaignType,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		TargetCriteria:   req.TargetCriteria,
		RewardDefinition: req.RewardDefinition,
	}

	if err := s.repo.Update(ctx, id, c, adminID); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *campaignService) UpdateStatus(ctx context.Context, id string, status string) (*model.Campaign, error) {
	return s.repo.UpdateStatus(ctx, id, status)
}

