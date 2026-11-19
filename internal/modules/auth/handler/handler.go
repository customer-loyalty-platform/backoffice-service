package handler

import (
	"net/http"

	"backoffice-service/internal/middleware"
	"backoffice-service/internal/modules/auth/model"
	"backoffice-service/internal/modules/auth/service"
	"backoffice-service/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid login credentials format: "+err.Error())
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid refresh token payload")
		return
	}

	resp, err := h.service.RefreshToken(c.Request.Context(), req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	response.Success(c, http.StatusOK, model.LogoutResponse{
		Message: "Successfully logged out",
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	adminIDVal, exists := c.Get(middleware.CtxAdminID)
	if !exists {
		response.Unauthorized(c, "Admin ID not found in session")
		return
	}

	adminID := adminIDVal.(string)
	resp, err := h.service.GetMe(c.Request.Context(), adminID)
	if err != nil {
		response.NotFound(c, "Admin profile not found: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, resp)
}
