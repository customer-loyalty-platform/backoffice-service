package handler

import (
	"math"
	"net/http"

	"backoffice-service/internal/domain"
	"backoffice-service/internal/middleware"
	"backoffice-service/internal/modules/member/model"
	"backoffice-service/internal/modules/member/service"
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	service service.MemberService
}

func NewMemberHandler(service service.MemberService) *MemberHandler {
	return &MemberHandler{service: service}
}

func (h *MemberHandler) GetMembers(c *gin.Context) {
	var pagination domain.PaginationQuery
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, "Invalid pagination parameters")
		return
	}

	statusFilter := c.Query("status")
	tierFilter := c.Query("tier_id")

	items, total, err := h.service.GetMembers(c.Request.Context(), pagination, statusFilter, tierFilter)
	if err != nil {
		response.InternalError(c, "Failed to retrieve members: "+err.Error())
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

func (h *MemberHandler) GetMemberByID(c *gin.Context) {
	id := c.Param("id")
	detail, err := h.service.GetMemberByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Member not found")
		return
	}

	response.Success(c, http.StatusOK, detail)
}

func (h *MemberHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateMemberStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid status update payload: "+err.Error())
		return
	}

	adminIDVal, _ := c.Get(middleware.CtxAdminID)
	adminID, _ := adminIDVal.(string)

	resp, err := h.service.UpdateStatus(c.Request.Context(), id, req, adminID)
	if err != nil {
		response.InternalError(c, "Failed to update member status: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, resp)
}

func (h *MemberHandler) SubmitAdjustment(c *gin.Context) {
	id := c.Param("id")
	var req model.PointAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid point adjustment payload: "+err.Error())
		return
	}

	adminIDVal, _ := c.Get(middleware.CtxAdminID)
	adminID, _ := adminIDVal.(string)

	resp, err := h.service.SubmitAdjustment(c.Request.Context(), id, req, adminID)
	if err != nil {
		response.InternalError(c, "Failed to submit point adjustment: "+err.Error())
		return
	}

	response.Success(c, http.StatusAccepted, resp)
}

func (h *MemberHandler) GetMemberTransactions(c *gin.Context) {
	id := c.Param("id")
	var pagination domain.PaginationQuery
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, "Invalid pagination query")
		return
	}

	entries, total, err := h.service.GetMemberTransactions(c.Request.Context(), id, pagination)
	if err != nil {
		response.InternalError(c, "Failed to retrieve transactions: "+err.Error())
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
