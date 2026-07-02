package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopcaovanson/auth-service/internal/config"
	"github.com/shopcaovanson/auth-service/internal/domain"
	"github.com/shopcaovanson/auth-service/internal/repository"
	"github.com/shopcaovanson/auth-service/internal/storage"
)

var (
	ErrAvatarStorageUnavailable = errors.New("avatar storage is not available")
	ErrAvatarAccessDenied       = errors.New("avatar access denied")
	ErrAvatarNotFound           = errors.New("avatar not found")
)

type AvatarService struct {
	cfg     *config.Config
	users   repository.UserRepository
	storage *storage.AvatarStorage
}

func NewAvatarService(cfg *config.Config, users repository.UserRepository, store *storage.AvatarStorage) *AvatarService {
	return &AvatarService{cfg: cfg, users: users, storage: store}
}

func (s *AvatarService) Upload(ctx context.Context, actorRole string, userID uuid.UUID, file *multipart.FileHeader) (*domain.UserProfile, error) {
	if !domain.CanManageUsers(actorRole) {
		return nil, ErrForbiddenRole
	}
	if s.storage == nil || !s.storage.Enabled() {
		return nil, ErrAvatarStorageUnavailable
	}
	if file == nil {
		return nil, fmt.Errorf("avatar file is required")
	}
	if file.Size > s.cfg.MaxAvatarBytes {
		return nil, fmt.Errorf("file too large (max %dMB)", s.cfg.MaxAvatarBytes/(1024*1024))
	}

	ext, contentType, err := storage.DetectImageExt(file.Filename, file.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open upload file: %w", err)
	}
	defer src.Close()

	key, err := s.storage.Upload(ctx, userID, src, file.Size, contentType, ext)
	if err != nil {
		return nil, err
	}

	if user.AvatarKey != nil && *user.AvatarKey != "" && *user.AvatarKey != key {
		_ = s.storage.Delete(ctx, *user.AvatarKey)
	}

	updated, err := s.users.UpdateAvatarKey(ctx, userID, &key)
	if err != nil {
		_ = s.storage.Delete(ctx, key)
		return nil, err
	}

	profile := updated.ToProfile()
	return &profile, nil
}

func (s *AvatarService) Delete(ctx context.Context, actorRole string, userID uuid.UUID) (*domain.UserProfile, error) {
	if !domain.CanManageUsers(actorRole) {
		return nil, ErrForbiddenRole
	}
	if s.storage == nil || !s.storage.Enabled() {
		return nil, ErrAvatarStorageUnavailable
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.AvatarKey != nil && *user.AvatarKey != "" {
		_ = s.storage.Delete(ctx, *user.AvatarKey)
	}

	updated, err := s.users.UpdateAvatarKey(ctx, userID, nil)
	if err != nil {
		return nil, err
	}
	profile := updated.ToProfile()
	return &profile, nil
}

func (s *AvatarService) Stream(c *fiber.Ctx, viewerID uuid.UUID, viewerRole string, targetID uuid.UUID) error {
	if !canViewAvatar(viewerID, viewerRole, targetID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": ErrAvatarAccessDenied.Error()})
	}
	if s.storage == nil || !s.storage.Enabled() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": ErrAvatarStorageUnavailable.Error()})
	}

	user, err := s.users.GetByID(c.Context(), targetID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if user.AvatarKey == nil || strings.TrimSpace(*user.AvatarKey) == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": ErrAvatarNotFound.Error()})
	}

	obj, err := s.storage.Open(c.Context(), *user.AvatarKey)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": ErrAvatarNotFound.Error()})
	}
	defer obj.Reader.Close()

	c.Set("Content-Type", obj.ContentType)
	c.Set("Cache-Control", "private, max-age=300")
	c.Set("X-Content-Type-Options", "nosniff")
	return c.SendStream(obj.Reader, int(obj.Size))
}

func canViewAvatar(viewerID uuid.UUID, viewerRole string, targetID uuid.UUID) bool {
	if viewerID == targetID {
		return true
	}
	return domain.HasMinRole(viewerRole, domain.RoleSupport)
}
