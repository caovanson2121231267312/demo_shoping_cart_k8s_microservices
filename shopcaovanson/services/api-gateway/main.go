package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"
	"github.com/shopcaovanson/api-gateway/internal/config"
	"github.com/shopcaovanson/api-gateway/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("parse redis url: %v", err)
	}
	redisClient := redis.NewClient(opt)
	defer redisClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("ping redis: %v", err)
	}

	publicKey, err := middleware.ParseRSAPublicKey(cfg.JWTPublicKeyPEM)
	if err != nil {
		log.Fatalf("parse jwt public key: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "api-gateway",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		ProxyHeader:  fiber.HeaderXForwardedFor,
	})

	app.Use(recover.New())
	app.Use(middleware.RequestLogger())
	app.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	app.Use(middleware.RateLimit(redisClient, cfg.RateLimitPerMinute))
	app.Use(middleware.JWTValidation(publicKey))

	app.Get("/health", func(c *fiber.Ctx) error {
		if err := redisClient.Ping(c.Context()).Err(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  "redis unavailable",
			})
		}
		return c.JSON(fiber.Map{"status": "ok", "service": "api-gateway"})
	})

	registerProxyRoutes(app, cfg)

	go func() {
		addr := ":" + cfg.Port
		log.Printf("api-gateway listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down api-gateway...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = app.ShutdownWithContext(shutdownCtx)
}

func registerProxyRoutes(app *fiber.App, cfg *config.Config) {
	authURL := strings.TrimRight(cfg.AuthServiceURL, "/")
	productURL := strings.TrimRight(cfg.ProductServiceURL, "/")
	orderURL := strings.TrimRight(cfg.OrderServiceURL, "/")
	chatURL := strings.TrimRight(cfg.ChatServiceURL, "/")
	rasaURL := strings.TrimRight(cfg.RasaServiceURL, "/")
	notificationURL := strings.TrimRight(cfg.NotificationServiceURL, "/")
	analyticsURL := strings.TrimRight(cfg.AnalyticsServiceURL, "/")

	app.All("/api/auth", middleware.PublicOrAuth(), proxyHandler(authURL))
	app.All("/api/auth/*", middleware.PublicOrAuth(), proxyHandler(authURL))

	app.All("/api/admin/users", middleware.PublicOrAuth(), proxyHandler(authURL))
	app.All("/api/admin/users/*", middleware.PublicOrAuth(), proxyHandler(authURL))
	app.All("/api/admin/roles", middleware.PublicOrAuth(), proxyHandler(authURL))
	app.All("/api/admin/stats", middleware.PublicOrAuth(), proxyHandler(authURL))

	app.All("/api/admin/categories", middleware.PublicOrAuth(), proxyHandler(productURL))
	app.All("/api/admin/categories/*", middleware.PublicOrAuth(), proxyHandler(productURL))
	app.All("/api/admin/articles", middleware.PublicOrAuth(), proxyHandler(productURL))
	app.All("/api/admin/articles/*", middleware.PublicOrAuth(), proxyHandler(productURL))

	app.All("/api/articles", middleware.PublicOrAuth(), proxyHandler(productURL))
	app.All("/api/articles/*", middleware.PublicOrAuth(), proxyHandler(productURL))

	app.All("/api/products", middleware.PublicOrAuth(), proxyHandler(productURL))
	app.All("/api/products/*", middleware.PublicOrAuth(), proxyHandler(productURL))
	app.All("/api/categories", middleware.PublicOrAuth(), proxyHandler(productURL))
	app.All("/api/categories/*", middleware.PublicOrAuth(), proxyHandler(productURL))

	app.All("/api/cart", middleware.PublicOrAuth(), proxyHandler(orderURL))
	app.All("/api/cart/*", middleware.PublicOrAuth(), proxyHandler(orderURL))
	app.All("/api/orders", middleware.PublicOrAuth(), proxyHandler(orderURL))
	app.All("/api/orders/*", middleware.PublicOrAuth(), proxyHandler(orderURL))
    app.All("/api/admin/orders/export", middleware.PublicOrAuth(), proxyHandler(analyticsURL))
    app.All("/api/admin/orders/export/*", middleware.PublicOrAuth(), proxyHandler(analyticsURL))
    app.All("/api/admin/orders", middleware.PublicOrAuth(), proxyHandler(orderURL))
	app.All("/api/admin/orders/*", middleware.PublicOrAuth(), proxyHandler(orderURL))
	app.All("/api/admin/coupons", middleware.PublicOrAuth(), proxyHandler(orderURL))
	app.All("/api/admin/coupons/*", middleware.PublicOrAuth(), proxyHandler(orderURL))
	app.All("/api/coupons", middleware.PublicOrAuth(), proxyHandler(orderURL))
	app.All("/api/coupons/*", middleware.PublicOrAuth(), proxyHandler(orderURL))

	app.All("/api/invoices", middleware.PublicOrAuth(), proxyHandler(notificationURL))
	app.All("/api/invoices/*", middleware.PublicOrAuth(), proxyHandler(notificationURL))

	app.All("/api/chat", middleware.PublicOrAuth(), proxyHandler(chatURL))
	app.All("/api/chat/*", middleware.PublicOrAuth(), proxyHandler(chatURL))
	app.All("/ws", middleware.PublicOrAuth(), proxyHandler(chatURL))

	app.All("/api/chatbot", middleware.PublicOrAuth(), proxyHandler(rasaURL))
	app.All("/api/chatbot/*", middleware.PublicOrAuth(), proxyHandler(rasaURL))

	app.All("/api/analytics/presence", middleware.PublicOrAuth(), proxyHandler(analyticsURL))
	app.All("/api/analytics/track", middleware.PublicOrAuth(), proxyHandler(analyticsURL))
	app.All("/api/admin/analytics", middleware.PublicOrAuth(), proxyHandler(analyticsURL))
	app.All("/api/admin/analytics/*", middleware.PublicOrAuth(), proxyHandler(analyticsURL))
}

func proxyHandler(baseURL string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		dest := baseURL + c.OriginalURL()
		return proxy.Do(c, dest)
	}
}
