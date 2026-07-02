package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopcaovanson/auth-service/internal/middleware"
	"github.com/shopcaovanson/auth-service/internal/service"
)

type AvatarHandler struct {
	avatarSvc *service.AvatarService
}

func NewAvatarHandler(avatarSvc *service.AvatarService) *AvatarHandler {
	return &AvatarHandler{avatarSvc: avatarSvc}
}

func (h *AvatarHandler) Serve(c *fiber.Ctx) error {
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	viewerID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	viewerRole := middleware.GetUserRole(c)
	return h.avatarSvc.Stream(c, viewerID, viewerRole, targetID)
}

func mapAvatarError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrForbiddenRole):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrAvatarStorageUnavailable):
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": err.Error()})
	default:
		return mapAdminError(c, err)
	}
}
