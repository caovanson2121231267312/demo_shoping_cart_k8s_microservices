package handler

import (
	"errors"
	"strconv"

	"github.com/caovanson/shopcaovanson/order-service/internal/domain"
	"github.com/caovanson/shopcaovanson/order-service/internal/middleware"
	"github.com/caovanson/shopcaovanson/order-service/internal/rbac"
	"github.com/caovanson/shopcaovanson/order-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CouponHandler struct {
	svc *service.CouponService
}

func NewCouponHandler(svc *service.CouponService) *CouponHandler {
	return &CouponHandler{svc: svc}
}

func (h *CouponHandler) RegisterRoutes(app fiber.Router) {
	app.Post("/api/coupons/validate", h.Validate)

	admin := app.Group("/api/admin", middleware.RequireAuth(), middleware.RequireMinRole(rbac.RoleManager))
	admin.Get("/coupons", h.List)
	admin.Post("/coupons", h.Create)
	admin.Get("/coupons/:id", h.Get)
	admin.Put("/coupons/:id", h.Update)
	admin.Delete("/coupons/:id", h.Delete)
}

func (h *CouponHandler) Validate(c *fiber.Ctx) error {
	var input domain.ValidateCouponInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	result, err := h.svc.Validate(c.Context(), input.Code, input.Amount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (h *CouponHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	result, err := h.svc.List(c.Context(), page, limit, c.Query("search"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (h *CouponHandler) Create(c *fiber.Ctx) error {
	var input domain.CreateCouponInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	coupon, err := h.svc.Create(c.Context(), input)
	if err != nil {
		return mapCouponError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(coupon)
}

func (h *CouponHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid coupon id"})
	}
	coupon, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return mapCouponError(c, err)
	}
	return c.JSON(coupon)
}

func (h *CouponHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid coupon id"})
	}
	var input domain.UpdateCouponInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	coupon, err := h.svc.Update(c.Context(), id, input)
	if err != nil {
		return mapCouponError(c, err)
	}
	return c.JSON(coupon)
}

func (h *CouponHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid coupon id"})
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return mapCouponError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func mapCouponError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidInput), errors.Is(err, service.ErrInvalidCoupon):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}
