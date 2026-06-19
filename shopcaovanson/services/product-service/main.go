package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caovanson/shopcaovanson/product-service/internal/config"
	"github.com/caovanson/shopcaovanson/product-service/internal/handler"
	"github.com/caovanson/shopcaovanson/product-service/internal/kafka"
	"github.com/caovanson/shopcaovanson/product-service/internal/middleware"
	"github.com/caovanson/shopcaovanson/product-service/internal/repository"
	"github.com/caovanson/shopcaovanson/product-service/internal/service"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()
	mongoDB := mongoClient.Database(cfg.MongoDB)

	esCfg := elasticsearch.Config{Addresses: []string{cfg.ElasticsearchURL}}
	esClient, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		log.Printf("elasticsearch client warning: %v", err)
	}

	producer := kafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	detailRepo := repository.NewProductDetailRepository(mongoDB)
	searchRepo := repository.NewSearchRepository(esClient, cfg.ElasticsearchIndex)

	categorySvc := service.NewCategoryService(categoryRepo)
	productSvc := service.NewProductService(productRepo, categoryRepo, detailRepo, searchRepo, producer)
	reviewSvc := service.NewReviewService(reviewRepo, productRepo)
	productHandler := handler.NewProductHandler(productSvc, categorySvc, reviewSvc)
	categoryHandler := handler.NewCategoryHandler(categorySvc)
	articleRepo := repository.NewArticleRepository(db)
	articleSvc := service.NewArticleService(articleRepo)
	articleHandler := handler.NewArticleHandler(articleSvc)

	app := fiber.New(fiber.Config{
		AppName: "product-service",
	})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.GatewayAuth())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "product-service"})
	})

	productHandler.RegisterRoutes(app)
	categoryHandler.RegisterAdminRoutes(app)
	articleHandler.RegisterRoutes(app)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Printf("product-service listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = app.ShutdownWithContext(ctx)
}
