package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopcaovanson/auth-service/internal/domain"
	"github.com/shopcaovanson/auth-service/internal/service"
)

func (h *AdminHandler) ListLoginHistory(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	from, to := parseCreatedRange(c)
	filter := domain.LoginHistoryFilter{
		Limit:       limit,
		Cursor:      c.Query("cursor"),
		Email:       c.Query("email"),
		UserID:      c.Query("user_id"),
		CreatedFrom: from,
		CreatedTo:   to,
	}
	if v := c.Query("success"); v != "" {
		b := v == "true" || v == "1"
		filter.Success = &b
	}
	result, err := h.loginSvc.List(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (h *AdminHandler) ListLoginReports(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	from, to := parseReportDateRange(c)
	result, err := h.loginSvc.ListReports(c.Context(), page, limit, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (h *AdminHandler) DownloadLoginReport(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid report id"})
	}
	obj, report, err := h.loginSvc.OpenReportFile(c.Context(), id)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": msg})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}
	defer obj.Reader.Close()

	c.Set("Content-Type", obj.ContentType)
	c.Set("Content-Disposition", `attachment; filename="`+report.FileName+`"`)
	c.Set("Content-Length", strconv.FormatInt(obj.Size, 10))
	return c.SendStream(obj.Reader, int(obj.Size))
}

func (h *AdminHandler) GenerateLoginReport(c *fiber.Ctx) error {
	var body struct {
		Date string `json:"date"`
	}
	_ = c.BodyParser(&body)
	day := service.YesterdayVN(time.Now())
	if body.Date != "" {
		parsed, err := time.ParseInLocation("2006-01-02", body.Date, service.VietnamLocation())
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid date, use YYYY-MM-DD"})
		}
		day = parsed
	}
	report, err := h.loginSvc.GenerateDailyReport(c.Context(), day)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(report)
}

func parseReportDateRange(c *fiber.Ctx) (from, to *time.Time) {
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = &t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = &t
		}
	}
	return from, to
}
