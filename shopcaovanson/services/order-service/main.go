package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caovanson/shopcaovanson/order-service/internal/client"
	"github.com/caovanson/shopcaovanson/order-service/internal/config"
	"github.com/caovanson/shopcaovanson/order-service/internal/handler"
	"github.com/caovanson/shopcaovanson/order-service/internal/kafka"
	"github.com/caovanson/shopcaovanson/order-service/internal/middleware"
	"github.com/caovanson/shopcaovanson/order-service/internal/repository"
	"github.com/caovanson/shopcaovanson/order-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := sqlx.Connect("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis: %v", err)
	}

	producer := kafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	cartRepo := repository.NewCartRepository(redisClient)
	couponRepo := repository.NewCouponRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	productClient := client.NewProductClient(cfg.ProductServiceURL)
	cartTTL := time.Duration(cfg.CartTTLHours) * time.Hour

	cartSvc := service.NewCartService(cartRepo, productClient, cartTTL)
	couponSvc := service.NewCouponService(couponRepo)
	orderSvc := service.NewOrderService(orderRepo, cartRepo, cartSvc, productClient, couponSvc, producer)
	orderHandler := handler.NewOrderHandler(cartSvc, orderSvc)
	couponHandler := handler.NewCouponHandler(couponSvc)

	app := fiber.New(fiber.Config{AppName: "order-service"})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.GatewayAuth())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "order-service"})
	})

	orderHandler.RegisterRoutes(app)
	couponHandler.RegisterRoutes(app)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Printf("order-service listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = app.ShutdownWithContext(shutdownCtx)
}
