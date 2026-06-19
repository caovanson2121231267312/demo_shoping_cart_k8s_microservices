package handler

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/caovanson/shopcaovanson/product-service/internal/middleware"
	"github.com/caovanson/shopcaovanson/product-service/internal/rbac"
	"github.com/caovanson/shopcaovanson/product-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ArticleHandler struct {
	articleSvc *service.ArticleService
}

func NewArticleHandler(articleSvc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{articleSvc: articleSvc}
}

func (h *ArticleHandler) RegisterRoutes(app fiber.Router) {
	api := app.Group("/api")
	api.Get("/articles", h.ListArticles)
	api.Get("/articles/:slug", h.GetArticleBySlug)

	admin := app.Group("/api/admin/articles", middleware.RequireAuth(), middleware.RequireMinRole(rbac.RoleManager))
	admin.Get("", h.AdminListArticles)
	admin.Post("", h.CreateArticle)
	admin.Put("/:id", h.UpdateArticle)
	admin.Delete("/:id", h.DeleteArticle)
}

func (h *ArticleHandler) ListArticles(c *fiber.Ctx) error {
	filter := parseArticleFilter(c)
	published := true
	filter.Published = &published
	result, err := h.articleSvc.List(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (h *ArticleHandler) AdminListArticles(c *fiber.Ctx) error {
	filter := parseArticleFilter(c)
	result, err := h.articleSvc.List(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (h *ArticleHandler) GetArticleBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if decoded, err := url.PathUnescape(slug); err == nil {
		slug = decoded
	}
	article, err := h.articleSvc.GetBySlug(c.Context(), slug, true)
	if err != nil {
		return mapArticleError(c, err)
	}
	return c.JSON(article)
}

func (h *ArticleHandler) CreateArticle(c *fiber.Ctx) error {
	var input domain.CreateArticleInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	article, err := h.articleSvc.Create(c.Context(), input)
	if err != nil {
		return mapArticleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(article)
}

func (h *ArticleHandler) UpdateArticle(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid article id"})
	}
	var input domain.UpdateArticleInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	article, err := h.articleSvc.Update(c.Context(), id, input)
	if err != nil {
		return mapArticleError(c, err)
	}
	return c.JSON(article)
}

func (h *ArticleHandler) DeleteArticle(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid article id"})
	}
	if err := h.articleSvc.Delete(c.Context(), id); err != nil {
		return mapArticleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseArticleFilter(c *fiber.Ctx) domain.ArticleListFilter {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "12"))
	filter := domain.ArticleListFilter{
		Page: page, Limit: limit,
		Category: c.Query("category"),
		Search:   strings.TrimSpace(c.Query("search")),
	}
	if c.Query("featured") == "true" {
		t := true
		filter.Featured = &t
	}
	return filter
}

func mapArticleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "article not found"})
	case errors.Is(err, service.ErrInvalidInput):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}
