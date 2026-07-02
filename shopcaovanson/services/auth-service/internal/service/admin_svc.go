package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/shopcaovanson/auth-service/internal/domain"
	"github.com/shopcaovanson/auth-service/internal/repository"
)

var (
	ErrForbiddenRole = errors.New("insufficient permissions")
	ErrInvalidRole   = errors.New("invalid role")
)

type AdminService struct {
	users repository.UserRepository
}

func NewAdminService(users repository.UserRepository) *AdminService {
	return &AdminService{users: users}
}

func (s *AdminService) ListUsers(ctx context.Context, filter domain.UserListFilter) (*domain.UserListResult, error) {
	return s.users.List(ctx, filter)
}

func (s *AdminService) GetUser(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, repository.ErrUserNotFound
	}
	profile := user.ToProfile()
	return &profile, nil
}

func (s *AdminService) UpdateUserRole(ctx context.Context, actorRole string, userID uuid.UUID, role string) (*domain.UserProfile, error) {
	if !domain.CanManageUsers(actorRole) {
		return nil, ErrForbiddenRole
	}
	if !isValidAssignableRole(role) {
		return nil, ErrInvalidRole
	}
	if !domain.CanAssignRole(actorRole, role) {
		return nil, ErrForbiddenRole
	}
	user, err := s.users.UpdateRole(ctx, userID, role)
	if err != nil {
		return nil, err
	}
	profile := user.ToProfile()
	return &profile, nil
}

func (s *AdminService) UpdateUserStatus(ctx context.Context, actorRole string, userID uuid.UUID, isActive bool) (*domain.UserProfile, error) {
	if !domain.CanManageUsers(actorRole) {
		return nil, ErrForbiddenRole
	}
	user, err := s.users.UpdateStatus(ctx, userID, isActive)
	if err != nil {
		return nil, err
	}
	profile := user.ToProfile()
	return &profile, nil
}

func (s *AdminService) UpdateUser(ctx context.Context, actorRole string, userID uuid.UUID, input domain.UpdateUserInput) (*domain.UserProfile, error) {
	if !domain.CanManageUsers(actorRole) {
		return nil, ErrForbiddenRole
	}
	fullName := strings.TrimSpace(input.FullName)
	if fullName == "" {
		return nil, fmt.Errorf("full_name is required")
	}
	user, err := s.users.UpdateUser(ctx, userID, fullName)
	if err != nil {
		return nil, err
	}
	profile := user.ToProfile()
	return &profile, nil
}

func (s *AdminService) Stats(ctx context.Context) (*domain.AdminStats, error) {
	return s.users.Stats(ctx)
}

func (s *AdminService) ListRoles() []map[string]interface{} {
	roles := []struct {
		ID          string
		Label       string
		Level       int
		Description string
	}{
		{domain.RoleSuperAdmin, "Siêu quản trị", 100, "Toàn quyền hệ thống"},
		{domain.RoleAdmin, "Quản trị", 80, "Quản lý người dùng, sản phẩm, đơn hàng"},
		{domain.RoleManager, "Quản lý", 60, "Quản lý sản phẩm, danh mục, xem thống kê"},
		{domain.RoleStaff, "Nhân viên", 40, "Xử lý đơn hàng"},
		{domain.RoleSupport, "Hỗ trợ", 20, "Tra cứu đơn hàng, xem người dùng"},
		{domain.RoleCustomer, "Khách hàng", 0, "Mua sắm"},
	}
	out := make([]map[string]interface{}, len(roles))
	for i, r := range roles {
		out[i] = map[string]interface{}{
			"id": r.ID, "label": r.Label, "level": r.Level, "description": r.Description,
		}
	}
	return out
}

func isValidAssignableRole(role string) bool {
	switch role {
	case domain.RoleSuperAdmin, domain.RoleAdmin, domain.RoleManager, domain.RoleStaff, domain.RoleSupport, domain.RoleCustomer:
		return true
	default:
		return false
	}
}

func validateActorCanViewUser(actorRole string) error {
	if domain.HasMinRole(actorRole, domain.RoleSupport) {
		return nil
	}
	return fmt.Errorf("%w", ErrForbiddenRole)
}
