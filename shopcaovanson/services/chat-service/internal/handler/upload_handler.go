package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/caovanson/shopcaovanson/chat-service/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UploadHandler struct {
	cfg *config.Config
}

func NewUploadHandler(cfg *config.Config) *UploadHandler {
	return &UploadHandler{cfg: cfg}
}

func (h *UploadHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is required"})
	}
	if file.Size > h.cfg.MaxUploadBytes {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("file too large (max %dMB)", h.cfg.MaxUploadBytes/(1024*1024))})
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "only jpg, png, gif, webp allowed"})
	}

	if err := os.MkdirAll(h.cfg.UploadDir, 0o755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot create upload dir"})
	}

	name := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), uuid.NewString()[:8], ext)
	dest := filepath.Join(h.cfg.UploadDir, name)
	if err := c.SaveFile(file, dest); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "save failed"})
	}

	url := "/api/chat/media/" + name
	return c.JSON(fiber.Map{"data": fiber.Map{"url": url}})
}

func (h *UploadHandler) ServeMedia(c *fiber.Ctx) error {
	name := filepath.Base(c.Params("filename"))
	if name == "" || strings.Contains(name, "..") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid file"})
	}
	path := filepath.Join(h.cfg.UploadDir, name)
	if _, err := os.Stat(path); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.SendFile(path)
}
