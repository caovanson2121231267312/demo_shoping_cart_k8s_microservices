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

type OrderHandler struct {
	cartSvc  *service.CartService
	orderSvc *service.OrderService
}

func NewOrderHandler(cartSvc *service.CartService, orderSvc *service.OrderService) *OrderHandler {
	return &OrderHandler{cartSvc: cartSvc, orderSvc: orderSvc}
}

func (h *OrderHandler) RegisterRoutes(app fiber.Router) {
	app.Post("/api/orders/track", h.TrackOrder)
	app.Post("/api/orders/lookup", h.LookupOrders)

	api := app.Group("/api", middleware.RequireAuth())

	api.Get("/cart", h.GetCart)
	api.Post("/cart/items", h.AddCartItem)
	api.Put("/cart/items/:productId", h.UpdateCartItem)
	api.Delete("/cart/items/:productId", h.RemoveCartItem)
	api.Delete("/cart", h.ClearCart)

	api.Post("/orders", h.Checkout)
	api.Get("/orders", h.ListMyOrders)
	api.Get("/orders/:id", h.GetOrder)
	api.Put("/orders/:id/cancel", h.CancelOrder)

	admin := app.Group("/api/admin", middleware.RequireAuth(), middleware.RequireSupport())
	admin.Get("/orders", h.ListAllOrders)
	admin.Get("/orders/stats", middleware.RequireMinRole(rbac.RoleManager), h.OrderStats)
	admin.Get("/orders/search", h.SearchOrders)
	admin.Get("/orders/:id", h.GetOrderAdmin)
	admin.Put("/orders/:id/status", middleware.RequireStaff(), h.UpdateOrderStatus)
}

func (h *OrderHandler) GetCart(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	cart, err := h.cartSvc.GetCart(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(cart)
}

func (h *OrderHandler) AddCartItem(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	var input domain.AddCartItemInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	cart, err := h.cartSvc.AddItem(c.Context(), userID, input)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(cart)
}

func (h *OrderHandler) UpdateCartItem(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	productID, err := uuid.Parse(c.Params("productId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid product id"})
	}
	var input domain.UpdateCartItemInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	cart, err := h.cartSvc.UpdateItem(c.Context(), userID, productID, input)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(cart)
}

func (h *OrderHandler) RemoveCartItem(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	productID, err := uuid.Parse(c.Params("productId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid product id"})
	}
	cart, err := h.cartSvc.RemoveItem(c.Context(), userID, productID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(cart)
}

func (h *OrderHandler) ClearCart(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	if err := h.cartSvc.ClearCart(c.Context(), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *OrderHandler) Checkout(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	var input domain.CheckoutInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	order, err := h.orderSvc.Checkout(c.Context(), userID, c.Get("X-User-Email"), input)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(order)
}

func (h *OrderHandler) ListMyOrders(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	result, err := h.orderSvc.ListMyOrders(c.Context(), userID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order id"})
	}
	order, err := h.orderSvc.GetByID(c.Context(), userID, role, orderID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(order)
}

func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order id"})
	}
	order, err := h.orderSvc.CancelOrder(c.Context(), userID, orderID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(order)
}

func (h *OrderHandler) ListAllOrders(c *fiber.Ctx) error {
	from, to := parseCreatedRange(c)
	filter := domain.OrderSearchFilter{
		Page:        queryInt(c, "page", 1),
		Limit:       queryInt(c, "limit", 20),
		Status:      c.Query("status"),
		Search:      c.Query("search"),
		CreatedFrom: from,
		CreatedTo:   to,
	}
	result, err := h.orderSvc.SearchOrders(c.Context(), filter)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(result)
}

func (h *OrderHandler) SearchOrders(c *fiber.Ctx) error {
	from, to := parseCreatedRange(c)
	filter := domain.OrderSearchFilter{
		Page:          queryInt(c, "page", 1),
		Limit:         queryInt(c, "limit", 20),
		Status:        c.Query("status"),
		Search:        c.Query("search"),
		OrderNumber:   c.Query("order_number"),
		ShippingPhone: c.Query("shipping_phone"),
		UserID:        c.Query("user_id"),
		CreatedFrom:   from,
		CreatedTo:     to,
	}
	result, err := h.orderSvc.SearchOrders(c.Context(), filter)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(result)
}

func (h *OrderHandler) OrderStats(c *fiber.Ctx) error {
	stats, err := h.orderSvc.OrderStats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(stats)
}

func (h *OrderHandler) GetOrderAdmin(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order id"})
	}
	order, err := h.orderSvc.GetByID(c.Context(), userID, role, orderID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(order)
}

func (h *OrderHandler) TrackOrder(c *fiber.Ctx) error {
	var input domain.TrackOrderInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	order, err := h.orderSvc.TrackOrder(c.Context(), input)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(order)
}

func (h *OrderHandler) LookupOrders(c *fiber.Ctx) error {
	var input domain.LookupOrdersInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	result, err := h.orderSvc.LookupOrders(c.Context(), input)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(result)
}

func queryInt(c *fiber.Ctx, key string, def int) int {
	v, err := strconv.Atoi(c.Query(key, strconv.Itoa(def)))
	if err != nil {
		return def
	}
	return v
}

func (h *OrderHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order id"})
	}
	var input domain.UpdateOrderStatusInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	order, err := h.orderSvc.UpdateOrderStatus(c.Context(), orderID, input.Status)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(order)
}

func mapServiceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidInput):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrEmptyCart):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidStatus):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidCoupon):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}
