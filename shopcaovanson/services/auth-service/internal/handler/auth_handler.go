package handler

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/shopcaovanson/auth-service/internal/middleware"
	"github.com/shopcaovanson/auth-service/internal/repository"
	"github.com/shopcaovanson/auth-service/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	tokens, err := h.svc.Register(c.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		return mapAuthError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(tokens)
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Token == "" {
		req.Token = c.Query("token")
	}

	tokens, err := h.svc.VerifyEmail(c.Context(), req.Token)
	if err != nil {
		return mapAuthError(c, err)
	}
	return c.JSON(tokens)
}

func (h *AuthHandler) ResendVerification(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.svc.ResendVerification(c.Context(), req.Email); err != nil {
		return mapAuthError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Nếu email tồn tại, link xác nhận đã được gửi."})
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	resp, err := h.svc.ForgotPassword(c.Context(), req.Email)
	if err != nil {
		return mapAuthError(c, err)
	}
	return c.JSON(resp)
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req struct {
		Email       string `json:"email"`
		OTP         string `json:"otp"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	resp, err := h.svc.ResetPassword(c.Context(), req.Email, req.OTP, req.NewPassword)
	if err != nil {
		return mapAuthError(c, err)
	}
	return c.JSON(resp)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	tokens, err := h.svc.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return mapAuthError(c, err)
	}

	return c.JSON(tokens)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	accessToken, err := h.svc.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return mapAuthError(c, err)
	}

	return c.JSON(fiber.Map{"access_token": accessToken})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.BodyParser(&req)

	if err := h.svc.Logout(c.Context(), userID, req.RefreshToken); err != nil {
		return mapAuthError(c, err)
	}

	return c.JSON(fiber.Map{"message": "logged out"})
}

func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	profile, err := h.svc.GetProfile(c.Context(), userID)
	if err != nil {
		return mapAuthError(c, err)
	}

	return c.JSON(profile)
}

func (h *AuthHandler) UpdateMe(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req struct {
		FullName string `json:"full_name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	profile, err := h.svc.UpdateProfile(c.Context(), userID, req.FullName)
	if err != nil {
		return mapAuthError(c, err)
	}

	return c.JSON(profile)
}

func mapAuthError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrUserExists):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already registered"})
	case errors.Is(err, service.ErrInvalidCredentials):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid email or password"})
	case errors.Is(err, service.ErrInvalidRefresh):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid refresh token"})
	case errors.Is(err, service.ErrWeakPassword):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidEmail):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrEmailNotVerified):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "email not verified", "code": "EMAIL_NOT_VERIFIED"})
	case errors.Is(err, service.ErrAccountDisabled):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "account is disabled", "code": "ACCOUNT_DISABLED"})
	case errors.Is(err, service.ErrInvalidVerifyToken):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid or expired verification token"})
	case errors.Is(err, service.ErrInvalidResetOTP):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid or expired otp", "code": "INVALID_RESET_OTP"})
	case errors.Is(err, repository.ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	default:
		if err.Error() == "full_name is required" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		log.Printf("unmapped auth error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
