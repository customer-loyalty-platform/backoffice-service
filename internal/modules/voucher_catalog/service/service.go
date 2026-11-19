package service

import (
	"bufio"
	"context"
	"io"
	"strings"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/voucher_catalog/model"
	"backoffice-service/internal/modules/voucher_catalog/repository"
)

type VoucherService interface {
	GetBatches(ctx context.Context, pagination domain.PaginationQuery) ([]model.VoucherBatch, int64, error)
	CreateBatch(ctx context.Context, req model.CreateVoucherBatchRequest) (*model.VoucherBatch, error)
	UploadCodes(ctx context.Context, batchID string, file io.Reader) (*model.UploadCodesResponse, error)
	GetInventory(ctx context.Context, batchID string, pagination domain.PaginationQuery, status string) ([]model.VoucherInventoryItem, int64, error)
	RevokeVouchers(ctx context.Context, batchID string, req model.RevokeVouchersRequest) error
}

type voucherService struct {
	repo repository.VoucherRepository
}

func NewVoucherService(repo repository.VoucherRepository) VoucherService {
	return &voucherService{repo: repo}
}

func (s *voucherService) GetBatches(ctx context.Context, pagination domain.PaginationQuery) ([]model.VoucherBatch, int64, error) {
	pagination.Normalize()
	return s.repo.FindBatches(ctx, pagination)
}

func (s *voucherService) CreateBatch(ctx context.Context, req model.CreateVoucherBatchRequest) (*model.VoucherBatch, error) {
	b := &model.VoucherBatch{
		BatchCode:       req.BatchCode,
		Title:           req.Title,
		Description:     req.Description,
		PartnerBrand:    req.PartnerBrand,
		MonetaryValue:   req.MonetaryValue,
		PointsCost:      req.PointsCost,
		TermsConditions: req.TermsConditions,
		Status:          "ACTIVE",
	}

	if err := s.repo.CreateBatch(ctx, b); err != nil {
		return nil, err
	}

	return b, nil
}

func (s *voucherService) UploadCodes(ctx context.Context, batchID string, file io.Reader) (*model.UploadCodesResponse, error) {
	var codes []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") && !strings.EqualFold(line, "voucher_code") {
			// Extract first token if CSV separated
			parts := strings.Split(line, ",")
			code := strings.TrimSpace(parts[0])
			if code != "" {
				codes = append(codes, code)
			}
		}
	}

	imported, skipped, err := s.repo.ImportCodes(ctx, batchID, codes)
	if err != nil {
		return nil, err
	}

	batch, err := s.repo.FindBatchByID(ctx, batchID)
	var totalStock, availableStock int64
	if err == nil {
		totalStock = batch.TotalStock
		availableStock = batch.AvailableStock
	}

	return &model.UploadCodesResponse{
		BatchID:           batchID,
		CodesImported:     imported,
		DuplicatesSkipped: skipped,
		NewTotalStock:     totalStock,
		NewAvailableStock: availableStock,
	}, nil
}

func (s *voucherService) GetInventory(ctx context.Context, batchID string, pagination domain.PaginationQuery, status string) ([]model.VoucherInventoryItem, int64, error) {
	pagination.Normalize()
	return s.repo.FindInventoryByBatch(ctx, batchID, pagination, status)
}

func (s *voucherService) RevokeVouchers(ctx context.Context, batchID string, req model.RevokeVouchersRequest) error {
	return s.repo.RevokeVouchers(ctx, batchID, req.VoucherIDs, req.Reason)
}

