package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const UserIDKey = "user_id"
const UserRoleKey = "user_role"
const UserEmailKey = "user_email"

type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func JWTMiddleware(publicKeyPEM string) fiber.Handler {
	pubKey, err := parseRSAPublicKey(publicKeyPEM)
	if err != nil {
		return func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "jwt public key not configured",
			})
		}
	}

	return func(c *fiber.Ctx) error {
		if userID := c.Get("X-User-Id"); userID != "" {
			c.Locals(UserIDKey, userID)
			c.Locals(UserRoleKey, c.Get("X-User-Role"))
			c.Locals(UserEmailKey, c.Get("X-User-Email"))
			return c.Next()
		}

		tokenStr := extractBearerToken(c.Get("Authorization"))
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization"})
		}

		claims, err := validateToken(tokenStr, pubKey)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		c.Locals(UserIDKey, claims.Subject)
		c.Locals(UserRoleKey, claims.Role)
		c.Locals(UserEmailKey, claims.Email)
		return c.Next()
	}
}

func ValidateTokenString(tokenStr, publicKeyPEM string) (*Claims, error) {
	pubKey, err := parseRSAPublicKey(publicKeyPEM)
	if err != nil {
		return nil, err
	}
	return validateToken(tokenStr, pubKey)
}

func GetUserID(c *fiber.Ctx) string {
	if v, ok := c.Locals(UserIDKey).(string); ok {
		return v
	}
	return ""
}

func validateToken(tokenStr string, pubKey *rsa.PublicKey) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func parseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	if pemStr == "" {
		return nil, errors.New("empty public key")
	}
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return rsaPub, nil
}

func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
