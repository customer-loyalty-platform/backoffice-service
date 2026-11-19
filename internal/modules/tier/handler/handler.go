package handler

import (
	"net/http"

	"backoffice-service/internal/middleware"
	"backoffice-service/internal/modules/tier/model"
	"backoffice-service/internal/modules/tier/service"
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type TierHandler struct {
	service service.TierService
}

func NewTierHandler(service service.TierService) *TierHandler {
	return &TierHandler{service: service}
}

func (h *TierHandler) GetTiers(c *gin.Context) {
	tiers, err := h.service.GetTiers(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to fetch tiers: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, tiers)
}

func (h *TierHandler) CreateTier(c *gin.Context) {
	var req model.CreateTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid tier payload: "+err.Error())
		return
	}

	adminIDVal, _ := c.Get(middleware.CtxAdminID)
	adminID, _ := adminIDVal.(string)

	tier, err := h.service.CreateTier(c.Request.Context(), req, adminID)
	if err != nil {
		response.InternalError(c, "Failed to create tier: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, tier)
}

func (h *TierHandler) UpdateTier(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid tier update payload: "+err.Error())
		return
	}

	tier, err := h.service.UpdateTier(c.Request.Context(), id, req)
	if err != nil {
		response.InternalError(c, "Failed to update tier: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, tier)
}

func (h *TierHandler) DeleteTier(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteTier(c.Request.Context(), id); err != nil {
		response.InternalError(c, "Failed to delete tier: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Tier deleted successfully"})
}
