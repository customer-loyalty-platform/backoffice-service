package service

import (
	"context"
	"time"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/maker_checker/model"
	"backoffice-service/internal/modules/maker_checker/repository"
)

type MakerCheckerService interface {
	GetPendingActions(ctx context.Context, pagination domain.PaginationQuery, status string) ([]model.PendingAction, int64, error)
	ApproveAction(ctx context.Context, id, checkerID string) (*model.ApproveActionResponse, error)
	RejectAction(ctx context.Context, id, checkerID string, req model.RejectActionRequest) error
}

type makerCheckerService struct {
	repo repository.MakerCheckerRepository
}

func NewMakerCheckerService(repo repository.MakerCheckerRepository) MakerCheckerService {
	return &makerCheckerService{repo: repo}
}

func (s *makerCheckerService) GetPendingActions(ctx context.Context, pagination domain.PaginationQuery, status string) ([]model.PendingAction, int64, error) {
	pagination.Normalize()
	return s.repo.FindPending(ctx, pagination, status)
}

func (s *makerCheckerService) ApproveAction(ctx context.Context, id, checkerID string) (*model.ApproveActionResponse, error) {
	if err := s.repo.Approve(ctx, id, checkerID); err != nil {
		return nil, err
	}

	return &model.ApproveActionResponse{
		Status:     "APPROVED",
		ActionID:   id,
		ExecutedAt: time.Now().Format(time.RFC3339),
		Message:    "Action approved and executed successfully",
	}, nil
}

func (s *makerCheckerService) RejectAction(ctx context.Context, id, checkerID string, req model.RejectActionRequest) error {
	return s.repo.Reject(ctx, id, checkerID, req.RejectionReason)
}

