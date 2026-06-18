package handler

import (
	"errors"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/caovanson/shopcaovanson/product-service/internal/middleware"
	"github.com/caovanson/shopcaovanson/product-service/internal/rbac"
	"github.com/caovanson/shopcaovanson/product-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	categorySvc *service.CategoryService
}

func NewCategoryHandler(categorySvc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categorySvc: categorySvc}
}

func (h *CategoryHandler) RegisterAdminRoutes(app fiber.Router) {
	admin := app.Group("/api/admin/categories", middleware.RequireAuth(), middleware.RequireMinRole(rbac.RoleManager))
	admin.Post("", h.CreateCategory)
	admin.Put("/:id", h.UpdateCategory)
	admin.Delete("/:id", h.DeleteCategory)
}

func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	var input domain.CreateCategoryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	cat, err := h.categorySvc.Create(c.Context(), input)
	if err != nil {
		return mapCategoryError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(cat)
}

func (h *CategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid category id"})
	}
	var input domain.UpdateCategoryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	cat, err := h.categorySvc.Update(c.Context(), id, input)
	if err != nil {
		return mapCategoryError(c, err)
	}
	return c.JSON(cat)
}

func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid category id"})
	}
	if err := h.categorySvc.Delete(c.Context(), id); err != nil {
		return mapCategoryError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func mapCategoryError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidInput):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}
