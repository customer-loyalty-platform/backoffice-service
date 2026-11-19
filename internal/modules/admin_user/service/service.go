package service

import (
	"context"

	"backoffice-service/internal/modules/admin_user/model"
	"backoffice-service/internal/modules/admin_user/repository"

	"golang.org/x/crypto/bcrypt"
)

type AdminUserService interface {
	GetAdminUsers(ctx context.Context) ([]model.AdminUserItem, int64, error)
	CreateAdminUser(ctx context.Context, req model.CreateAdminUserRequest) (*model.AdminUserItem, error)
	UpdateAdminUser(ctx context.Context, id string, req model.UpdateAdminUserRequest) (*model.AdminUserItem, error)
}

type adminUserService struct {
	repo repository.AdminUserRepository
}

func NewAdminUserService(repo repository.AdminUserRepository) AdminUserService {
	return &adminUserService{repo: repo}
}

func (s *adminUserService) GetAdminUsers(ctx context.Context) ([]model.AdminUserItem, int64, error) {
	return s.repo.FindAll(ctx)
}

func (s *adminUserService) CreateAdminUser(ctx context.Context, req model.CreateAdminUserRequest) (*model.AdminUserItem, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &model.AdminUserItem{
		Username: req.Username,
		Email:    req.Email,
		Role:     req.Role,
		Status:   "ACTIVE",
	}

	if err := s.repo.Create(ctx, u, string(hashedPassword)); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *adminUserService) UpdateAdminUser(ctx context.Context, id string, req model.UpdateAdminUserRequest) (*model.AdminUserItem, error) {
	return s.repo.Update(ctx, id, req)
}

