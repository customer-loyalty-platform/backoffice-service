package handler

import (
	"net/http"

	"backoffice-service/internal/modules/admin_user/model"
	"backoffice-service/internal/modules/admin_user/service"
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type AdminUserHandler struct {
	service service.AdminUserService
}

func NewAdminUserHandler(service service.AdminUserService) *AdminUserHandler {
	return &AdminUserHandler{service: service}
}

func (h *AdminUserHandler) GetAdminUsers(c *gin.Context) {
	users, total, err := h.service.GetAdminUsers(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to fetch admin users: "+err.Error())
		return
	}

	pagination := &response.PaginationBody{
		Page:       1,
		Limit:      int(total),
		TotalItems: total,
		TotalPages: 1,
	}

	response.SuccessWithMeta(c, http.StatusOK, users, pagination)
}

func (h *AdminUserHandler) CreateAdminUser(c *gin.Context) {
	var req model.CreateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid admin user payload: "+err.Error())
		return
	}

	user, err := h.service.CreateAdminUser(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, "Failed to create admin user: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, user)
}

func (h *AdminUserHandler) UpdateAdminUser(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid admin user update payload: "+err.Error())
		return
	}

	user, err := h.service.UpdateAdminUser(c.Request.Context(), id, req)
	if err != nil {
		response.InternalError(c, "Failed to update admin user: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, user)
}
