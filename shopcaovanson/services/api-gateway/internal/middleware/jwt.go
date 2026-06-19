package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ContextUserIDKey   = "userID"
	ContextUserRoleKey = "userRole"
	HeaderUserID       = "X-User-Id"
	HeaderUserRole     = "X-User-Role"
	HeaderUserEmail    = "X-User-Email"
)

type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func ParseRSAPublicKey(pemData string) (*rsa.PublicKey, error) {
	pemData = strings.ReplaceAll(pemData, "\\n", "\n")
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return rsaPub, nil
}

func JWTValidation(publicKey *rsa.PublicKey) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return publicKey, nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		if claims.Subject == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token subject",
			})
		}

		c.Locals(ContextUserIDKey, claims.Subject)
		c.Locals(ContextUserRoleKey, claims.Role)

		c.Request().Header.Set(HeaderUserID, claims.Subject)
		c.Request().Header.Set(HeaderUserRole, claims.Role)
		if claims.Email != "" {
			c.Request().Header.Set(HeaderUserEmail, claims.Email)
		}

		return c.Next()
	}
}

func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals(ContextUserIDKey)
		if userID == nil || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}
		return c.Next()
	}
}

func IsPublicRoute(method, path string) bool {
	method = strings.ToUpper(method)

	if path == "/health" {
		return true
	}

	switch {
	case path == "/api/auth/register",
		path == "/api/auth/login",
		path == "/api/auth/refresh",
		path == "/api/auth/verify-email",
		path == "/api/auth/resend-verification",
		path == "/api/auth/forgot-password",
		path == "/api/auth/reset-password":
		return true
	case path == "/api/chatbot/message" && method == "POST":
		return true
	case path == "/api/orders/track" && method == "POST":
		return true
	case path == "/api/coupons/validate" && method == "POST":
		return true
	case strings.HasPrefix(path, "/api/products") && method == "GET":
		return true
	case path == "/api/products/reviews/summary" && method == "POST":
		return true
	case strings.HasPrefix(path, "/api/articles") && method == "GET":
		return true
	case path == "/api/categories" || strings.HasPrefix(path, "/api/categories/"):
		return method == "GET"
	default:
		return false
	}
}

func PublicOrAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if IsPublicRoute(c.Method(), c.Path()) {
			return c.Next()
		}
		return RequireAuth()(c)
	}
}
