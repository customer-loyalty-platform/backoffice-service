package handler

import (
	"net/http"

	"backoffice-service/internal/modules/settlement/model"
	"backoffice-service/internal/modules/settlement/service"
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type SettlementHandler struct {
	service service.SettlementService
}

func NewSettlementHandler(service service.SettlementService) *SettlementHandler {
	return &SettlementHandler{service: service}
}

func (h *SettlementHandler) GetPartners(c *gin.Context) {
	partners, total, err := h.service.GetPartners(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to fetch partners: "+err.Error())
		return
	}

	pagination := &response.PaginationBody{
		Page:       1,
		Limit:      int(total),
		TotalItems: total,
		TotalPages: 1,
	}

	response.SuccessWithMeta(c, http.StatusOK, partners, pagination)
}

func (h *SettlementHandler) CreatePartner(c *gin.Context) {
	var req model.CreatePartnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid partner payload: "+err.Error())
		return
	}

	partner, err := h.service.CreatePartner(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, "Failed to create partner: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, partner)
}

func (h *SettlementHandler) UpdatePartner(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdatePartnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid partner update payload: "+err.Error())
		return
	}

	partner, err := h.service.UpdatePartner(c.Request.Context(), id, req)
	if err != nil {
		response.InternalError(c, "Failed to update partner: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, partner)
}

func (h *SettlementHandler) GetSettlementReport(c *gin.Context) {
	var params model.SettlementReportParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "partner_id is required for settlement report")
		return
	}

	report, err := h.service.GetSettlementReport(c.Request.Context(), params)
	if err != nil {
		response.InternalError(c, "Failed to generate settlement report: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, report)
}
