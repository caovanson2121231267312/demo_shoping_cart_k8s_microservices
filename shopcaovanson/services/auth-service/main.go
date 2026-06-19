package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/shopcaovanson/auth-service/internal/config"
	"github.com/shopcaovanson/auth-service/internal/handler"
	"github.com/shopcaovanson/auth-service/internal/kafka"
	"github.com/shopcaovanson/auth-service/internal/middleware"
	"github.com/shopcaovanson/auth-service/internal/repository"
	"github.com/shopcaovanson/auth-service/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

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

	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("create kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	privateKey, err := middleware.ParseRSAPrivateKey(cfg.JWTPrivateKeyPEM)
	if err != nil {
		log.Fatalf("parse jwt private key: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)
	authSvc := service.NewAuthService(cfg, userRepo, refreshRepo, redisClient, kafkaProducer, privateKey)
	authHandler := handler.NewAuthHandler(authSvc)
	adminSvc := service.NewAdminService(userRepo)
	adminHandler := handler.NewAdminHandler(adminSvc)

	app := fiber.New(fiber.Config{
		AppName:      "auth-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Use(recover.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		if err := db.PingContext(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  "database unavailable",
			})
		}
		if err := redisClient.Ping(c.Context()).Err(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  "redis unavailable",
			})
		}
		return c.JSON(fiber.Map{"status": "ok", "service": "auth-service"})
	})

	api := app.Group("/api/auth")
	api.Post("/register", authHandler.Register)
	api.Post("/login", authHandler.Login)
	api.Post("/refresh", authHandler.Refresh)
	api.Post("/verify-email", authHandler.VerifyEmail)
	api.Post("/resend-verification", authHandler.ResendVerification)
	api.Post("/forgot-password", authHandler.ForgotPassword)
	api.Post("/reset-password", authHandler.ResetPassword)

	protected := api.Group("", middleware.JWTAuth(privateKey))
	protected.Post("/logout", authHandler.Logout)
	protected.Get("/me", authHandler.GetMe)
	protected.Put("/me", authHandler.UpdateMe)

	adminHandler.RegisterRoutes(app.Group("", middleware.JWTAuth(privateKey)))

	go func() {
		addr := ":" + cfg.Port
		log.Printf("auth-service listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down auth-service...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = app.ShutdownWithContext(shutdownCtx)
}
