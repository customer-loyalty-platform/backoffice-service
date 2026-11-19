package service

import (
	"context"

	"backoffice-service/internal/domain"
	memberModel "backoffice-service/internal/modules/member/model"
	"backoffice-service/internal/modules/monitoring/model"
	"backoffice-service/internal/modules/monitoring/repository"
)

type MonitoringService interface {
	GetDashboardSummary(ctx context.Context) (*model.DashboardSummary, error)
	GetTransactionFeed(ctx context.Context, pagination domain.PaginationQuery) ([]memberModel.PointLedgerEntry, int64, error)
	GetEarningTrend(ctx context.Context) ([]model.ChartDataPoint, error)
	GetAuditLogs(ctx context.Context, pagination domain.PaginationQuery, adminID, action, startDate, endDate string) ([]model.AuditLogEntry, int64, error)
}

type monitoringService struct {
	repo repository.MonitoringRepository
}

func NewMonitoringService(repo repository.MonitoringRepository) MonitoringService {
	return &monitoringService{repo: repo}
}

func (s *monitoringService) GetDashboardSummary(ctx context.Context) (*model.DashboardSummary, error) {
	return s.repo.GetDashboardSummary(ctx)
}

func (s *monitoringService) GetTransactionFeed(ctx context.Context, pagination domain.PaginationQuery) ([]memberModel.PointLedgerEntry, int64, error) {
	pagination.Normalize()
	return s.repo.GetTransactionFeed(ctx, pagination)
}

func (s *monitoringService) GetEarningTrend(ctx context.Context) ([]model.ChartDataPoint, error) {
	return s.repo.GetEarningTrend(ctx)
}

func (s *monitoringService) GetAuditLogs(ctx context.Context, pagination domain.PaginationQuery, adminID, action, startDate, endDate string) ([]model.AuditLogEntry, int64, error) {
	pagination.Normalize()
	return s.repo.GetAuditLogs(ctx, pagination, adminID, action, startDate, endDate)
}

