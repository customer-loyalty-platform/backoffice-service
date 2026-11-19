package service

import (
	"context"

	"backoffice-service/internal/modules/settlement/model"
	"backoffice-service/internal/modules/settlement/repository"
)

type SettlementService interface {
	GetPartners(ctx context.Context) ([]model.PartnerMerchant, int64, error)
	CreatePartner(ctx context.Context, req model.CreatePartnerRequest) (*model.PartnerMerchant, error)
	UpdatePartner(ctx context.Context, id string, req model.UpdatePartnerRequest) (*model.PartnerMerchant, error)
	GetSettlementReport(ctx context.Context, params model.SettlementReportParams) (*model.SettlementReport, error)
}

type settlementService struct {
	repo repository.SettlementRepository
}

func NewSettlementService(repo repository.SettlementRepository) SettlementService {
	return &settlementService{repo: repo}
}

func (s *settlementService) GetPartners(ctx context.Context) ([]model.PartnerMerchant, int64, error) {
	return s.repo.FindAll(ctx)
}

func (s *settlementService) CreatePartner(ctx context.Context, req model.CreatePartnerRequest) (*model.PartnerMerchant, error) {
	p := &model.PartnerMerchant{
		MerchantCode:       req.MerchantCode,
		MerchantName:       req.MerchantName,
		ContactEmail:       req.ContactEmail,
		ContactPhone:       req.ContactPhone,
		ParValuePerPoint:   req.ParValuePerPoint,
		CommissionRate:     req.CommissionRate,
		SettlementCurrency: req.SettlementCurrency,
		ContractStartDate:  req.ContractStartDate,
		ContractEndDate:    req.ContractEndDate,
		Status:             "ACTIVE",
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *settlementService) UpdatePartner(ctx context.Context, id string, req model.UpdatePartnerRequest) (*model.PartnerMerchant, error) {
	p := &model.PartnerMerchant{
		MerchantCode:       req.MerchantCode,
		MerchantName:       req.MerchantName,
		ContactEmail:       req.ContactEmail,
		ContactPhone:       req.ContactPhone,
		ParValuePerPoint:   req.ParValuePerPoint,
		CommissionRate:     req.CommissionRate,
		SettlementCurrency: req.SettlementCurrency,
		ContractStartDate:  req.ContractStartDate,
		ContractEndDate:    req.ContractEndDate,
	}

	if err := s.repo.Update(ctx, id, p); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *settlementService) GetSettlementReport(ctx context.Context, params model.SettlementReportParams) (*model.SettlementReport, error) {
	return s.repo.GetSettlementReport(ctx, params.PartnerID, params.PeriodStart, params.PeriodEnd)
}

