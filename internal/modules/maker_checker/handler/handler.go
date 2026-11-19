package handler

import (
	"math"
	"net/http"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/middleware"
	"backoffice-service/internal/modules/maker_checker/model"
	"backoffice-service/internal/modules/maker_checker/service"
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type MakerCheckerHandler struct {
	service service.MakerCheckerService
}

func NewMakerCheckerHandler(service service.MakerCheckerService) *MakerCheckerHandler {
	return &MakerCheckerHandler{service: service}
}

func (h *MakerCheckerHandler) GetPendingActions(c *gin.Context) {
	var pagination domain.PaginationQuery
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, "Invalid pagination query")
		return
	}

	status := c.Query("status")
	items, total, err := h.service.GetPendingActions(c.Request.Context(), pagination, status)
	if err != nil {
		response.InternalError(c, "Failed to fetch pending actions: "+err.Error())
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

func (h *MakerCheckerHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	adminIDVal, _ := c.Get(middleware.CtxAdminID)
	adminID, _ := adminIDVal.(string)

	resp, err := h.service.ApproveAction(c.Request.Context(), id, adminID)
	if err != nil {
		response.BadRequest(c, "Approval failed: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, resp)
}

func (h *MakerCheckerHandler) Reject(c *gin.Context) {
	id := c.Param("id")
	var req model.RejectActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Rejection reason is required")
		return
	}

	adminIDVal, _ := c.Get(middleware.CtxAdminID)
	adminID, _ := adminIDVal.(string)

	if err := h.service.RejectAction(c.Request.Context(), id, adminID, req); err != nil {
		response.BadRequest(c, "Rejection failed: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Action rejected successfully"})
}
