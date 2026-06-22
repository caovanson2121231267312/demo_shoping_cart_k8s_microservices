package handler

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/caovanson/shopcaovanson/product-service/internal/middleware"
	"github.com/caovanson/shopcaovanson/product-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProductHandler struct {
	productSvc  *service.ProductService
	categorySvc *service.CategoryService
	reviewSvc   *service.ReviewService
}

func NewProductHandler(productSvc *service.ProductService, categorySvc *service.CategoryService, reviewSvc *service.ReviewService) *ProductHandler {
	return &ProductHandler{
		productSvc:  productSvc,
		categorySvc: categorySvc,
		reviewSvc:   reviewSvc,
	}
}

func (h *ProductHandler) RegisterRoutes(app fiber.Router) {
	api := app.Group("/api")
	api.Get("/categories", h.ListCategories)
	api.Get("/products", h.ListProducts)
	api.Post("/products", middleware.RequireAuth(), middleware.RequireAdmin(), h.CreateProduct)
	api.Post("/products/reviews/summary", h.BatchReviewSummaries)
	api.Get("/products/:id/reviews", h.ListReviews)
	api.Post("/products/:id/reviews", middleware.RequireAuth(), h.CreateReview)
	api.Put("/products/:id", middleware.RequireAuth(), middleware.RequireAdmin(), h.UpdateProduct)
	api.Delete("/products/:id", middleware.RequireAuth(), middleware.RequireAdmin(), h.DeleteProduct)
	api.Get("/products/:slug", h.GetProductBySlug)

	app.Get("/internal/products/:id", h.GetStockInfo)
}

func (h *ProductHandler) ListCategories(c *fiber.Ctx) error {
	tree, err := h.categorySvc.GetTree(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tree)
}

func (h *ProductHandler) ListProducts(c *fiber.Ctx) error {
	filter := parseProductFilter(c)
	result, err := h.productSvc.List(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (h *ProductHandler) GetProductBySlug(c *fiber.Ctx) error {
	param := c.Params("slug")
	if param == "" {
		param = c.Params("id")
	}
	if param == "" {
		uri := string(c.Request().URI().Path())
		const prefix = "/api/products/"
		if strings.HasPrefix(uri, prefix) {
			param = strings.TrimPrefix(uri, prefix)
		}
	}
	if param == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "product not found"})
	}
	if decoded, err := url.PathUnescape(param); err == nil && decoded != "" {
		param = decoded
	}
	if isUUID(param) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "product not found"})
	}
	product, err := h.productSvc.GetBySlug(c.Context(), param)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "product not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(product)
}

func (h *ProductHandler) GetStockInfo(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid product id"})
	}
	info, err := h.productSvc.GetStockInfo(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "product not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(info)
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var input domain.CreateProductInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	product, err := h.productSvc.Create(c.Context(), input)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(product)
}

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid product id"})
	}
	var input domain.UpdateProductInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	product, err := h.productSvc.Update(c.Context(), id, input)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(product)
}

func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid product id"})
	}
	if err := h.productSvc.Delete(c.Context(), id); err != nil {
		return mapServiceError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ProductHandler) BatchReviewSummaries(c *fiber.Ctx) error {
	var body struct {
		ProductIDs []string `json:"product_ids"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	ids := make([]uuid.UUID, 0, len(body.ProductIDs))
	for _, raw := range body.ProductIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	result, err := h.reviewSvc.SummarizeByProducts(c.Context(), ids)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"data": result})
}

func (h *ProductHandler) ListReviews(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid product id"})
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	result, err := h.reviewSvc.ListByProduct(c.Context(), id, page, limit)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(result)
}

func (h *ProductHandler) CreateReview(c *fiber.Ctx) error {
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid product id"})
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authentication required"})
	}
	var input domain.CreateReviewInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	review, err := h.reviewSvc.Create(c.Context(), productID, userID, input)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(review)
}

func parseProductFilter(c *fiber.Ctx) domain.ProductListFilter {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	from, to := parseCreatedRange(c)
	filter := domain.ProductListFilter{
		Page:            page,
		Limit:           limit,
		Category:        c.Query("category"),
		Search:          c.Query("search"),
		Sort:            c.Query("sort"),
		IncludeInactive: c.Query("include_inactive") == "true",
		CreatedFrom:     from,
		CreatedTo:       to,
	}
	if v := c.Query("min_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MinPrice = &f
		}
	}
	if v := c.Query("max_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MaxPrice = &f
		}
	}
	return filter
}

func mapServiceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidInput):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrDuplicateReview):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}

func isUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
