package handler

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/shopcaovanson/auth-service/internal/repository"
)

type InternalHandler struct {
	userRepo repository.UserRepository
}

func NewInternalHandler(userRepo repository.UserRepository) *InternalHandler {
	return &InternalHandler{userRepo: userRepo}
}

func (h *InternalHandler) GetUserByEmail(c *fiber.Ctx) error {
	email := strings.TrimSpace(strings.ToLower(c.Query("email")))
	if email == "" || !strings.Contains(email, "@") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
	}
	user, err := h.userRepo.GetByEmail(c.Context(), email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(fiber.Map{
		"id":    user.ID.String(),
		"email": user.Email,
	})
}
