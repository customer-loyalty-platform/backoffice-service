package service

import (
	"context"
	"time"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/member/model"
	"backoffice-service/internal/modules/member/repository"
)

type MemberService interface {
	GetMembers(ctx context.Context, pagination domain.PaginationQuery, statusFilter, tierFilter string) ([]model.MemberListItem, int64, error)
	GetMemberByID(ctx context.Context, id string) (*model.MemberDetail, error)
	UpdateStatus(ctx context.Context, id string, req model.UpdateMemberStatusRequest, adminID string) (*model.UpdateMemberStatusResponse, error)
	SubmitAdjustment(ctx context.Context, id string, req model.PointAdjustmentRequest, makerID string) (*model.PointAdjustmentResponse, error)
	GetMemberTransactions(ctx context.Context, memberID string, pagination domain.PaginationQuery) ([]model.PointLedgerEntry, int64, error)
}

type memberService struct {
	repo repository.MemberRepository
}

func NewMemberService(repo repository.MemberRepository) MemberService {
	return &memberService{repo: repo}
}

func (s *memberService) GetMembers(ctx context.Context, pagination domain.PaginationQuery, statusFilter, tierFilter string) ([]model.MemberListItem, int64, error) {
	pagination.Normalize()
	return s.repo.FindAll(ctx, pagination, statusFilter, tierFilter)
}

func (s *memberService) GetMemberByID(ctx context.Context, id string) (*model.MemberDetail, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *memberService) UpdateStatus(ctx context.Context, id string, req model.UpdateMemberStatusRequest, adminID string) (*model.UpdateMemberStatusResponse, error) {
	prevStatus, err := s.repo.UpdateStatus(ctx, id, req.Status)
	if err != nil {
		return nil, err
	}

	return &model.UpdateMemberStatusResponse{
		MemberID:       id,
		PreviousStatus: prevStatus,
		NewStatus:      req.Status,
		UpdatedBy:      adminID,
		UpdatedAt:      time.Now(),
	}, nil
}

func (s *memberService) SubmitAdjustment(ctx context.Context, id string, req model.PointAdjustmentRequest, makerID string) (*model.PointAdjustmentResponse, error) {
	// Maker-checker workflow
	requiresChecker := req.Amount > 1000 || req.Amount < -1000

	actionID, err := s.repo.CreatePendingAction(ctx, "POINT_ADJUSTMENT", "members", id, makerID, req)
	if err != nil {
		return nil, err
	}

	return &model.PointAdjustmentResponse{
		Status:          "PENDING_APPROVAL",
		ActionID:        actionID,
		RequiresChecker: requiresChecker,
		MakerID:         makerID,
	}, nil
}

func (s *memberService) GetMemberTransactions(ctx context.Context, memberID string, pagination domain.PaginationQuery) ([]model.PointLedgerEntry, int64, error) {
	pagination.Normalize()
	return s.repo.FindTransactionsByMemberID(ctx, memberID, pagination)
}

