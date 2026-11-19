package handler

import (
	"math"
	"net/http"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/middleware"
	"backoffice-service/internal/modules/campaign/model"
	"backoffice-service/internal/modules/campaign/service"
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type CampaignHandler struct {
	service service.CampaignService
}

func NewCampaignHandler(service service.CampaignService) *CampaignHandler {
	return &CampaignHandler{service: service}
}

func (h *CampaignHandler) GetCampaigns(c *gin.Context) {
	var pagination domain.PaginationQuery
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, "Invalid pagination query")
		return
	}

	status := c.Query("status")
	items, total, err := h.service.GetCampaigns(c.Request.Context(), pagination, status)
	if err != nil {
		response.InternalError(c, "Failed to fetch campaigns: "+err.Error())
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

func (h *CampaignHandler) CreateCampaign(c *gin.Context) {
	var req model.CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid campaign payload: "+err.Error())
		return
	}

	adminIDVal, _ := c.Get(middleware.CtxAdminID)
	adminID, _ := adminIDVal.(string)

	campaign, err := h.service.CreateCampaign(c.Request.Context(), req, adminID)
	if err != nil {
		response.InternalError(c, "Failed to create campaign: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, campaign)
}

func (h *CampaignHandler) UpdateCampaign(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid campaign update payload: "+err.Error())
		return
	}

	adminIDVal, _ := c.Get(middleware.CtxAdminID)
	adminID, _ := adminIDVal.(string)

	campaign, err := h.service.UpdateCampaign(c.Request.Context(), id, req, adminID)
	if err != nil {
		response.InternalError(c, "Failed to update campaign: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, campaign)
}

func (h *CampaignHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateCampaignStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid status update payload: "+err.Error())
		return
	}

	campaign, err := h.service.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		response.InternalError(c, "Failed to update campaign status: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, campaign)
}
