package handler

import (
	"math"
	"net/http"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/voucher_catalog/model"
	"backoffice-service/internal/modules/voucher_catalog/service"
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type VoucherHandler struct {
	service service.VoucherService
}

func NewVoucherHandler(service service.VoucherService) *VoucherHandler {
	return &VoucherHandler{service: service}
}

func (h *VoucherHandler) GetBatches(c *gin.Context) {
	var pagination domain.PaginationQuery
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, "Invalid pagination parameters")
		return
	}

	items, total, err := h.service.GetBatches(c.Request.Context(), pagination)
	if err != nil {
		response.InternalError(c, "Failed to fetch voucher batches: "+err.Error())
		return
	}

	totalPages := 0
	if pagination.Limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))
	}

	pagResp := &response.PaginationBody{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		TotalItems: total,
		TotalPages: totalPages,
	}

	response.SuccessWithMeta(c, http.StatusOK, items, pagResp)
}

func (h *VoucherHandler) CreateBatch(c *gin.Context) {
	var req model.CreateVoucherBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid voucher batch payload: "+err.Error())
		return
	}

	batch, err := h.service.CreateBatch(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, "Failed to create voucher batch: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, batch)
}

func (h *VoucherHandler) UploadCodes(c *gin.Context) {
	batchID := c.Param("id")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File upload is required: "+err.Error())
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.InternalError(c, "Failed to open uploaded file: "+err.Error())
		return
	}
	defer file.Close()

	resp, err := h.service.UploadCodes(c.Request.Context(), batchID, file)
	if err != nil {
		response.InternalError(c, "Failed to process voucher codes: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, resp)
}

func (h *VoucherHandler) GetBatchInventory(c *gin.Context) {
	batchID := c.Param("id")
	var pagination domain.PaginationQuery
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, "Invalid pagination query")
		return
	}

	status := c.Query("status")
	items, total, err := h.service.GetInventory(c.Request.Context(), batchID, pagination, status)
	if err != nil {
		response.InternalError(c, "Failed to fetch voucher inventory: "+err.Error())
		return
	}

	totalPages := 0
	if pagination.Limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))
	}

	pagResp := &response.PaginationBody{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		TotalItems: total,
		TotalPages: totalPages,
	}

	response.SuccessWithMeta(c, http.StatusOK, items, pagResp)
}

func (h *VoucherHandler) RevokeVouchers(c *gin.Context) {
	batchID := c.Param("id")
	var req model.RevokeVouchersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid revoke request payload: "+err.Error())
		return
	}

	if err := h.service.RevokeVouchers(c.Request.Context(), batchID, req); err != nil {
		response.InternalError(c, "Failed to revoke vouchers: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Vouchers successfully revoked"})
}
