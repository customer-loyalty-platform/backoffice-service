package handler

import (
	"math"
	"net/http"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/modules/monitoring/service"
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type MonitoringHandler struct {
	service service.MonitoringService
}

func NewMonitoringHandler(service service.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{service: service}
}

func (h *MonitoringHandler) GetSummary(c *gin.Context) {
	summary, err := h.service.GetDashboardSummary(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to fetch dashboard summary: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, summary)
}

func (h *MonitoringHandler) GetTransactionFeed(c *gin.Context) {
	var pagination domain.PaginationQuery
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, "Invalid pagination query")
		return
	}

	entries, total, err := h.service.GetTransactionFeed(c.Request.Context(), pagination)
	if err != nil {
		response.InternalError(c, "Failed to fetch transaction feed: "+err.Error())
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

	response.SuccessWithMeta(c, http.StatusOK, entries, pagResp)
}

func (h *MonitoringHandler) GetEarningTrend(c *gin.Context) {
	points, err := h.service.GetEarningTrend(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to fetch earning trend: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, points)
}

func (h *MonitoringHandler) GetAuditLogs(c *gin.Context) {
	var pagination domain.PaginationQuery
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, "Invalid pagination query")
		return
	}

	adminID := c.Query("admin_id")
	action := c.Query("action")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	logs, total, err := h.service.GetAuditLogs(c.Request.Context(), pagination, adminID, action, startDate, endDate)
	if err != nil {
		response.InternalError(c, "Failed to fetch audit logs: "+err.Error())
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

	response.SuccessWithMeta(c, http.StatusOK, logs, pagResp)
}
