package handler

import (
	"math"
	"net/http"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/middleware"
	"backoffice-service/internal/modules/earning_rule/model"
	"backoffice-service/internal/modules/earning_rule/service"
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type EarningRuleHandler struct {
	service service.EarningRuleService
}

func NewEarningRuleHandler(service service.EarningRuleService) *EarningRuleHandler {
	return &EarningRuleHandler{service: service}
}

func (h *EarningRuleHandler) GetRules(c *gin.Context) {
	var pagination domain.PaginationQuery
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, "Invalid pagination parameters")
		return
	}

	status := c.Query("status")
	items, total, err := h.service.GetRules(c.Request.Context(), pagination, status)
	if err != nil {
		response.InternalError(c, "Failed to fetch rules: "+err.Error())
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

func (h *EarningRuleHandler) CreateRule(c *gin.Context) {
	var req model.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid rule payload: "+err.Error())
		return
	}

	adminIDVal, _ := c.Get(middleware.CtxAdminID)
	adminID, _ := adminIDVal.(string)

	resp, err := h.service.CreateRule(c.Request.Context(), req, adminID)
	if err != nil {
		response.BadRequest(c, "Failed to create rule: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, resp)
}

func (h *EarningRuleHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid rule update payload: "+err.Error())
		return
	}

	adminIDVal, _ := c.Get(middleware.CtxAdminID)
	adminID, _ := adminIDVal.(string)

	rule, err := h.service.UpdateRule(c.Request.Context(), id, req, adminID)
	if err != nil {
		response.InternalError(c, "Failed to update rule: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, rule)
}

func (h *EarningRuleHandler) UpdateRuleStatus(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateRuleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid status update payload: "+err.Error())
		return
	}

	adminIDVal, _ := c.Get(middleware.CtxAdminID)
	adminID, _ := adminIDVal.(string)

	rule, err := h.service.UpdateRuleStatus(c.Request.Context(), id, req.Status, adminID)
	if err != nil {
		response.InternalError(c, "Failed to update rule status: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, rule)
}

func (h *EarningRuleHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteRule(c.Request.Context(), id); err != nil {
		response.InternalError(c, "Failed to delete rule: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Rule deleted successfully"})
}
