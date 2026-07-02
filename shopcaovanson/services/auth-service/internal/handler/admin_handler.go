package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopcaovanson/auth-service/internal/domain"
	"github.com/shopcaovanson/auth-service/internal/middleware"
	"github.com/shopcaovanson/auth-service/internal/repository"
	"github.com/shopcaovanson/auth-service/internal/service"
)

type AdminHandler struct {
	adminSvc  *service.AdminService
	avatarSvc *service.AvatarService
}

func NewAdminHandler(adminSvc *service.AdminService, avatarSvc *service.AvatarService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc, avatarSvc: avatarSvc}
}

func (h *AdminHandler) RegisterRoutes(router fiber.Router) {
	admin := router.Group("/api/admin", middleware.RequireMinRole(domain.RoleSupport))
	admin.Get("/stats", middleware.RequireMinRole(domain.RoleManager), h.Stats)
	admin.Get("/roles", h.ListRoles)
	admin.Get("/users", h.ListUsers)
	admin.Get("/users/:id/avatar", h.ServeUserAvatar)
	admin.Get("/users/:id", h.GetUser)
	admin.Put("/users/:id/role", middleware.RequireMinRole(domain.RoleAdmin), h.UpdateUserRole)
	admin.Put("/users/:id/status", middleware.RequireMinRole(domain.RoleAdmin), h.UpdateUserStatus)
	admin.Put("/users/:id", middleware.RequireMinRole(domain.RoleAdmin), h.UpdateUser)
	admin.Post("/users/:id/avatar", middleware.RequireMinRole(domain.RoleAdmin), h.UploadUserAvatar)
	admin.Delete("/users/:id/avatar", middleware.RequireMinRole(domain.RoleAdmin), h.DeleteUserAvatar)
}

func (h *AdminHandler) Stats(c *fiber.Ctx) error {
	stats, err := h.adminSvc.Stats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(stats)
}

func (h *AdminHandler) ListRoles(c *fiber.Ctx) error {
	return c.JSON(h.adminSvc.ListRoles())
}

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	from, to := parseCreatedRange(c)
	filter := domain.UserListFilter{
		Page:        page,
		Limit:       limit,
		Search:      c.Query("search"),
		Role:        c.Query("role"),
		CreatedFrom: from,
		CreatedTo:   to,
	}
	result, err := h.adminSvc.ListUsers(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (h *AdminHandler) GetUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	user, err := h.adminSvc.GetUser(c.Context(), id)
	if err != nil {
		return mapAdminError(c, err)
	}
	return c.JSON(user)
}

func (h *AdminHandler) UpdateUserRole(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	role, _ := c.Locals(middleware.ContextUserRoleKey).(string)
	var input domain.UpdateUserRoleInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	user, err := h.adminSvc.UpdateUserRole(c.Context(), role, id, input.Role)
	if err != nil {
		return mapAdminError(c, err)
	}
	return c.JSON(user)
}

func (h *AdminHandler) UpdateUserStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	role, _ := c.Locals(middleware.ContextUserRoleKey).(string)
	var input domain.UpdateUserStatusInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	user, err := h.adminSvc.UpdateUserStatus(c.Context(), role, id, input.IsActive)
	if err != nil {
		return mapAdminError(c, err)
	}
	return c.JSON(user)
}

func (h *AdminHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	role, _ := c.Locals(middleware.ContextUserRoleKey).(string)
	var input domain.UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	user, err := h.adminSvc.UpdateUser(c.Context(), role, id, input)
	if err != nil {
		if err.Error() == "full_name is required" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return mapAdminError(c, err)
	}
	return c.JSON(user)
}

func (h *AdminHandler) ServeUserAvatar(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	viewerID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	viewerRole := middleware.GetUserRole(c)
	return h.avatarSvc.Stream(c, viewerID, viewerRole, id)
}

func (h *AdminHandler) UploadUserAvatar(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	role, _ := c.Locals(middleware.ContextUserRoleKey).(string)
	file, err := c.FormFile("avatar")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "avatar file is required"})
	}
	user, err := h.avatarSvc.Upload(c.Context(), role, id, file)
	if err != nil {
		return mapAvatarError(c, err)
	}
	return c.JSON(user)
}

func (h *AdminHandler) DeleteUserAvatar(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	role, _ := c.Locals(middleware.ContextUserRoleKey).(string)
	user, err := h.avatarSvc.Delete(c.Context(), role, id)
	if err != nil {
		return mapAvatarError(c, err)
	}
	return c.JSON(user)
}

func mapAdminError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	case errors.Is(err, service.ErrForbiddenRole):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidRole):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}
