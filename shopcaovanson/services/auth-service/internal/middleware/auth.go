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
	"github.com/google/uuid"
)

const (
	ContextUserIDKey   = "userID"
	ContextUserEmail   = "userEmail"
	ContextUserRoleKey = "userRole"
)

type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func ParseRSAPrivateKey(pemData string) (*rsa.PrivateKey, error) {
	pemData = strings.ReplaceAll(pemData, "\\n", "\n")
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	return rsaKey, nil
}

func ParseRSAPublicKey(pemData string) (*rsa.PublicKey, error) {
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

func PublicKeyFromPrivate(privateKey *rsa.PrivateKey) *rsa.PublicKey {
	return &privateKey.PublicKey
}

func JWTAuth(privateKey *rsa.PrivateKey) fiber.Handler {
	publicKey := PublicKeyFromPrivate(privateKey)

	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
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

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token subject",
			})
		}

		c.Locals(ContextUserIDKey, userID)
		c.Locals(ContextUserEmail, claims.Email)
		c.Locals(ContextUserRoleKey, claims.Role)

		return c.Next()
	}
}

func RequireMinRole(minRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals(ContextUserRoleKey).(string)
		if role == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}
		if !hasMinRole(role, minRole) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient permissions",
			})
		}
		return c.Next()
	}
}

func hasMinRole(role, minRole string) bool {
	levels := map[string]int{
		"super_admin": 100, "admin": 80, "manager": 60, "staff": 40, "support": 20, "customer": 0,
	}
	level := func(r string) int {
		if v, ok := levels[r]; ok {
			return v
		}
		return 0
	}
	return level(role) >= level(minRole)
}

func GetUserRole(c *fiber.Ctx) string {
	role, _ := c.Locals(ContextUserRoleKey).(string)
	return role
}

func GetUserID(c *fiber.Ctx) (uuid.UUID, error) {
	val := c.Locals(ContextUserIDKey)
	if val == nil {
		return uuid.Nil, errors.New("user not authenticated")
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user id in context")
	}
	return id, nil
}
